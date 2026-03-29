package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/inventage-ai/asylum/internal/term"
)

// StepKind determines whether a wizard step is a single-choice or multi-choice prompt.
type StepKind int

const (
	StepSelect      StepKind = iota // single-choice (like tui.Select)
	StepMultiSelect                 // multi-choice (like tui.MultiSelect)
)

// WizardStep defines one step in the wizard flow.
type WizardStep struct {
	Title      string
	Kind       StepKind
	Options    []Option
	DefaultIdx int   // default selection for StepSelect
	DefaultSel []int // default selections for StepMultiSelect
}

// StepResult holds the outcome of a single wizard step.
type StepResult struct {
	SelectIdx int   // selected index for StepSelect
	MultiIdx  []int // selected indices for StepMultiSelect
	Completed bool  // false if the user cancelled at or before this step
}

// Wizard runs a multi-step wizard and returns results for each step.
// Steps that were completed before a cancel are marked Completed.
// Returns default selections without prompting if stdin is not a TTY.
func Wizard(steps []WizardStep) ([]StepResult, error) {
	results := make([]StepResult, len(steps))

	if !term.IsTerminal() {
		for i, s := range steps {
			results[i].Completed = true
			if s.Kind == StepSelect {
				results[i].SelectIdx = s.DefaultIdx
			} else {
				results[i].MultiIdx = append([]int(nil), s.DefaultSel...)
			}
		}
		return results, nil
	}

	m := wizardModel{
		steps:   steps,
		results: results,
	}
	m.initStep(0)

	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return nil, err
	}

	final := result.(wizardModel)
	return final.results, nil
}

type wizardModel struct {
	steps   []WizardStep
	results []StepResult
	current int

	// Current step's sub-model state
	selModel   selectModel
	multiModel multiModel
}

func (m *wizardModel) initStep(idx int) {
	s := m.steps[idx]
	if s.Kind == StepSelect {
		m.selModel = selectModel{
			title:   s.Title,
			options: s.Options,
			cursor:  s.DefaultIdx,
		}
	} else {
		selected := map[int]bool{}
		for _, i := range s.DefaultSel {
			selected[i] = true
		}
		m.multiModel = multiModel{
			title:    s.Title,
			options:  s.Options,
			selected: selected,
		}
	}
}

func (m wizardModel) Init() tea.Cmd { return nil }

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c", "q":
			// Cancel: mark current and remaining steps as not completed
			return m, tea.Quit

		case "enter":
			// Collect result from current step
			s := m.steps[m.current]
			m.results[m.current].Completed = true
			if s.Kind == StepSelect {
				m.results[m.current].SelectIdx = m.selModel.cursor
			} else {
				var indices []int
				for i := range m.multiModel.options {
					if m.multiModel.selected[i] {
						indices = append(indices, i)
					}
				}
				m.results[m.current].MultiIdx = indices
			}

			// Advance or finish
			if m.current+1 >= len(m.steps) {
				return m, tea.Quit
			}
			m.current++
			m.initStep(m.current)
			return m, nil

		default:
			// Delegate to current step's model
			s := m.steps[m.current]
			if s.Kind == StepSelect {
				switch msg.String() {
				case "up", "k":
					if m.selModel.cursor > 0 {
						m.selModel.cursor--
					}
				case "down", "j":
					if m.selModel.cursor < len(m.selModel.options)-1 {
						m.selModel.cursor++
					}
				}
			} else {
				switch msg.String() {
				case "up", "k":
					if m.multiModel.cursor > 0 {
						m.multiModel.cursor--
					}
				case "down", "j":
					if m.multiModel.cursor < len(m.multiModel.options)-1 {
						m.multiModel.cursor++
					}
				case " ":
					m.multiModel.selected[m.multiModel.cursor] = !m.multiModel.selected[m.multiModel.cursor]
				}
			}
		}
	}
	return m, nil
}

var (
	tabActiveStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	tabDoneStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	tabFutureStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	tabSepStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func (m wizardModel) View() string {
	var b strings.Builder

	// Tab bar
	for i, s := range m.steps {
		if i > 0 {
			b.WriteString(tabSepStyle.Render("  ▸  "))
		}
		label := fmt.Sprintf("%d. %s", i+1, s.Title)
		if m.results[i].Completed {
			b.WriteString(tabDoneStyle.Render(label + " ✓"))
		} else if i == m.current {
			b.WriteString(tabActiveStyle.Render(label))
		} else {
			b.WriteString(tabFutureStyle.Render(label))
		}
	}
	b.WriteByte('\n')

	// Current step content
	s := m.steps[m.current]
	if s.Kind == StepSelect {
		b.WriteString(m.renderSelect())
	} else {
		b.WriteString(m.renderMulti())
	}

	return "\n" + borderStyle.Render(b.String()) + "\n"
}

func (m wizardModel) renderSelect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(m.selModel.title))
	b.WriteByte('\n')

	for i, opt := range m.selModel.options {
		if i > 0 {
			b.WriteByte('\n')
		}
		cursor := "  "
		label := labelStyle.Render(opt.Label)
		if i == m.selModel.cursor {
			cursor = selectedStyle.Render("▸ ")
			label = selectedStyle.Render(opt.Label)
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, label))
		if opt.Description != "" {
			b.WriteString(descStyle.Render(opt.Description))
			b.WriteByte('\n')
		}
	}
	b.WriteString(hintStyle.Render("↑/↓ navigate  •  enter select  •  esc cancel"))
	return b.String()
}

func (m wizardModel) renderMulti() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(m.multiModel.title))
	b.WriteByte('\n')

	for i, opt := range m.multiModel.options {
		if i > 0 {
			b.WriteByte('\n')
		}
		cursor := "  "
		if i == m.multiModel.cursor {
			cursor = selectedStyle.Render("▸ ")
		}
		check := uncheckStyle.Render("[ ]")
		label := labelStyle.Render(opt.Label)
		if m.multiModel.selected[i] {
			check = checkStyle.Render("[✓]")
		}
		if i == m.multiModel.cursor {
			label = selectedStyle.Render(opt.Label)
		}
		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, check, label))
		if opt.Description != "" {
			b.WriteString(descStyle.Render(opt.Description))
			b.WriteByte('\n')
		}
	}
	b.WriteString(hintStyle.Render("↑/↓ navigate  •  space toggle  •  enter confirm  •  esc cancel"))
	return b.String()
}
