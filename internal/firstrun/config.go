package firstrun

import (
	"os"

	"gopkg.in/yaml.v3"
)

// SetKitCredentials writes `credentials: <value>` under the named kit in a
// config file, using yaml.Node to preserve comments and formatting.
func SetKitCredentials(path, kitName, value string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return err
	}

	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return nil
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil
	}

	kitsNode := findMapValue(root, "kits")
	if kitsNode == nil {
		root.Content = append(root.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "kits"},
			&yaml.Node{Kind: yaml.MappingNode},
		)
		kitsNode = root.Content[len(root.Content)-1]
	}

	kitNode := findMapValue(kitsNode, kitName)
	if kitNode == nil {
		kitsNode.Content = append(kitsNode.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: kitName},
			&yaml.Node{Kind: yaml.MappingNode},
		)
		kitNode = kitsNode.Content[len(kitsNode.Content)-1]
	}

	// If the kit node is null/empty scalar, convert to mapping
	if kitNode.Kind == yaml.ScalarNode {
		kitNode.Kind = yaml.MappingNode
		kitNode.Value = ""
		kitNode.Tag = ""
		kitNode.Content = nil
	}

	credNode := findMapValue(kitNode, "credentials")
	if credNode == nil {
		kitNode.Content = append(kitNode.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "credentials"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: value},
		)
	} else {
		credNode.Value = value
	}

	out, err := yaml.Marshal(&doc)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}

func findMapValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}
