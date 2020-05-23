package suggestions

import (
	"fmt"
	"github.com/jwdevantier/spellbook/ui/table"
	"github.com/jwdevantier/spellbook/utils"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"hash/fnv"
	"sort"
)

func hash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

// CommandRow
// Concrete implementation of Row mapping a utils.Command
////////////////////////////////////////////////
type CommandRow struct {
	id uint64
	command utils.Command
}

func (cr *CommandRow) Id() uint64 {
	return cr.id
}

func (cr *CommandRow) Len() int {
	return 2 // Cmd: 0, Desc: 1
}

func (cr *CommandRow) Command() *utils.Command {
	return &cr.command
}

func (cr *CommandRow) CellValue(col int) interface{} {
	switch col {
	case 0:
		return cr.command.Cmd
	case 1:
		return cr.command.Desc
	default:
		panic(fmt.Sprintf("out of range! [0-%d[, got: %d", cr.Len(), col))
	}
	return cr.command
}

func NewCommandRow(command utils.Command) table.Row {
	return &CommandRow{
		id:      hash(command.Cmd),
		command: command,
	}
}

func ToRowsCommands(commands []utils.Command) []table.Row {
	out := make([]table.Row, len(commands))
	for i, command := range commands {
		out[i] = NewCommandRow(command)
	}
	return out
}

// Command Renderers
//
type CommandRenderer struct {}
func (cr *CommandRenderer) Render(row table.Row) []string {
	crow, ok := row.(*CommandRow)
	if !ok {
		panic("Invalid renderer")
	}
	return []string{
		crow.command.Cmd + "    ", // Poor man's padding
		crow.command.Desc,
	}
}

func NewCommandRenderer() *CommandRenderer {
	return &CommandRenderer{}
}

// Command Filter
//
type CommandFuzzyFilter struct {
	filterString string
}

type match struct {
	rank int
	row *CommandRow
}

type matches []match

func (m matches) Len() int {
	return len(m)
}

func (m matches) Less(i, j int) bool {
	return m[i].rank < m[j].rank
}

func (m matches) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (cf *CommandFuzzyFilter) Filter(rows []table.Row) []table.Row {
	if cf.filterString == "" {
		return rows
	}

	// TODO: use ranking from fuzzy package to sort list accordingly.
	m := make(matches, len(rows))
	for i, row := range rows {
		crow := row.(*CommandRow)
		m[i] = match{
			rank: fuzzy.RankMatch(cf.filterString, crow.command.Cmd),
			row: crow}
	}
	sort.Sort(sort.Reverse(m))

	out := make([]table.Row, 0, len(rows))
	for _, v := range m {
		if v.rank == -1 {
			continue
		}
		out = append(out, v.row)
		//cr := v.(*CommandRow)
		//rank := fuzzy.RankMatch(cf.filterString, cr.command.Cmd)
		////fuzzy.RankMatch()
		//if fuzzy.Match(cf.filterString, cr.command.Cmd) {
		//	out = append(out, cr)
		//}
	}
	return out
}

func (cf *CommandFuzzyFilter) SetSearchString(s string) {
	cf.filterString = s
}

func NewCommandFuzzyFilter() *CommandFuzzyFilter {
	return &CommandFuzzyFilter{}
}