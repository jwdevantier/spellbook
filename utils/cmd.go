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

type ParseCmdResult struct {
	Strings []string
	VarIndices []int
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

func ParseCmd(cmd string) (*ParseCmdResult, error) {
	strs := make([]string, 0)
	vars := make([]int, 0)

	buf := make([]byte, 0)

	emitBuf := func() {
		if len(buf) != 0 {
			strs = append(strs, string(buf))
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
			emitBuf()
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
			emitBuf()
			break
		}

		buf = append(buf, cmd[start:pos]...) // up to first '%'
		emitBuf()

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
		vars = append(vars, len(strs))
		emitBuf()
		pos += 1
		start = pos
	}
	return &ParseCmdResult{
		Strings: strs,
		VarIndices: vars,
	}, nil
}