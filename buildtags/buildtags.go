package buildtags

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// Reference: https://github.com/golang/go/blob/204a8f55dc2e0ac8d27a781dab0da609b98560da/src/go/build/doc.go#L73-L92
//
//	// A build constraint is evaluated as the OR of space-separated options;
//	// each option evaluates as the AND of its comma-separated terms;
//	// and each term is an alphanumeric word or, preceded by !, its negation.
//	// That is, the build constraint:
//	//
//	//	// +build linux,386 darwin,!cgo
//	//
//	// corresponds to the boolean formula:
//	//
//	//	(linux AND 386) OR (darwin AND (NOT cgo))
//	//
//	// A file may have multiple build constraints. The overall constraint is the AND
//	// of the individual constraints. That is, the build constraints:
//	//
//	//	// +build linux darwin
//	//	// +build 386
//	//
//	// corresponds to the boolean formula:
//	//
//	//	(linux OR darwin) AND 386
//

type Interface interface {
	ConstraintsConvertable
	fmt.GoStringer
	Evaluate(v map[string]bool) bool
	Validate() error
}

type ConstraintsConvertable interface {
	ToConstraints() Constraints
}

type ConstraintConvertable interface {
	ToConstraint() Constraint
}

type OptionConvertable interface {
	ToOption() Option
}

type (
	Constraints []Constraint
	Constraint  []Option
	Option      []Term
	Term        string
)

func (cs Constraints) ToConstraints() Constraints { return cs }

func (cs Constraints) Validate() error {
	for _, c := range cs {
		if err := c.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (cs Constraints) Evaluate(v map[string]bool) bool {
	r := true
	for _, c := range cs {
		r = r && c.Evaluate(v)
	}
	return r
}

func (cs Constraints) GoString() string {
	s := ""
	for _, c := range cs {
		s += c.GoString()
	}
	return s
}

func (c Constraint) ToConstraints() Constraints { return Constraints{c} }
func (c Constraint) ToConstraint() Constraint   { return c }

func (c Constraint) Validate() error {
	for _, o := range c {
		if err := o.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c Constraint) Evaluate(v map[string]bool) bool {
	r := false
	for _, o := range c {
		r = r || o.Evaluate(v)
	}
	return r
}

func (c Constraint) GoString() string {
	s := "// +build"
	for _, o := range c {
		s += " " + o.GoString()
	}
	return s + "\n"
}

func (o Option) ToConstraints() Constraints { return o.ToConstraint().ToConstraints() }
func (o Option) ToConstraint() Constraint   { return Constraint{o} }
func (o Option) ToOption() Option           { return o }

func (o Option) Validate() error {
	for _, t := range o {
		if err := t.Validate(); err != nil {
			return fmt.Errorf("invalid term \"%s\": %s", t, err)
		}
	}
	return nil
}

func (o Option) Evaluate(v map[string]bool) bool {
	r := true
	for _, t := range o {
		r = r && t.Evaluate(v)
	}
	return r
}

func (o Option) GoString() string {
	var ts []string
	for _, t := range o {
		ts = append(ts, t.GoString())
	}
	return strings.Join(ts, ",")
}

func (t Term) ToConstraints() Constraints { return t.ToOption().ToConstraints() }
func (t Term) ToConstraint() Constraint   { return t.ToOption().ToConstraint() }
func (t Term) ToOption() Option           { return Option{t} }

func (t Term) IsNegated() bool { return strings.HasPrefix(string(t), "!") }

func (t Term) Name() string {
	return strings.TrimPrefix(string(t), "!")
}

func (t Term) Validate() error {
	// Reference: https://github.com/golang/go/blob/204a8f55dc2e0ac8d27a781dab0da609b98560da/src/cmd/go/internal/imports/build.go#L110-L112
	//
	//		if strings.HasPrefix(name, "!!") { // bad syntax, reject always
	//			return false
	//		}
	//
	if strings.HasPrefix(string(t), "!!") {
		return errors.New("at most one '!' allowed")
	}

	if len(t.Name()) == 0 {
		return errors.New("empty tag name")
	}

	// Reference: https://github.com/golang/go/blob/204a8f55dc2e0ac8d27a781dab0da609b98560da/src/cmd/go/internal/imports/build.go#L121-L127
	//
	//		// Tags must be letters, digits, underscores or dots.
	//		// Unlike in Go identifiers, all digits are fine (e.g., "386").
	//		for _, c := range name {
	//			if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' && c != '.' {
	//				return false
	//			}
	//		}
	//
	for _, c := range t.Name() {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' && c != '.' {
			return fmt.Errorf("character '%c' disallowed in tags", c)
		}
	}

	return nil
}

func (t Term) Evaluate(v map[string]bool) bool {
	return (t.Validate() == nil) && (v[t.Name()] == !t.IsNegated())
}

func (t Term) GoString() string { return string(t) }

func Not(ident string) Term {
	return Term("!" + ident)
}

func And(cs ...ConstraintConvertable) Constraints {
	constraints := Constraints{}
	for _, c := range cs {
		constraints = append(constraints, c.ToConstraint())
	}
	return constraints
}

func Any(opts ...OptionConvertable) Constraint {
	c := Constraint{}
	for _, opt := range opts {
		c = append(c, opt.ToOption())
	}
	return c
}

func Opt(terms ...Term) Option {
	return Option(terms)
}

func ParseOption(expr string) (Option, error) {
	opt := Option{}
	for _, t := range strings.Split(expr, ",") {
		opt = append(opt, Term(t))
	}
	return opt, opt.Validate()
}

func ParseConstraint(expr string) (Constraint, error) {
	c := Constraint{}
	for _, field := range strings.Fields(expr) {
		opt, err := ParseOption(field)
		if err != nil {
			return c, err
		}
		c = append(c, opt)
	}
	return c, nil
}

func SetTags(names ...string) map[string]bool {
	v := map[string]bool{}
	for _, n := range names {
		v[n] = true
	}
	return v
}
