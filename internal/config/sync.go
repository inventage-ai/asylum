package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	goyaml "github.com/goccy/go-yaml/ast"
	goyamlparser "github.com/goccy/go-yaml/parser"
	goyamltoken "github.com/goccy/go-yaml/token"
	"gopkg.in/yaml.v3"
)

// SyncKitToConfig inserts a kit's config snippet into the config file's kits
// block using text-based insertion (no YAML roundtrip, so comments and
// indentation are preserved). If the kit key already exists, no modification
// is made.
func SyncKitToConfig(path string, kitName string, snippet string) error {
	// Parse YAML read-only to check if kit already exists.
	doc, err := parseConfigDoc(path)
	if err != nil {
		return err
	}
	if kitsNode := findKitsMapping(doc); kitsNode != nil {
		if kitExistsInMapping(kitsNode, kitName) {
			return nil
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")

	// Find the "kits:" line.
	kitsIdx := -1
	kitsLineIndent := 0
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		if strings.HasPrefix(trimmed, "kits:") {
			kitsIdx = i
			kitsLineIndent = len(line) - len(trimmed)
			break
		}
	}
	if kitsIdx < 0 {
		return fmt.Errorf("no kits: mapping found in %s", path)
	}

	entryIndent := kitsLineIndent + 2

	// Find insertion point: after the last active kit entry's full block,
	// before any commented-out kit entries at the entry indent level.
	insertIdx := kitsIdx + 1
	inKitBlock := false
	for i := kitsIdx + 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))
		isComment := strings.HasPrefix(trimmed, "#")

		// Past the kits block entirely.
		if indent <= kitsLineIndent && !isComment {
			break
		}

		if indent == entryIndent {
			if isComment {
				break // commented-out kit entry — insert before this
			}
			inKitBlock = true
			insertIdx = i + 1
			continue
		}

		// Deeper than entry level: part of the current kit's config.
		if inKitBlock {
			insertIdx = i + 1
		}
	}

	snippetText := strings.TrimRight(snippet, "\n")
	var insert []string
	if insertIdx > kitsIdx+1 {
		insert = append(insert, "") // blank line separator
	}
	insert = append(insert, snippetText)

	result := make([]string, 0, len(lines)+len(insert))
	result = append(result, lines[:insertIdx]...)
	result = append(result, insert...)
	result = append(result, lines[insertIdx:]...)

	return os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644)
}

// RemoveKitComment removes a commented-out kit block from the kits section.
// It finds the comment line matching "# <name>:" at kit entry indent (2 spaces)
// and removes it plus any subsequent deeper-indented comment lines.
func RemoveKitComment(path string, kitName string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")

	kitsIdx := -1
	kitsLineIndent := 0
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		if strings.HasPrefix(trimmed, "kits:") {
			kitsIdx = i
			kitsLineIndent = len(line) - len(trimmed)
			break
		}
	}
	if kitsIdx < 0 {
		return nil
	}

	entryIndent := kitsLineIndent + 2
	prefix := strings.Repeat(" ", entryIndent) + "# " + kitName + ":"

	// Find the commented kit line.
	startIdx := -1
	for i := kitsIdx + 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent <= kitsLineIndent && !strings.HasPrefix(trimmed, "#") {
			break
		}
		if strings.HasPrefix(line, prefix) {
			startIdx = i
			break
		}
	}
	if startIdx < 0 {
		return nil
	}

	// Find the end of the commented block: subsequent comment lines at deeper indent.
	endIdx := startIdx + 1
	for endIdx < len(lines) {
		line := lines[endIdx]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			endIdx++
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if !strings.HasPrefix(trimmed, "#") || indent <= entryIndent {
			break
		}
		endIdx++
	}

	// Remove any leading blank line before the block.
	if startIdx > 0 && strings.TrimSpace(lines[startIdx-1]) == "" {
		startIdx--
	}

	result := make([]string, 0, len(lines)-(endIdx-startIdx))
	result = append(result, lines[:startIdx]...)
	result = append(result, lines[endIdx:]...)

	return os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644)
}


// SyncKitCommentToConfig appends a commented-out kit entry to the end of
// the kits block, preserving file formatting via goccy/go-yaml's AST.
func SyncKitCommentToConfig(path string, comment string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	f, err := goyamlparser.ParseBytes(data, goyamlparser.ParseComments)
	if err != nil {
		return err
	}

	// Find the kits mapping in the AST.
	kitsMapping := findKitsMappingGoccy(f)
	if kitsMapping == nil || len(kitsMapping.Values) == 0 {
		return fmt.Errorf("no kits mapping found in %s", path)
	}

	// Append comment to the FootComment of the last entry in the kits mapping.
	// Column 3 (1-indexed) gives 2-space indentation matching kit entries.
	// The AST can't produce a true empty line between comments, so we insert
	// a sentinel that gets replaced with a blank line in the final output.
	const blankSentinel = "<BLANK>"
	last := kitsMapping.Values[len(kitsMapping.Values)-1]
	blank := &goyaml.CommentNode{BaseNode: &goyaml.BaseNode{}, Token: goyamltoken.New(blankSentinel, "", &goyamltoken.Position{Column: 3})}
	tok := goyamltoken.New(" "+comment, "", &goyamltoken.Position{Column: 3})
	cn := &goyaml.CommentNode{BaseNode: &goyaml.BaseNode{}, Token: tok}

	if last.FootComment != nil {
		last.FootComment.Comments = append(last.FootComment.Comments, blank, cn)
	} else {
		last.FootComment = &goyaml.CommentGroupNode{
			BaseNode: &goyaml.BaseNode{},
			Comments: []*goyaml.CommentNode{blank, cn},
		}
	}

	out := regexp.MustCompile(`(?m)^\s*#`+blankSentinel+`$`).ReplaceAllString(f.String(), "")
	return os.WriteFile(path, []byte(out), 0644)
}

// findKitsMappingGoccy finds the "kits" mapping node in a goccy/go-yaml AST.
func findKitsMappingGoccy(f *goyaml.File) *goyaml.MappingNode {
	if len(f.Docs) == 0 || f.Docs[0].Body == nil {
		return nil
	}
	root, ok := f.Docs[0].Body.(*goyaml.MappingNode)
	if !ok {
		return nil
	}
	for _, v := range root.Values {
		if v.Key.String() == "kits" {
			if mn, ok := v.Value.(*goyaml.MappingNode); ok {
				return mn
			}
		}
	}
	return nil
}

// parseConfigDoc reads a YAML file into a yaml.Node document tree.
func parseConfigDoc(path string) (*yaml.Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// findKitsMapping walks the document to find the "kits" mapping node.
// Returns nil if not found.
func findKitsMapping(doc *yaml.Node) *yaml.Node {
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return nil
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(root.Content)-1; i += 2 {
		if root.Content[i].Value == "kits" && root.Content[i+1].Kind == yaml.MappingNode {
			return root.Content[i+1]
		}
	}
	return nil
}

// kitExistsInMapping checks if a key is already present in a mapping node.
func kitExistsInMapping(mapping *yaml.Node, name string) bool {
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == name {
			return true
		}
	}
	return false
}

// KitExistsInFile checks whether a kit key exists as an active (non-comment)
// YAML entry in the kits mapping of the given config file.
func KitExistsInFile(path string, kitName string) bool {
	doc, err := parseConfigDoc(path)
	if err != nil {
		return false
	}
	kitsNode := findKitsMapping(doc)
	if kitsNode == nil {
		return false
	}
	return kitExistsInMapping(kitsNode, kitName)
}

// SetKitDisabled adds `disabled: true` as the first property under a kit's
// entry in the config file. If the kit already has `disabled: true`, no
// modification is made. Uses text-based editing to preserve formatting.
func SetKitDisabled(path, kitName string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")

	kitsIdx := findSection(lines, "kits")
	if kitsIdx == -1 {
		return nil
	}
	kitIdx := findKey(lines, kitsIdx, kitName)
	if kitIdx == -1 {
		return nil
	}

	disabledIdx := findKey(lines, kitIdx, "disabled")
	if disabledIdx != -1 {
		if strings.TrimSpace(lines[disabledIdx]) == "disabled: true" {
			return nil // already disabled
		}
		// Replace disabled: false → disabled: true
		indent := leadingSpaces(lines[disabledIdx])
		lines[disabledIdx] = strings.Repeat(" ", indent) + "disabled: true"
		return writeLines(path, lines)
	}

	parentIndent := leadingSpaces(lines[kitIdx])
	lines = insertAfter(lines, kitIdx, parentIndent+2, "disabled: true")
	return writeLines(path, lines)
}

// RemoveKitDisabled removes the `disabled: true` (or `disabled: false`) line
// from a kit's entry in the config file. If no disabled field is present, no
// modification is made.
func RemoveKitDisabled(path, kitName string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")

	kitsIdx := findSection(lines, "kits")
	if kitsIdx == -1 {
		return nil
	}
	kitIdx := findKey(lines, kitsIdx, kitName)
	if kitIdx == -1 {
		return nil
	}
	disabledIdx := findKey(lines, kitIdx, "disabled")
	if disabledIdx == -1 {
		return nil
	}

	result := make([]string, 0, len(lines)-1)
	result = append(result, lines[:disabledIdx]...)
	result = append(result, lines[disabledIdx+1:]...)
	return writeLines(path, result)
}
