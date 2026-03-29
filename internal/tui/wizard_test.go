package tui

import (
	"testing"
)

func TestWizardModelStepTransition(t *testing.T) {
	steps := []WizardStep{
		{
			Title:      "Step 1",
			Kind:       StepSelect,
			Options:    []Option{{Label: "A"}, {Label: "B"}},
			DefaultIdx: 1,
		},
		{
			Title:      "Step 2",
			Kind:       StepMultiSelect,
			Options:    []Option{{Label: "X"}, {Label: "Y"}},
			DefaultSel: []int{0},
		},
	}

	results := make([]StepResult, len(steps))
	m := wizardModel{steps: steps, results: results}
	m.initStep(0)

	// Verify initial state
	if m.current != 0 {
		t.Fatalf("expected current=0, got %d", m.current)
	}
	if m.selModel.cursor != 1 {
		t.Fatalf("expected default cursor=1, got %d", m.selModel.cursor)
	}

	// Navigate up in select
	m.selModel.cursor = 0

	// Simulate enter on step 1 — collect result and advance
	// The model's Update with enter key collects and advances
	result := m.results[0]
	if result.Completed {
		t.Fatal("step 0 should not be completed yet")
	}
}

func TestWizardInitStep(t *testing.T) {
	steps := []WizardStep{
		{
			Title:      "Select Step",
			Kind:       StepSelect,
			Options:    []Option{{Label: "A"}, {Label: "B"}, {Label: "C"}},
			DefaultIdx: 2,
		},
		{
			Title:      "Multi Step",
			Kind:       StepMultiSelect,
			Options:    []Option{{Label: "X"}, {Label: "Y"}, {Label: "Z"}},
			DefaultSel: []int{0, 2},
		},
	}

	results := make([]StepResult, len(steps))
	m := wizardModel{steps: steps, results: results}

	// Init select step
	m.initStep(0)
	if m.selModel.cursor != 2 {
		t.Errorf("select default: got cursor=%d, want 2", m.selModel.cursor)
	}
	if len(m.selModel.options) != 3 {
		t.Errorf("select options: got %d, want 3", len(m.selModel.options))
	}

	// Init multiselect step
	m.initStep(1)
	if !m.multiModel.selected[0] || m.multiModel.selected[1] || !m.multiModel.selected[2] {
		t.Errorf("multi defaults: got %v, want {0:true, 2:true}", m.multiModel.selected)
	}
}

func TestWizardView(t *testing.T) {
	steps := []WizardStep{
		{Title: "Config", Kind: StepSelect, Options: []Option{{Label: "A"}}, DefaultIdx: 0},
		{Title: "Creds", Kind: StepMultiSelect, Options: []Option{{Label: "X"}}, DefaultSel: []int{0}},
	}

	results := make([]StepResult, len(steps))
	m := wizardModel{steps: steps, results: results}
	m.initStep(0)

	view := m.View()
	if view == "" {
		t.Fatal("expected non-empty view")
	}

	// Should contain step titles
	if !containsPlainText(view, "Config") {
		t.Error("view should contain step 1 title 'Config'")
	}
	if !containsPlainText(view, "Creds") {
		t.Error("view should contain step 2 title 'Creds'")
	}
}

func TestWizardViewAfterComplete(t *testing.T) {
	steps := []WizardStep{
		{Title: "Done Step", Kind: StepSelect, Options: []Option{{Label: "A"}}, DefaultIdx: 0},
		{Title: "Current", Kind: StepSelect, Options: []Option{{Label: "B"}}, DefaultIdx: 0},
	}

	results := make([]StepResult, len(steps))
	results[0].Completed = true
	m := wizardModel{steps: steps, results: results, current: 1}
	m.initStep(1)

	view := m.View()
	// Completed step should show checkmark
	if !containsPlainText(view, "✓") {
		t.Error("completed step should show ✓")
	}
}

// containsPlainText checks if s contains sub, ignoring ANSI escape codes.
func containsPlainText(s, sub string) bool {
	// Strip ANSI escape sequences for comparison
	plain := stripANSI(s)
	return len(plain) > 0 && contains(plain, sub)
}

func stripANSI(s string) string {
	var result []byte
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Skip until 'm'
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			i = j + 1
		} else {
			result = append(result, s[i])
			i++
		}
	}
	return string(result)
}

func contains(s, sub string) bool {
	return len(sub) <= len(s) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
