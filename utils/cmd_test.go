package utils

import "testing"


type EnvParseTestCase struct {
	desc string
	input string
	toks []Token
}

func min(a int, b int) int {
	if a <= b {
		return a
	}
	return b
}

func examineToks(t *testing.T, actual, expected []Token) {
	numElems := min(len(actual), len(expected))

	// iter over strings, check for differences
	for i := 0; i < numElems; i++ {
		if actual[i] != expected[i] {
			t.Errorf("strs[%d]: expected string '%v', got '%v'",
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


func testEnvParse(t *testing.T, test *EnvParseTestCase) error {
	t.Logf("\n================\nTest: %s\n", test.desc)
	toks, err := ParseCmd(test.input)
	if err != nil {
		return err
	}

	examineToks(t, toks, test.toks)
	return nil
}

func TestEnvParse_NoEnvs(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "no vars (#1)",
		input: `echo "hello world"`,
		toks: []Token{
			{TokLiteral, `echo "hello world"`},
		},
	})
}

func TestEnvParse_TrailingPct(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "empty %() literal @ string end",
		input: `something%`,
		toks: []Token{
			{TokLiteral, `something%`},
		},
	})
}

func TestEnvParse_TrailingEscapedPct(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "empty %() literal @ string end",
		input: `something%%`,
		toks: []Token{
			{TokLiteral, `something%`},
		},
	})
}

func TestEnvParse_TrailingDoublePct(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "empty %() literal @ string end",
		input: `something%%%`,
		toks: []Token{
			{TokLiteral,`something%%`},
		},
	})
}

func TestEnvParse_TrailingDoubleEscapedPct(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "empty %() literal @ string end",
		input: `something%%%%`,
		toks: []Token{
			{TokLiteral, `something%%`},
		},
	})
}

func TestEnvParse_EmptyVarLiteral(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'%()' interpreted as literal",
		input: `something%()`,
		toks: []Token{
			{TokLiteral, `something%()`},
		},
	})
}

func TestEnvParse_EmptyVarLiteral2(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'%()' interpreted as literal",
		input: `something%()else`,
		toks: []Token{
			{TokLiteral, `something%()else`},
		},
	})
}

func TestEnvParse_TrailingLiteral1(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'%(' interpreted as literal",
		input: `something%(`,
		toks: []Token{
			{TokLiteral, `something%(`},
		},
	})
}

func TestEnvParse_TrailingLiteral2(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'%(other' interpreted as literal",
		input: `something%(other`,
		toks: []Token{
			{TokLiteral, `something%(other`},
		},
	})
}

func TestEnvParse_Var(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'%(other)' interpreted as a variable",
		input: `something%(other)`,
		toks: []Token{
			{TokLiteral, `something`},
			{TokVar, `other`},
		},
	})
}

func TestEnvParse_Var2(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'%(other)' interpreted as a variable",
		input: `something%(other)else`,
		toks: []Token{
			{TokLiteral, `something`},
			{TokVar, `other`},
			{TokLiteral, `else`},
		},
	})
}

func TestEnvParse_Var3_TwoVars(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'test with two vars",
		input: `something%(foo)else%(bar)baz`,
		toks: []Token{
			{TokLiteral, `something`},
			{TokVar, `foo`},
			{TokLiteral, `else`},
			{TokVar, `bar`},
			{TokLiteral, `baz`},
		},
	})
}

func TestEnvParse_Var4_LeadingVar(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'test with two vars",
		input: `%(foo)else%(bar)baz`,
		toks: []Token{
			{TokVar, `foo`},
			{TokLiteral, `else`},
			{TokVar, `bar`},
			{TokLiteral, `baz`},
		},
	})
}

func TestEnvParse_EscapedVar(t *testing.T) {
	testEnvParse(t, &EnvParseTestCase{
		desc: "'%%(other)' will be escaped as literal",
		input: `something%%(other)else`,
		toks: []Token{
			{TokLiteral, `something%(other)else`},
		},
	})
}

func TestEnvParse_BadVarname(t *testing.T) {
	testCase := &EnvParseTestCase{
		desc: "'%%(other)' will be escaped as literal",
		input: `something%(ot%(her)else`,
		toks: []Token{}, // skipped
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

