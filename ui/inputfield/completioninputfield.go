package inputfield

import (
	"github.com/gdamore/tcell"
	"github.com/jwdevantier/spellbook/utils"
	"github.com/rivo/tview"
	"reflect"
)

type CompletionInputField struct {
	*tview.InputField

	toks            []utils.Token
	posCompletes	[]int

	previousText    string
}

func NewCompletionInputField() *CompletionInputField {
	return &CompletionInputField{
		InputField: tview.NewInputField()}
}

func (ci *CompletionInputField) tokNdx() int {
	if len(ci.posCompletes) < 1 {
		panic("posCompletes must always have 1 entry")
	}
	return len(ci.posCompletes) - 1
}

func (ci *CompletionInputField) cursorPos() int {
	pos := int(reflect.ValueOf(ci).Elem().FieldByName("cursorPos").Int())
	// clamping logic fetched from Draw() routine
	if pos < 0 {
		return 0
	} else if pos > len(ci.GetText()) {
		return len(ci.GetText())
	}
	return pos
}

func (ci *CompletionInputField) posLastCompletion() int {
	if !ci.CompletionMode() {
		panic("method called outside completion mode")
	}
	return ci.posCompletes[len(ci.posCompletes)-1]
}

func (ci *CompletionInputField) atCompletionPos() bool {
	return ci.cursorPos() == ci.posLastCompletion()
}

func (ci *CompletionInputField) CompletionMode() bool {
	return ci.toks != nil
}

func (ci *CompletionInputField) EnterCompletionMode(cmd string) error {
	if ci.CompletionMode() {
		ci.exitCompletionMode()
	}
	toks, err := utils.ParseCmd(cmd)
	if err != nil {
		return err
	}
	ci.toks = toks
	ci.previousText = ci.GetText()

	ci.SetText("")
	ci.posCompletes = []int{0}
	ci.complete()
	return nil
}

func (ci *CompletionInputField) exitCompletionMode() {
	ci.SetText(ci.previousText)
	ci.toks = nil
	ci.posCompletes = nil

	ci.SetText(ci.previousText)
	ci.previousText = ""
}

func (ci *CompletionInputField) cursorAtLineEnd() bool {
	return ci.cursorPos() == len(ci.GetText())
}

func (ci *CompletionInputField) deleteLastCompletion() {
	if !ci.CompletionMode() {
		panic("method called outside completion mode")
	}
	// remove last completion and everything following it.
	switch len(ci.posCompletes) {
	case 0:
		panic("posCompletes should NEVER be empty, starts as []int{0}")
	case 1:
		panic("should not be called when having no completions to delete")
	default:
		ci.posCompletes = ci.posCompletes[:len(ci.posCompletes)-1]
		endPos := ci.posCompletes[len(ci.posCompletes)-1]
		ci.SetText(ci.GetText()[0:endPos])
	}
	// TODO - if last completion => empty input => exit completion mode
	//		VERIFY THIS
	if ci.GetText() == "" {
		ci.exitCompletionMode()
	}
}

func (ci *CompletionInputField) complete() {
	if !ci.CompletionMode() {
		panic("method called outside completion mode")
	}
	if !ci.cursorAtLineEnd() {
		return
	}

	Loop:
	for i := ci.tokNdx(); i < len(ci.toks); i++ {
		tok := ci.toks[i]
		switch tok.Type {
		case utils.TokVar:
			if ci.cursorPos() > ci.posLastCompletion() {
				// TODO: won't I need to treat this also as a completion?
				ci.posCompletes = append(ci.posCompletes, ci.cursorPos())
			} else {
				break Loop
			}
		case utils.TokLiteral:
			ci.SetText(ci.GetText() + tok.Lexeme)
			ci.posCompletes = append(ci.posCompletes, ci.cursorPos())
		}
	}
}

func (ci *CompletionInputField) Draw(screen tcell.Screen) {
	ci.InputField.Draw(screen)

	if ! ci.CompletionMode() {
		return
	}

	// Only draw these additional bits if in auto-completion mode
	x, y, width, _ := ci.InputField.GetInnerRect()

	fieldWidth := ci.InputField.GetFieldWidth()
	if fieldWidth == 0 { // extend as much as possible
		fieldWidth = width
	}
	fieldStyle := tcell.StyleDefault.Background(tcell.ColorGreen)

	// start drawing AFTER given input
	offset := len(ci.GetLabel()) + len(ci.GetText())

	for ndx := offset; ndx < fieldWidth; ndx++ {
		screen.SetContent(x+ndx, y, ' ', nil, fieldStyle)
	}

	previewTok := ci.nextLiteralTok()
	if previewTok != nil {
		tview.Print(
			screen, previewTok.Lexeme,
			offset + x, y, fieldWidth - offset,
			tview.AlignLeft, tcell.ColorRed)
	}
	// TODO: render variable parts in different color
}

func (ci *CompletionInputField) nextLiteralTok() *utils.Token {
	for i := ci.tokNdx(); i < len(ci.toks); i++ {
		tok := ci.toks[i]
		if tok.Type == utils.TokLiteral {
			return &tok
		}
	}
	return nil
}

func (ci *CompletionInputField) SetInputCapture(handler func(event *tcell.EventKey) *tcell.EventKey) {
	ci.InputField.SetInputCapture(func (event *tcell.EventKey) *tcell.EventKey {
		out := ci.defaultInputCapture(event)
		if out == nil {
			return nil
		}
		return handler(out)
	})
}

func (ci *CompletionInputField) onBackspace(event *tcell.EventKey) *tcell.EventKey {
	cursorPos := ci.cursorPos()
	if cursorPos == 0 {
		// nothing to delete
	} else if cursorPos != ci.posLastCompletion() {
		// deleting some regular character, allow this by bubbling up the event
		return event
	} else if ci.cursorAtLineEnd() {
		// we are at the point of deleting part of the prior block
		// and there is no input following the cursor.

		lastTok := ci.toks[ci.tokNdx()-1]
		// the token we are deleting into is a user-provided variable value
		// => delete char-by-char
		if lastTok.Type == utils.TokVar {
			// pop the completion value (TODO: I think this is ALWAYS going to be necessary)
			if cursorPos == ci.posCompletes[ci.tokNdx()] {
				ci.posCompletes = ci.posCompletes[:len(ci.posCompletes)-1]
			}
			return event
		}

		ci.deleteLastCompletion()
	}
	return nil
}

func (ci *CompletionInputField) defaultInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		if ci.CompletionMode() {
			ci.complete()
			return nil
		}
	case tcell.KeyEscape:
		if ci.CompletionMode() {
			ci.exitCompletionMode()
			return nil
		}
	case tcell.KeyLeft:
		if ci.CompletionMode() && ci.cursorPos() == ci.posLastCompletion() {
			return nil
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if ci.CompletionMode() {
			return ci.onBackspace(event)
		}
	case tcell.KeyHome, tcell.KeyCtrlA, tcell.KeyCtrlW, tcell.KeyCtrlU:
		// Disallow any action which moves behind into line
		if ci.CompletionMode() {
			return nil
		}
	}
	return event
}