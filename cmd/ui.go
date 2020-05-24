package cmd

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/jwdevantier/spellbook/ui/inputfield"
	"github.com/jwdevantier/spellbook/ui/suggestions"
	table2 "github.com/jwdevantier/spellbook/ui/table"
	"github.com/jwdevantier/spellbook/utils"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	"os"
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

func NewInputField() *inputfield.CompletionInputField {
	inputField := inputfield.NewCompletionInputField()
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

		inputField.SetChangedFunc(func(text string) {
			if !inputField.CompletionMode() {
				fuzzy.SetSearchString(text)
				table.Render()
			}
		})

		inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyUp:
				if !inputField.CompletionMode() {
					table.SelectionUp()
					return nil
				}
			case tcell.KeyDown:
				if !inputField.CompletionMode() {
					table.SelectionDown()
					return nil
				}
			case tcell.KeyTab:
				row, found := table.GetSelectedRow()
				if found {
					// TODO: handle error here..?
					_ = inputField.EnterCompletionMode(row.(*suggestions.CommandRow).Command().Cmd)
				}
				return nil
			case tcell.KeyEscape:
				app.Stop()
				return nil
			case tcell.KeyEnter:
				if !inputField.CompletionMode() {
					// not in completion mode, enter it
					row, found := table.GetSelectedRow()
					if found {
						// TODO: handle error here..?
						_ = inputField.EnterCompletionMode(row.(*suggestions.CommandRow).Command().Cmd)
					}
					return nil
				} else if inputField.CompletionDone() {
					app.Stop()
					// Required because of some bug in tcell when cleaning up the screen.
					utils.PressEnterKey()

					rawCmd := inputField.GetText()
					cmd, err := utils.ResolveEnvVars(rawCmd)
					if err != nil {
						fmt.Printf("$ %s\n", rawCmd)
						fmt.Println(err)
						return nil
					}
					fmt.Printf("$ %s\n", cmd)
					os.Exit(utils.ExitCode(utils.Run(cmd)))
					return nil
				}
			}
			return event
		})

		rootGrid.AddItem(inputField, 2, 0, 1, 1, 0, 0, true)
		rootGrid.AddItem(table.Primitive(), 1, 0, 1, 1, 0, 0, true)

		if err := app.SetRoot(rootGrid, true).SetFocus(rootGrid).Run(); err != nil {
			panic(err)
		}
	},
}