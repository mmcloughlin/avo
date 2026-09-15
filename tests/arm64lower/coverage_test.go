package arm64lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// The tests in this file are coverage gates rather than correctness tests: they
// compare what the lowering printer is willing to translate against what this
// suite actually executes on both architectures.
//
// The distinction matters because the printer's failure mode is silence. An
// unsupported opcode panics during generation, which is loud and safe; a
// supported-but-wrong lowering produces assembly that looks plausible and
// computes the wrong thing on one architecture only. Review alone has not been
// enough to find those -- four rounds have now turned up bugs -- so the standing
// requirement is that every lowering the printer offers is executed against an
// amd64 reference, not merely read.
//
// The generated amd64 assembly is the record of what the suite exercises: every
// instruction in it also went through the arm64 lowering, since both files come
// from one avo source.

const printerSource = "../../printer/arm64.go"

// exemptOpcodes lists opcodes the dispatch switch handles that the suite cannot
// execute, with the reason. It is deliberately empty: an entry here is a hole in
// the differential coverage, so adding one should require an argument.
var exemptOpcodes = map[string]string{}

func parsePrinter(t *testing.T) *ast.File {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, printerSource, nil, 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", printerSource, err)
	}
	return f
}

// findFunc returns the named top-level function, method or not.
func findFunc(t *testing.T, f *ast.File, name string) *ast.FuncDecl {
	t.Helper()
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == name {
			return fn
		}
	}
	t.Fatalf("%s: no function %q; the coverage gate needs updating", printerSource, name)
	return nil
}

// clauseStrings returns the string literals labelling one case clause.
func clauseStrings(cc *ast.CaseClause) []string {
	var out []string
	for _, e := range cc.List {
		if lit, ok := e.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			if s, err := strconv.Unquote(lit.Value); err == nil {
				out = append(out, s)
			}
		}
	}
	return out
}

// caseStrings returns the string literals labelling the cases of sw.
func caseStrings(sw *ast.SwitchStmt) []string {
	var out []string
	for _, stmt := range sw.Body.List {
		if cc, ok := stmt.(*ast.CaseClause); ok {
			out = append(out, clauseStrings(cc)...)
		}
	}
	return out
}

// namedOpcodes returns every opcode a function names in a comparison against
// some .Opcode field, in either direction and with == or !=.
//
// Not everything the printer translates reaches lower's switch. BTL is
// recognized by a pre-pass and emitted from function's main loop, so scanning
// the switch alone reported full coverage for it while nothing executed it.
// Collecting opcode literals wherever they are compared catches that shape, and
// the next one, without the gate having to know how each is dispatched.
func namedOpcodes(t *testing.T, f *ast.File, fnName string) []string {
	t.Helper()
	var out []string
	ast.Inspect(findFunc(t, f, fnName), func(n ast.Node) bool {
		be, ok := n.(*ast.BinaryExpr)
		if !ok || (be.Op != token.EQL && be.Op != token.NEQ) {
			return true
		}
		for _, sides := range [][2]ast.Expr{{be.X, be.Y}, {be.Y, be.X}} {
			sel, ok := sides[0].(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Opcode" {
				continue
			}
			if lit, ok := sides[1].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				if s, err := strconv.Unquote(lit.Value); err == nil {
					out = append(out, s)
				}
			}
		}
		return true
	})
	return out
}

// lowerDispatchFuncs lists the functions lower calls to dispatch an opcode,
// one per opcode-table section (see the "----" comments in arm64.go). Keep
// this in sync with lower's if-chain: dispatchedOpcodes reads the top-level
// switch on i.Opcode in each of these rather than in lower itself, since
// lower no longer contains one switch -- it delegates to these.
var lowerDispatchFuncs = []string{
	"lowerMoveOrLoad", "lowerArithOrLogic", "lowerIncDecShiftOp", "lowerShiftXOp",
	"lowerMiscOp", "lowerBitFieldOp", "lowerMultiply", "lowerCompareOp",
}

// dispatchedOpcodes returns the opcodes the printer's instruction dispatch
// handles: the cases of the top-level switch on i.Opcode in each of
// lowerDispatchFuncs, plus any named outside them (see namedOpcodes). Nested
// switches inside a case are ignored -- they refine a lowering rather than
// add an entry point.
func dispatchedOpcodes(t *testing.T) []string {
	t.Helper()
	f := parsePrinter(t)
	var out []string
	for _, fnName := range []string{"btPairs", "function"} {
		out = append(out, namedOpcodes(t, f, fnName)...)
	}
	for _, fnName := range lowerDispatchFuncs {
		fn := findFunc(t, f, fnName)
		found := false
		for _, stmt := range fn.Body.List {
			sw, ok := stmt.(*ast.SwitchStmt)
			if !ok {
				continue
			}
			sel, ok := sw.Tag.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Opcode" {
				continue
			}
			out = append(out, caseStrings(sw)...)
			found = true
			break
		}
		if !found {
			t.Fatalf("%s: no switch on i.Opcode in %s; the coverage gate needs updating", printerSource, fnName)
		}
	}
	return out
}

// mnemonics returns the distinct instruction mnemonics in the generated amd64
// assembly, which is exactly the set of x86 opcodes this suite executes.
func mnemonics(t *testing.T) map[string]bool {
	t.Helper()
	src, err := os.ReadFile("lower_amd64.s")
	if err != nil {
		t.Fatalf("reading generated assembly: %v", err)
	}
	re := regexp.MustCompile(`^\t([A-Z][A-Z0-9]*)\b`)
	out := map[string]bool{}
	for _, line := range strings.Split(string(src), "\n") {
		if m := re.FindStringSubmatch(line); m != nil {
			out[m[1]] = true
		}
	}
	return out
}

// TestOpcodeDispatchIsCovered fails when the printer can lower an opcode that
// this suite never runs. Such a lowering has been reviewed but never executed,
// and every silent miscompile found so far has been in code that looked right.
func TestOpcodeDispatchIsCovered(t *testing.T) {
	exercised := mnemonics(t)
	dispatched := dispatchedOpcodes(t)
	// A gate that silently reads nothing passes for the wrong reason, so hold
	// the parse to a floor well below the real count.
	if len(dispatched) < 50 || len(exercised) < 50 {
		t.Fatalf("read %d dispatched opcodes and %d exercised mnemonics; the gate is not reading what it thinks it is",
			len(dispatched), len(exercised))
	}
	var missing []string
	for _, op := range dispatched {
		if exercised[op] || exemptOpcodes[op] != "" {
			continue
		}
		missing = append(missing, op)
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("the printer lowers these opcodes but no generated function uses them: %s\n"+
			"add a case to asm.go with a pure-Go reference, or add an entry to exemptOpcodes explaining why it cannot be executed",
			strings.Join(missing, " "))
	}
}

// condAliases maps every x86 condition suffix armCond accepts to the arm64
// condition it produces.
func condAliases(t *testing.T) map[string]string {
	t.Helper()
	fn := findFunc(t, parsePrinter(t), "armCond")
	out := map[string]string{}
	for _, stmt := range fn.Body.List {
		sw, ok := stmt.(*ast.SwitchStmt)
		if !ok {
			continue
		}
		for _, s := range sw.Body.List {
			cc, ok := s.(*ast.CaseClause)
			if !ok || len(cc.Body) == 0 {
				continue
			}
			ret, ok := cc.Body[0].(*ast.ReturnStmt)
			if !ok || len(ret.Results) == 0 {
				continue
			}
			lit, ok := ret.Results[0].(*ast.BasicLit)
			if !ok {
				continue
			}
			arm, err := strconv.Unquote(lit.Value)
			if err != nil || arm == "" {
				continue
			}
			for _, alias := range clauseStrings(cc) {
				out[alias] = arm
			}
		}
	}
	if len(out) == 0 {
		t.Fatalf("%s: could not read armCond's mapping; the coverage gate needs updating", printerSource)
	}
	return out
}

// TestConditionTableIsCovered fails when an arm64 condition the translation
// table can produce is never executed in one of the three consumer families.
// The families do not share a code path -- a branch becomes B.cond, a CMOV
// becomes CSEL, a SETcc becomes CSET -- so covering a condition in one says
// nothing about the others.
func TestConditionTableIsCovered(t *testing.T) {
	aliases := condAliases(t)
	want := map[string]bool{}
	for _, arm := range aliases {
		want[arm] = true
	}

	families := []struct {
		name string
		// suffix extracts the condition suffix from a mnemonic, reporting false
		// when the mnemonic is not a member of the family.
		suffix func(string) (string, bool)
	}{
		{"SETcc", func(m string) (string, bool) { return strings.CutPrefix(m, "SET") }},
		{"Jcc", func(m string) (string, bool) {
			if m == "JMP" {
				return "", false
			}
			return strings.CutPrefix(m, "J")
		}},
		{"CMOVcc", func(m string) (string, bool) {
			s, ok := strings.CutPrefix(m, "CMOV")
			if !ok {
				return "", false
			}
			// Strip exactly one operand-size letter: CMOVQCS, CMOVLNE, CMOVWEQ.
			// Trimming a set would eat the L of CMOVQLS and leave "S", which is
			// itself a valid condition alias meaning something else entirely.
			if len(s) > 1 && strings.ContainsRune("QLW", rune(s[0])) {
				s = s[1:]
			}
			return s, true
		}},
	}

	exercised := mnemonics(t)
	for _, f := range families {
		covered := map[string]bool{}
		for m := range exercised {
			if cc, ok := f.suffix(m); ok {
				if arm, known := aliases[cc]; known {
					covered[arm] = true
				}
			}
		}
		var missing []string
		for arm := range want {
			if !covered[arm] {
				missing = append(missing, arm)
			}
		}
		sort.Strings(missing)
		if len(missing) > 0 {
			t.Errorf("%s: arm64 conditions never executed: %s\n"+
				"condspec.Conds drives SetAll/JmpAll/CmovAll; every condition armCond can return needs to appear in all three",
				f.name, strings.Join(missing, " "))
		}
	}
}
