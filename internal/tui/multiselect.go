package tui

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/inventage-ai/asylum/internal/term"
)

// MultiSelect shows a multi-choice prompt and returns selected indices.
// Returns defaultSelected without prompting if stdin is not a TTY.
func MultiSelect(title string, options []Option, defaultSelected []int) ([]int, error) {
	if !term.IsTerminal() {
		return defaultSelected, nil
	}

	selected := map[int]bool{}
	for _, i := range defaultSelected {
		selected[i] = true
	}

	m := multiModel{
		title:    title,
		options:  options,
		selected: selected,
	}

	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return nil, err
	}

	final := result.(multiModel)
	if final.cancelled {
		return nil, ErrCancelled
	}

	var indices []int
	for i := range final.options {
		if final.selected[i] {
			indices = append(indices, i)
		}
	}
	return indices, nil
}

type multiModel struct {
	title     string
	options   []Option
	cursor    int
	selected  map[int]bool
	cancelled bool
	height    int // terminal height, 0 until known
	offset    int // first visible option
}

// multiChrome is the number of lines the box takes besides the options:
// surrounding newlines, border, padding, title and hint with their margins.
const multiChrome = 10

func (m multiModel) Init() tea.Cmd { return nil }

func (m multiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case " ":
			m.selected[m.cursor] = !m.selected[m.cursor]
		case "enter":
			return m, tea.Quit
		case "esc", "ctrl+c", "q":
			m.cancelled = true
			return m, tea.Quit
		}
	}
	vis := m.visible()
	m.offset = min(max(m.offset, m.cursor-vis+1), m.cursor)
	return m, nil
}

// visible returns how many options fit into the terminal.
func (m multiModel) visible() int {
	if m.height == 0 {
		return len(m.options)
	}
	perOption := 2
	if slices.ContainsFunc(m.options, func(o Option) bool { return o.Description != "" }) {
		perOption = 3
	}
	// The last option has no blank line after it.
	return min(max((m.height-multiChrome+1)/perOption, 1), len(m.options))
}

var (
	checkStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	uncheckStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

func (m multiModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(m.title))
	b.WriteByte('\n')

	vis := m.visible()
	for i := m.offset; i < m.offset+vis; i++ {
		opt := m.options[i]
		if i > m.offset {
			b.WriteByte('\n')
		}

		cursor := "  "
		if i == m.cursor {
			cursor = selectedStyle.Render("▸ ")
		}

		check := uncheckStyle.Render("[ ]")
		label := labelStyle.Render(opt.Label)
		if m.selected[i] {
			check = checkStyle.Render("[✓]")
		}
		if i == m.cursor {
			label = selectedStyle.Render(opt.Label)
		}

		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, check, label))
		if opt.Description != "" {
			b.WriteString(descStyle.Render(opt.Description))
			b.WriteByte('\n')
		}
	}

	hint := "↑/↓ navigate  •  space toggle  •  enter confirm  •  esc cancel"
	if vis < len(m.options) {
		hint += fmt.Sprintf("  •  %d–%d of %d", m.offset+1, m.offset+vis, len(m.options))
	}
	b.WriteString(hintStyle.Render(hint))

	return "\n" + borderStyle.Render(b.String()) + "\n"
}
