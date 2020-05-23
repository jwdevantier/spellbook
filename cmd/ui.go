package cmd

import (
	"github.com/gdamore/tcell"
	"github.com/jwdevantier/spellbook/ui/suggestions"
	table2 "github.com/jwdevantier/spellbook/ui/table"
	"github.com/jwdevantier/spellbook/utils"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	"reflect"
)

func initStyle() {
	// go to def on tcell.ColorWhite
	// first 16 colors match theme in e.g. iterm2
	// ContrastBackgroundColor is used for input fields
	// ContrastSecondaryTextColor - placeholder text for input fields
	// PrimitiveBackgroundColor - general bg, not used for grid border cells, though

	tview.Styles.PrimitiveBackgroundColor = 0 // 0:black
}

func init() {
	rootCmd.AddCommand(uiCmd)

	initStyle()
}

// TODO: investigate tview Theme structure
//var backgroundColor tcell.Color = tcell.ColorDarkSlateGrey
//var borderColor tcell.Color = tcell.ColorLightSlateGray

type MyInputField struct {
	*tview.InputField

	toks            []utils.Token
	tokNdx          int
	posCompletes	[]int
	//posLastComplete int
	previousText    string
}

func NewMyInputField() *MyInputField {
	return &MyInputField{
		InputField: tview.NewInputField(),
		tokNdx:     -1}
}

func (mi *MyInputField) cursorPos() int {
	pos := int(reflect.ValueOf(mi).Elem().FieldByName("cursorPos").Int())
	// clamping logic fetched from Draw() routine
	if pos < 0 {
		return 0
	} else if pos > len(mi.GetText()) {
		return len(mi.GetText())
	}
	return pos
}

func (mi *MyInputField) posLastCompletion() int {
	if ! mi.completionMode() || len(mi.posCompletes) == 0 {
		return -1
	}
	return mi.posCompletes[len(mi.posCompletes)-1]
}

func (mi *MyInputField) atCompletionPos() bool {
	return mi.cursorPos() == mi.posLastCompletion()
}

func (mi *MyInputField) completionMode() bool {
	return mi.toks != nil
}

func (mi *MyInputField) enterCompletionMode(cmd string) error {
	if mi.completionMode() {
		mi.exitCompletionMode()
	}
	mi.tokNdx = 0
	toks, err := utils.ParseCmd(cmd)
	if err != nil {
		return err
	}
	mi.toks = toks
	mi.previousText = mi.GetText()

	mi.SetText("")
	mi.posCompletes = make([]int, 0)
	mi.complete()
	return nil
}

func (mi *MyInputField) exitCompletionMode() {
	mi.SetText(mi.previousText)
	mi.toks = nil
	mi.tokNdx = -1
	mi.posCompletes = nil

	mi.SetText(mi.previousText)
	mi.previousText = ""
}

func (mi *MyInputField) completionEnd() bool {
	return mi.tokNdx == -1
}

func (mi *MyInputField) cursorAtEnd() bool {
	return mi.cursorPos() == len(mi.GetText())
}

func (mi *MyInputField) complete() {
	if !mi.completionMode() || mi.completionEnd() || !mi.cursorAtEnd() {
		return
	}

	for i := mi.tokNdx; i < len(mi.toks); i++ {
		tok := mi.toks[i]
		if tok.Type != utils.TokVar {
			mi.SetText(mi.GetText() +  tok.Lexeme)
			mi.posCompletes = append(mi.posCompletes, mi.cursorPos())
			continue
		}
		mi.tokNdx = i // record that we're at a variable position
		if mi.cursorPos() <= mi.posLastCompletion() {
			// if user hasn't provided any input for the var, refuse expansion
			break
		}
	}
	if mi.tokNdx == len(mi.toks) {
		// TODO: NEVER hit (for cmds ending w var token, at least)
		mi.tokNdx = -1 // nothing more to auto-complete
	}
}

func (mi *MyInputField) Draw(screen tcell.Screen) {
	mi.InputField.Draw(screen)

	if ! mi.completionMode() {
		return
	}

	// Only draw these additional bits if in auto-completion mode
	x, y, width, _ := mi.InputField.GetInnerRect()

	fieldWidth := mi.InputField.GetFieldWidth()
	if fieldWidth == 0 { // extend as much as possible
		fieldWidth = width
	}
	fieldStyle := tcell.StyleDefault.Background(tcell.ColorGreen)

	// start drawing AFTER given input
	offset := len(mi.GetLabel()) + len(mi.GetText())

	for ndx := offset; ndx < fieldWidth; ndx++ {
		screen.SetContent(x+ndx, y, ' ', nil, fieldStyle)
	}
	// print : return (bytes-written, drawnWidth: int)
	tview.Print(screen, "hello, world", offset + x, y, fieldWidth - offset, tview.AlignLeft, tcell.ColorRed)

	// TODO: Render auto-complete text.

}

func NewInputField() *MyInputField {
	inputField := NewMyInputField()
	inputField.SetLabel("> ").SetFieldWidth(0)
	inputField.SetPlaceholder("type command")

	return inputField
}

func NewTextView(label string) *tview.TextView {
	text := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(label)
	return text
}


var uiCmd = &cobra.Command{
	Use: "ui",
	Short: "testing go-prompt",
	Run: func(cmd *cobra.Command, args []string) {
		app := tview.NewApplication()

		tableModel := table2.NewTableModel(suggestions.ToRowsCommands(Config.Commands))
		renderer := suggestions.NewCommandRenderer()
		table := table2.NewTable(tableModel, renderer)
		fuzzy := suggestions.NewCommandFuzzyFilter()
		table.SetFilter(fuzzy)
		table.SetOnSelected(func(cell *tview.TableCell) {
			//cell.SetTextColor(tcell.ColorRebeccaPurple)
		})
		// TODO: move up to app-level if possible
		table.SetOnEsc(func() {
			app.Stop()
		})

		rootGrid := tview.NewGrid().
			SetRows(1, -1, 1). // height of each row
			SetColumns(0).
			SetBorders(true)

		rootGrid.AddItem(NewTextView("Header"), 0, 0, 1, 1, 0, 0, false)

		inputField := NewInputField()
		//inputField.SetDoneFunc(func(key tcell.Key) {
		//	if key == tcell.KeyEnter {
		//		row, found := table.GetSelectedRow()
		//		if !found {
		//			return
		//		}
		//		app.Stop()
		//		toks := row.(*suggestions.CommandRow).Command().Cmd
		//		fmt.Printf("$ %s\n", toks)
		//		// TODO: uncomment to run
		//		//utils.Run(toks)
		//	} else if key == tcell.KeyEscape {
		//		app.Stop()
		//	}
		//})
		inputField.SetChangedFunc(func(text string) {
			fuzzy.SetSearchString(text)
			table.Render()
		})
		inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyUp:
				if ! inputField.completionMode() {
					table.SelectionUp()
				}
				return nil
			case tcell.KeyDown:
				if ! inputField.completionMode() {
					table.SelectionDown()
				}
				return nil
			case tcell.KeyTab:
				if inputField.completionMode() {
					inputField.complete()
					return nil
				}
				row, found := table.GetSelectedRow()
				if found {
					inputField.enterCompletionMode(row.(*suggestions.CommandRow).Command().Cmd)
				}
				return nil
			case tcell.KeyEscape:
				if inputField.completionMode() {
					inputField.exitCompletionMode()
				} else {
					app.Stop()
				}
				return nil
			case tcell.KeyLeft, tcell.KeyBackspace, tcell.KeyBackspace2:
				if inputField.completionMode() && inputField.cursorPos() == inputField.posLastCompletion() {
					return nil
				}
			case tcell.KeyHome, tcell.KeyCtrlA, tcell.KeyCtrlW, tcell.KeyCtrlU:
				// Disallow any action which moves behind into line
				if inputField.completionMode() {
					return nil
				}
			}
			return event
		})

		table.SetOnTab(func() {
			app.SetFocus(inputField)
		})
		rootGrid.AddItem(inputField, 2, 0, 1, 1, 0, 0, true)
		rootGrid.AddItem(table.Primitive(), 1, 0, 1, 1, 0, 0, true)

		if err := app.SetRoot(rootGrid, true).SetFocus(rootGrid).Run(); err != nil {
			panic(err)
		}
	},
}