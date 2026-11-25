package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMultiSelectFitsTerminal(t *testing.T) {
	var options []Option
	for i := range 30 {
		options = append(options, Option{Label: fmt.Sprintf("project-%02d", i), Description: "idle"})
	}
	var model tea.Model = multiModel{title: "Pick", options: options, selected: map[int]bool{}}
	model, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	for range 25 {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	}

	view := model.View()
	if lines := strings.Count(view, "\n"); lines > 24 {
		t.Errorf("view has %d lines, terminal has 24:\n%s", lines, view)
	}
	if !strings.Contains(view, "project-25") || strings.Contains(view, "project-00") {
		t.Errorf("view does not follow the cursor:\n%s", view)
	}
	if !strings.Contains(view, "of 30") {
		t.Errorf("view has no position hint:\n%s", view)
	}

	for range 25 {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	}
	if view := model.View(); !strings.Contains(view, "project-00") {
		t.Errorf("view does not scroll back up:\n%s", view)
	}
}

func TestMultiSelectWithoutSizeShowsAll(t *testing.T) {
	m := multiModel{options: []Option{{Label: "a"}, {Label: "b"}}, selected: map[int]bool{}}
	if view := m.View(); !strings.Contains(view, "a") || !strings.Contains(view, "b") || strings.Contains(view, " of ") {
		t.Errorf("unexpected view:\n%s", view)
	}
}
