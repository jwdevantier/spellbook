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
	tokNdx          int
	posCompletes	[]int
	//posLastComplete int
	previousText    string
}

func NewCompletionInputField() *CompletionInputField {
	return &CompletionInputField{
		InputField: tview.NewInputField(),
		tokNdx:     -1}
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
	if ! ci.CompletionMode() || len(ci.posCompletes) == 0 {
		return -1
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
	ci.tokNdx = 0
	toks, err := utils.ParseCmd(cmd)
	if err != nil {
		return err
	}
	ci.toks = toks
	ci.previousText = ci.GetText()

	ci.SetText("")
	ci.posCompletes = make([]int, 0)
	ci.complete()
	return nil
}

func (ci *CompletionInputField) exitCompletionMode() {
	ci.SetText(ci.previousText)
	ci.toks = nil
	ci.tokNdx = -1
	ci.posCompletes = nil

	ci.SetText(ci.previousText)
	ci.previousText = ""
}

func (ci *CompletionInputField) completionEnd() bool {
	return ci.tokNdx == -1
}

func (ci *CompletionInputField) cursorAtEnd() bool {
	return ci.cursorPos() == len(ci.GetText())
}

func (ci *CompletionInputField) complete() {
	if !ci.CompletionMode() || ci.completionEnd() || !ci.cursorAtEnd() {
		return
	}

	for i := ci.tokNdx; i < len(ci.toks); i++ {
		tok := ci.toks[i]
		if tok.Type != utils.TokVar {
			ci.SetText(ci.GetText() +  tok.Lexeme)
			ci.posCompletes = append(ci.posCompletes, ci.cursorPos())
			continue
		}
		ci.tokNdx = i // record that we're at a variable position
		if ci.cursorPos() <= ci.posLastCompletion() {
			// if user hasn't provided any input for the var, refuse expansion
			break
		}
	}
	if ci.tokNdx == len(ci.toks) {
		// TODO: NEVER hit (for cmds ending w var token, at least)
		ci.tokNdx = -1 // nothing more to auto-complete
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
	// print : return (bytes-written, drawnWidth: int)
	tview.Print(screen, "hello, world", offset + x, y, fieldWidth - offset, tview.AlignLeft, tcell.ColorRed)

	// TODO: Render auto-complete text.

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
	case tcell.KeyLeft, tcell.KeyBackspace, tcell.KeyBackspace2:
		if ci.CompletionMode() && ci.cursorPos() == ci.posLastCompletion() {
			return nil
		}
	case tcell.KeyHome, tcell.KeyCtrlA, tcell.KeyCtrlW, tcell.KeyCtrlU:
		// Disallow any action which moves behind into line
		if ci.CompletionMode() {
			return nil
		}
	}
	return event
}