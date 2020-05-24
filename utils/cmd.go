package utils

import (
	"fmt"
	"github.com/google/shlex"
	"os"
	"os/exec"
	"strings"
)

func Run(cmd string) error {
	lexemes, err := shlex.Split(cmd)
	if err != nil {
		return err
	}

	c := exec.Command(lexemes[0], lexemes[1:]...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// ExitCode returns the program exit code (if any)
// -1: (probably) ignore - most commands set error-codes between 0-255
// 0: the command exited successfully
// 0-255: program exit code
func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	if e, ok := err.(*exec.ExitError); ok {
		return e.ExitCode()
	}

	// Negative exit codes are technically allowed by POSIX
	// but e.g. the wait-family of system calls truncates value to
	// be an unsigned 1B (0-255) val are supported.
	// Hence most well-behaved programs limit exit codes to this range.
	return -1
}

type TokType uint8
const (
	TokLiteral = iota
	TokVar
)

type Token struct {
	Type TokType
	Lexeme string
}

type InvalidVarNameError struct {
	Command string
	VarName string
}

func (ie *InvalidVarNameError) Error() string {
	return fmt.Sprintf("Invalid variable name '%s' in command '%s'. Only a-zA-Z_- allowed", ie.VarName, ie.Command)
}

func NewInvalidVarNameError(cmd string, varName string) *InvalidVarNameError {
	return &InvalidVarNameError{
		Command: cmd, VarName: varName,
	}
}

func ParseCmd(cmd string) ([]Token, error) {
	toks := make([]Token, 0)

	buf := make([]byte, 0)

	emitBuf := func(typ TokType) {
		if len(buf) != 0 {
			toks = append(toks, Token{typ, string(buf)})
			buf = make([]byte, 0)
		}
	}

	start := 0
	pos := start
	for true {
		off := strings.IndexRune(cmd[pos:], '%')
		if off == -1 || pos + off == len(cmd) - 1 {
			// the end of the string is reached. Treat rest as a literal
			buf = append(buf, cmd[start:]...)
			emitBuf(TokLiteral)
			break
		}
		pos += off

		chNext := cmd[pos+1]
		if chNext == '%' {
			// '%%' => escape sequence for %, write one, skip one, loop
			buf = append(buf, cmd[start:pos + 1]...)
			pos += 2
			start = pos
			continue
		} else if chNext != '(' {
			// not '%%' or '%(', treat as literal character
			pos += 2
			continue
		}

		// var parsing
		off = strings.IndexRune(cmd[pos:], ')')
		if off == 2 { // '%()' => treat as literal
			pos += off
			continue
		} else if off == -1 {
			// the end of the string is reached
			buf = append(buf, cmd[start:]...)
			emitBuf(TokLiteral)
			break
		}

		buf = append(buf, cmd[start:pos]...) // up to first '%'
		emitBuf(TokLiteral)

		// skip '%('
		start = pos + 2
		pos += off

		// variable identifier
		ident := cmd[start:pos]
		for _, ch := range ident {
			if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '-' || ch == '_' {
			} else {
				return nil, NewInvalidVarNameError(cmd, ident)
			}
		}
		buf = append(buf, ident...)
		emitBuf(TokVar)
		pos += 1
		start = pos
	}
	return toks, nil
}