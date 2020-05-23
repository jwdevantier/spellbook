package cmd

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/jwdevantier/spellbook/ui/suggestions"
	table2 "github.com/jwdevantier/spellbook/ui/table"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
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
}

func NewMyInputField() *MyInputField {
	return &MyInputField{tview.NewInputField()}

}

func (mi *MyInputField) Draw(screen tcell.Screen) {
	mi.InputField.Draw(screen)
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
	//text.SetBackgroundColor(backgroundColor)
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
		//table.SetBackgroundColor(backgroundColor)

		rootGrid := tview.NewGrid().
			SetRows(1, -1, 1). // height of each row
			SetColumns(0).
			SetBorders(true)
		//rootGrid.SetBackgroundColor(tcell.ColorDarkSlateGrey)
		//rootGrid.SetBordersColor(borderColor)

		rootGrid.AddItem(NewTextView("Header"), 0, 0, 1, 1, 0, 0, false)

		inputField := NewInputField()
		inputField.SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				row, found := table.GetSelectedRow()
				if !found {
					return
				}
				app.Stop()
				cmd := row.(*suggestions.CommandRow).Command().Cmd
				fmt.Printf("$ %s\n", cmd)
				// TODO: uncomment to run
				//utils.Run(cmd)
			} else if key == tcell.KeyEscape {
				app.Stop()
			}
		})
		inputField.SetChangedFunc(func(text string) {
			fuzzy.SetSearchString(text)
			table.Render()
		})
		inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyUp {
				table.SelectionUp()
				return nil
			} else if event.Key() == tcell.KeyDown {
				table.SelectionDown()
				return nil
			} else if event.Key() == tcell.KeyTab {
				row, found := table.GetSelectedRow()
				if found {
					inputField.SetText(row.(*suggestions.CommandRow).Command().Cmd)
				}
				return nil
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