package utils

import "testing"


type EnvParseTestCase struct {
	desc string
	input string
	strs []string
	vars []int
}

func min(a int, b int) int {
	if a <= b {
		return a
	}
	return b
}

func examineStrs(t *testing.T, actual, expected []string) {
	numElems := min(len(actual), len(expected))

	// iter over strings, check for differences
	for i := 0; i < numElems; i++ {
		if actual[i] != expected[i] {
			t.Errorf("strs[%d]: expected string '%s', got '%s'",
				i, expected[i], actual[i])
		}
	}

	// if actual and expected number of strings differ, log this
	if len(actual) != len(expected) {
		if len(actual) < len(expected) {
			t.Errorf("Expected %d strings, got %d, missing %v",
				len(expected), len(actual), expected[numElems:])
		} else {
			t.Errorf("Expected %d strings, got %d, extra: %v",
				len(expected), len(actual), actual[numElems:])
		}
	}
}

func examineVars(t *testing.T, actual, expected []int) {
	numElems := min(len(actual), len(expected))
	for i := 0; i < numElems; i++ {
		if actual[i] != expected[i] {
			t.Errorf("vars[%d]: expected var index %d, got %d",
				i, expected[i], actual[i])
		}
	}
	if len(actual) != len(expected) {
		if len(actual) < len(expected) {
			t.Errorf("Expected %d vars, got %d, missing: %v",
				len(expected), len(actual), expected[numElems:])
		} else {
			t.Errorf("Expected %d vars, got %d, extra: %v",
				len(expected), len(actual), actual[numElems:1])
		}
	}
}

func testEnvParse(t *testing.T, test *EnvParseTestCase) error {
	t.Logf("\n================\nTest: %s\n", test.desc)
	res, err := ParseCmd(test.input)
	if err != nil {
		return err
	}
	strs, vars := res.Strings, res.VarIndices
	examineStrs(t, strs, test.strs)
	examineVars(t, vars, test.vars)
	return nil
}

func TestEnvParse_NoEnvs(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "no vars (#1)",
		input: `echo "hello world"`,
		strs: []string{`echo "hello world"`},
		vars: []int{},
	})
}

func TestEnvParse_TrailingPct(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "empty %() literal @ string end",
		input: `something%`,
		strs: []string{`something%`},
		vars: []int{},
	})
}

func TestEnvParse_TrailingEscapedPct(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "empty %() literal @ string end",
		input: `something%%`,
		strs: []string{`something%`},
		vars: []int{},
	})
}

func TestEnvParse_TrailingDoublePct(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "empty %() literal @ string end",
		input: `something%%%`,
		strs: []string{`something%%`},
		vars: []int{},
	})
}

func TestEnvParse_TrailingDoubleEscapedPct(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "empty %() literal @ string end",
		input: `something%%%%`,
		strs: []string{`something%%`},
		vars: []int{},
	})
}

func TestEnvParse_EmptyVarLiteral(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'%()' interpreted as literal",
		input: `something%()`,
		strs: []string{`something%()`},
		vars: []int{},
	})
}

func TestEnvParse_EmptyVarLiteral2(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'%()' interpreted as literal",
		input: `something%()else`,
		strs: []string{`something%()else`},
		vars: []int{},
	})
}

func TestEnvParse_TrailingLiteral1(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'%(' interpreted as literal",
		input: `something%(`,
		strs: []string{`something%(`},
		vars: []int{},
	})
}

func TestEnvParse_TrailingLiteral2(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'%(other' interpreted as literal",
		input: `something%(other`,
		strs: []string{`something%(other`},
		vars: []int{},
	})
}

func TestEnvParse_Var(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'%(other)' interpreted as a variable",
		input: `something%(other)`,
		strs: []string{`something`, `other`},
		vars: []int{1},
	})
}

func TestEnvParse_Var2(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'%(other)' interpreted as a variable",
		input: `something%(other)else`,
		strs: []string{`something`, `other`, `else`},
		vars: []int{1},
	})
}

func TestEnvParse_Var3_TwoVars(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'test with two vars",
		input: `something%(foo)else%(bar)baz`,
		strs: []string{`something`, `foo`, `else`, `bar`, `baz`},
		vars: []int{1, 3},
	})
}

func TestEnvParse_Var4_LeadingVar(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'test with two vars",
		input: `%(foo)else%(bar)baz`,
		strs: []string{`foo`, `else`, `bar`, `baz`},
		vars: []int{0, 2},
	})
}

func TestEnvParse_EscapedVar(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'%%(other)' will be escaped as literal",
		input: `something%%(other)else`,
		strs: []string{`something%(other)else`},
		vars: []int{},
	})
}

func TestEnvParse_BadVarname(t *testing.T) {
	testCase := &EnvParseTestCase{
		desc: "'%%(other)' will be escaped as literal",
		input: `something%(ot%(her)else`,
		strs: []string{}, // skipped
		vars: []int{}, // skipped
	}
	err := testEnvParse(t, testCase)
	if err == nil {
		t.Error("Expected error for using invalid characters in variable name")
	}
	ie, ok := err.(*InvalidVarNameError)
	if !ok {
		t.Errorf("Expected InvalidVarNameError, got %v", err)
	}
	if ie.VarName != "ot%(her" {
		t.Errorf("Expected var name '%s', got '%s'", "%(ot%(her)", ie.VarName)
	}
	if ie.Command != testCase.input {
		t.Errorf("Error reported wrong command, expected '%s', got '%s'", testCase.input, ie.Command)
	}
}

