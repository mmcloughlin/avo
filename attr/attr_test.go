package attr

import "testing"

func TestAttributeAsm(t *testing.T) {
	cases := []struct {
		Attribute Attribute
		Expect    string
	}{
		{0, "0"},
		{32768, "32768"},
		{1, "NOPROF"},
		{DUPOK, "DUPOK"},
		{RODATA | NOSPLIT, "NOSPLIT|RODATA"},
		{WRAPPER | 16384 | NOPTR, "NOPTR|WRAPPER|16384"},
		{NEEDCTXT + NOFRAME + TLSBSS, "NEEDCTXT|TLSBSS|NOFRAME"},
		{REFLECTMETHOD, "REFLECTMETHOD"},
		{TOPFRAME, "TOPFRAME"},
	}
	for _, c := range cases {
		got := c.Attribute.Asm()
		if got != c.Expect {
			t.Errorf("Attribute(%d).Asm() = %#v; expect %#v", c.Attribute, got, c.Expect)
		}
	}
}

func TestAttributeContainsTextFlags(t *testing.T) {
	cases := []struct {
		Attribute Attribute
		Expect    bool
	}{
		{0, false},
		{32768, false},
		{1, true},
		{DUPOK, true},
		{WRAPPER | 16384 | NOPTR, true},
	}
	for _, c := range cases {
		if c.Attribute.ContainsTextFlags() != c.Expect {
			t.Errorf("%s: ContainsTextFlags() expected %#v", c.Attribute.Asm(), c.Expect)
		}
	}
}

func TestAttributeTestMethods(t *testing.T) {
	cases := []struct {
		Attribute Attribute
		Predicate func(Attribute) bool
		Expect    bool
	}{
		// Confirm logic works as expected.
		{DUPOK | NOSPLIT, Attribute.DUPOK, true},
		{DUPOK | NOSPLIT, Attribute.NOSPLIT, true},
		{DUPOK | NOSPLIT, Attribute.NOFRAME, false},

		// Basic test for every method.
		{NOPROF, Attribute.NOPROF, true},
		{DUPOK, Attribute.DUPOK, true},
		{NOSPLIT, Attribute.NOSPLIT, true},
		{RODATA, Attribute.RODATA, true},
		{NOPTR, Attribute.NOPTR, true},
		{WRAPPER, Attribute.WRAPPER, true},
		{NEEDCTXT, Attribute.NEEDCTXT, true},
		{TLSBSS, Attribute.TLSBSS, true},
		{NOFRAME, Attribute.NOFRAME, true},
		{REFLECTMETHOD, Attribute.REFLECTMETHOD, true},
		{TOPFRAME, Attribute.TOPFRAME, true},
	}
	for _, c := range cases {
		if c.Predicate(c.Attribute) != c.Expect {
			t.Errorf("%s: expected %#v", c.Attribute.Asm(), c.Expect)
		}
	}
}
