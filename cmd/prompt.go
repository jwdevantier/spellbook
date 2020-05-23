package cmd

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	rootCmd.AddCommand(promptCmd)
}

//var promptCmd = &cobra.Command{
//	Use: "promptx",
//	Short: "testing go-prompt",
//	Run: func(cmd *cobra.Command, args []string) {
//		fmt.Println("smth")
//		in := prompt.Input(">>> ", completer,
//			prompt.OptionTitle("-option-title-"),
//			prompt.OptionHistory([]string{"select * from users", "select * from devils"}),
//			prompt.OptionPrefixTextColor(prompt.Yellow),
//			prompt.OptionPreviewSuggestionBGColor(prompt.LightGray),
//			prompt.OptionSuggestionBGColor(prompt.DarkGray),)
//		fmt.Println("your option: " + in)
//	},
//}

//type Track struct {
//	Name string
//	AlbumName string
//	Artist string
//}

//var promptCmd = &cobra.Command{
//	Use: "promptx",
//	Short: "testing go-prompt",
//	Run: func(cmd *cobra.Command, args []string) {
//		fmt.Println("smth")
//		tracks := []Track{
//			{"foo1", "album1", "artist1"},
//			{"foo2", "album2", "artist2"},
//			{"foo3", "album3", "artist3"},
//			{"foo3", "album4", "artist4"},
//		}
//		idx, err := ff.Find(tracks, func(i int) string {
//			return tracks[i].Name
//		},
//		ff.WithPreviewWindow(func(i, w, h int) string {
//			if i == -1 {
//				return ""
//			}
//			return fmt.Sprintf("Track: %s (%s)\nAlbum: %s",
//				tracks[i].Name, tracks[i].Artist, tracks[i].AlbumName)
//		}))
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Printf("selected: %v\n", idx)
//	},
//}


//type Command struct {
//	Command string `json:"command"`
//	Description string `json:"description"`
//	//Tags []string `json:"tags"`
//}
//
//// Create suggestion searching on description
//func (c *Command) promptDescription() prompt.Suggest {
//	return prompt.Suggest{Text: c.Description, Description: c.Command}
//}
//
//// Create suggestion searching on command
//func (c *Command) promptCommand() prompt.Suggest {
//	return prompt.Suggest{Text: c.Command, Description: c.Description}
//}
//
//type Commands struct {
//	Commands []Command `json:"Command"`
//}
//
//func(c *Commands) PromptDescriptions() []prompt.Suggest {
//	res := make([]prompt.Suggest, len(c.Commands))
//	for i, v := range c.Commands {
//		res[i] = v.promptDescription()
//	}
//	return res
//}
//
//func(c *Commands) PromptCommands() []prompt.Suggest {
//	res := make([]prompt.Suggest, len(c.Commands))
//	for i, v := range c.Commands {
//		res[i] = v.promptCommand()
//	}
//	return res
//}
//
//var commands = Commands{
//	Commands: []Command{
//		Command{"git ${commit}~ ${commit}", "show changes in commit"},
//		Command{"git diff ${one}..${other}", "show differences between branches"},
//		Command{"git log --pretty=oneline", "log on one line"},
//		Command{"git push origin --delete ${tag}", "delete remote tag"},
//	},
//}
//

//func completer(in prompt.Document) []prompt.Suggest {
//	s := Config.SuggestByCommands()
//	if in.GetWordBeforeCursor() == "" {
//		return []prompt.Suggest{}
//	}
//
//	for _, word := range strings.Split(in.Text, " ") {
//		s = prompt.FilterContains(s, word, true)
//		if len(s) == 0 {
//			break
//		}
//	}
//	return s
//}
//
//var promptCmd = &cobra.Command{
//	Use: "promptx",
//	Short: "testing go-prompt",
//	Run: func(cmd *cobra.Command, args []string) {
//		fmt.Println("LATER")
//		in := prompt.Input(">>> ", completer,
//			prompt.OptionTitle("-option-title-"),
//			// TODO: consider implementing history of finished commands
//			//prompt.OptionHistory([]string{"select * from users", "select * from devils"}),
//			prompt.OptionPrefixTextColor(prompt.Yellow),
//			prompt.OptionPreviewSuggestionBGColor(prompt.LightGray),
//			prompt.OptionSuggestionBGColor(prompt.DarkGray),)
//		fmt.Println("your option: " + in)
//	},
//}

type PromptState struct {

}

func (p *PromptState) completer(in prompt.Document) []prompt.Suggest {
	s := Config.SuggestByCommands()
	if in.GetWordBeforeCursor() == "" {
		return []prompt.Suggest{}
	}

	for _, word := range strings.Split(in.Text, " ") {
		s = prompt.FilterContains(s, word, true)
		if len(s) == 0 {
			break
		}
	}
	return s
}

// Create stateful completer
// Initial complete
// Foreach arg, complete


func (p *PromptState) Run(cmd *cobra.Command, args []string) {
	fmt.Println("LATER!!")
	in := prompt.Input(">>> ", p.completer,
		prompt.OptionTitle("-option-title-"),
		// TODO: consider implementing history of finished commands
		//prompt.OptionHistory([]string{"select * from users", "select * from devils"}),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),)
	fmt.Println("your option: " + in)
}

var promptCmd = &cobra.Command{
	Use: "promptx",
	Short: "testing go-prompt",
	Run: func(cmd *cobra.Command, args []string) {
		ps := &PromptState{}
		ps.Run(cmd, args)
	},
}