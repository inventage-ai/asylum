package config

import (
	"bytes"
	"os"

	"gopkg.in/yaml.v3"
)

// SyncKitToConfig inserts a kit's config nodes into the config file's kits
// mapping. If the kit key already exists, no modification is made.
// nodes must be a [key, value] pair of yaml.Node pointers.
func SyncKitToConfig(path string, kitName string, nodes []*yaml.Node) error {
	doc, err := parseConfigDoc(path)
	if err != nil {
		return err
	}

	kitsNode := findOrCreateKitsMapping(doc)
	if kitExistsInMapping(kitsNode, kitName) {
		return nil
	}

	hadContent := len(kitsNode.Content) > 0
	kitsNode.Content = append(kitsNode.Content, nodes...)

	if err := writeConfigDoc(path, doc); err != nil {
		return err
	}

	// Add blank line before the new kit entry for readability
	if hadContent {
		return spaceKitEntries(path)
	}
	return nil
}

// SyncKitCommentToConfig appends a commented-out kit block to the config
// file's kits mapping as a foot comment.
func SyncKitCommentToConfig(path string, comment string) error {
	doc, err := parseConfigDoc(path)
	if err != nil {
		return err
	}

	kitsNode := findOrCreateKitsMapping(doc)

	// Append as foot comment on the kits mapping node
	if kitsNode.FootComment != "" {
		kitsNode.FootComment += "\n\n" + comment
	} else {
		kitsNode.FootComment = comment
	}
	return writeConfigDoc(path, doc)
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

// writeConfigDoc encodes a yaml.Node document tree back to a file.
func writeConfigDoc(path string, doc *yaml.Node) error {
	data, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// spaceKitEntries reads the config file and inserts blank lines between
// top-level entries in the kits mapping block.
func spaceKitEntries(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := bytes.Split(data, []byte("\n"))

	// Find the "kits:" line
	kitsIdx := -1
	for i, line := range lines {
		if bytes.HasPrefix(bytes.TrimLeft(line, " "), []byte("kits:")) {
			kitsIdx = i
			break
		}
	}
	if kitsIdx < 0 || kitsIdx+1 >= len(lines) {
		return nil
	}

	// Detect kit entry indent from the first entry after "kits:"
	kitIndent := -1
	for _, line := range lines[kitsIdx+1:] {
		if len(bytes.TrimSpace(line)) > 0 {
			kitIndent = len(line) - len(bytes.TrimLeft(line, " "))
			break
		}
	}
	if kitIndent < 0 {
		return nil
	}

	kitsLineIndent := len(lines[kitsIdx]) - len(bytes.TrimLeft(lines[kitsIdx], " "))
	var result [][]byte
	firstKit := true

	for i, line := range lines {
		if i > kitsIdx && len(bytes.TrimSpace(line)) > 0 {
			indent := len(line) - len(bytes.TrimLeft(line, " "))

			// Past the kits block
			if indent <= kitsLineIndent {
				result = append(result, lines[i:]...)
				return os.WriteFile(path, bytes.Join(result, []byte("\n")), 0644)
			}

			// Kit-level entry
			if indent == kitIndent && bytes.Contains(line, []byte(":")) {
				if !firstKit && len(result) > 0 && len(bytes.TrimSpace(result[len(result)-1])) > 0 {
					result = append(result, []byte(""))
				}
				firstKit = false
			}
		}
		result = append(result, line)
	}

	return os.WriteFile(path, bytes.Join(result, []byte("\n")), 0644)
}

// findOrCreateKitsMapping walks the document to find the "kits" mapping node.
// If it doesn't exist, it creates one and appends it to the root mapping.
func findOrCreateKitsMapping(doc *yaml.Node) *yaml.Node {
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		// Create minimal document structure
		root := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		doc.Kind = yaml.DocumentNode
		doc.Content = []*yaml.Node{root}
	}

	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	}

	// Walk mapping pairs looking for "kits" key
	for i := 0; i < len(root.Content)-1; i += 2 {
		if root.Content[i].Value == "kits" {
			if root.Content[i+1].Kind == yaml.MappingNode {
				return root.Content[i+1]
			}
			// Key exists but value isn't a mapping — replace it
			root.Content[i+1] = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			return root.Content[i+1]
		}
	}

	// No "kits" key — create it
	kitsKey := &yaml.Node{Kind: yaml.ScalarNode, Value: "kits", Tag: "!!str"}
	kitsVal := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	root.Content = append(root.Content, kitsKey, kitsVal)
	return kitsVal
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
