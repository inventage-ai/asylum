package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// WriteDefaultResume sets the top-level `default-resume` key in the global
// asylum config file (`<asylumDir>/config.yaml`), preserving all other keys.
//
// The file is loaded as a generic YAML node tree so unknown keys are kept
// intact. If the file does not exist it is created with just this key. If it
// exists but cannot be parsed as a YAML mapping, an error is returned rather
// than overwriting unfamiliar content.
func WriteDefaultResume(asylumDir string, value bool) error {
	path := filepath.Join(asylumDir, "config.yaml")

	if err := os.MkdirAll(asylumDir, 0755); err != nil {
		return fmt.Errorf("create %s: %w", asylumDir, err)
	}

	var root yaml.Node
	data, err := os.ReadFile(path)
	switch {
	case err == nil && len(data) > 0:
		if err := yaml.Unmarshal(data, &root); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
	case err != nil && !os.IsNotExist(err):
		return fmt.Errorf("read %s: %w", path, err)
	}

	doc := ensureMappingDoc(&root)
	setMappingBool(doc, "default-resume", value)

	out, err := yaml.Marshal(&root)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	return os.WriteFile(path, out, 0644)
}

// ensureMappingDoc returns the mapping node inside a YAML document, creating
// an empty mapping at the document root if necessary.
func ensureMappingDoc(root *yaml.Node) *yaml.Node {
	if root.Kind == 0 {
		root.Kind = yaml.DocumentNode
	}
	if len(root.Content) == 0 {
		root.Content = []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}
	}
	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		// Refuse to clobber a non-mapping document (e.g. a YAML scalar).
		return nil
	}
	return doc
}

// setMappingBool sets `key: value` on a YAML mapping node, replacing any
// existing value for that key. Returns silently if the doc is nil (refused
// upstream because the document root was not a mapping).
func setMappingBool(doc *yaml.Node, key string, value bool) {
	if doc == nil {
		return
	}
	v := "false"
	if value {
		v = "true"
	}
	for i := 0; i < len(doc.Content); i += 2 {
		if doc.Content[i].Value == key {
			doc.Content[i+1].Kind = yaml.ScalarNode
			doc.Content[i+1].Tag = "!!bool"
			doc.Content[i+1].Value = v
			return
		}
	}
	doc.Content = append(doc.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: v},
	)
}
