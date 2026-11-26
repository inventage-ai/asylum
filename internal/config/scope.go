package config

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/inventage-ai/asylum/internal/kit"
	"gopkg.in/yaml.v3"
)

// layer identifies a config file by how much it is trusted.
type layer int

const (
	layerGlobal  layer = iota // ~/.asylum/config.yaml
	layerLocal                // .asylum.local, personal and never committed
	layerProject              // .asylum, usually committed
)

func (l layer) accepts(s kit.Scope) bool {
	switch s {
	case kit.ScopeGlobal:
		return l == layerGlobal
	case kit.ScopePersonal:
		return l != layerProject
	}
	return true
}

// Rejection is a config entry removed from a file that does not accept it.
type Rejection struct {
	File    string    // config file path
	Line    int       // line of the entry's key
	Path    string    // dotted option path, e.g. kits.docker or agents.claude.seed
	Snippet string    // the entry as YAML, nested under its full path, ready to paste
	Scope   kit.Scope // scope of the option
}

// AcceptedIn names the files that accept the rejected option.
func (r Rejection) AcceptedIn() string {
	if r.Scope == kit.ScopeGlobal {
		return "~/.asylum/config.yaml"
	}
	return ".asylum.local or ~/.asylum/config.yaml"
}

var (
	configScopes = fieldScopes(reflect.TypeFor[Config]())
	kitScopes    = fieldScopes(reflect.TypeFor[KitConfig]())
	agentScopes  = fieldScopes(reflect.TypeFor[AgentConfig]())
)

// fieldScopes maps each field's YAML key to its scope tag.
func fieldScopes(t reflect.Type) map[string]kit.Scope {
	scopes := map[string]kit.Scope{}
	for f := range t.Fields() {
		name, _, _ := strings.Cut(f.Tag.Get("yaml"), ",")
		if name != "" && name != "-" {
			scopes[name] = kit.Scope(f.Tag.Get("scope"))
		}
	}
	return scopes
}

func kitScope(name string) kit.Scope {
	if k := kit.Lookup(name); k != nil {
		return k.Scope
	}
	return kit.ScopeAny
}

func kitOptionScope(kitName, option string) kit.Scope {
	if k := kit.Lookup(kitName); k != nil {
		if s, ok := k.OptionScopes[option]; ok {
			return s
		}
	}
	return kitScopes[option]
}

// scopeFilter removes the entries of one config file that its layer does not
// accept, recording each as a Rejection.
type scopeFilter struct {
	layer    layer
	file     string
	rejected []Rejection
}

// filterScopes removes out-of-scope entries from a parsed config document.
// Aliases, merge keys and non-string keys are refused wherever entries are
// checked, because they could pull an entry in from a key the filter does not
// inspect or decode to a different key than the one checked.
func filterScopes(doc *yaml.Node, l layer, file string) ([]Rejection, error) {
	if l == layerGlobal || doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return nil, nil
	}
	f := &scopeFilter{layer: l, file: file}
	err := f.mapping(doc.Content[0], "", func(key string, _ *yaml.Node) kit.Scope { return configScopes[key] },
		func(key string, value *yaml.Node) error {
			switch key {
			case "kits":
				return f.mapping(value, "kits.", kitEntryScope, func(name string, kc *yaml.Node) error {
					return f.mapping(kc, "kits."+name+".", func(option string, _ *yaml.Node) kit.Scope { return kitOptionScope(name, option) }, nil)
				})
			case "agents":
				return f.mapping(value, "agents.", nil, func(name string, ac *yaml.Node) error {
					return f.mapping(ac, "agents."+name+".", func(option string, _ *yaml.Node) kit.Scope { return agentScopes[option] }, nil)
				})
			}
			return nil
		})
	return f.rejected, err
}

// kitEntryScope is the scope of a kit's entry under kits. Disabling a kit only
// narrows the sandbox, so an entry that only disables it is accepted anywhere.
func kitEntryScope(name string, value *yaml.Node) kit.Scope {
	var disabled bool
	if value.Kind == yaml.MappingNode && len(value.Content) == 2 && value.Content[0].Value == "disabled" &&
		value.Content[1].Decode(&disabled) == nil && disabled {
		return kit.ScopeAny
	}
	return kitScope(name)
}

// mapping filters the entries of m by scopeOf (nil accepts all) and calls
// descend (nil skips) for each accepted entry. Non-mapping values are left to
// the decoder.
func (f *scopeFilter) mapping(m *yaml.Node, prefix string, scopeOf func(string, *yaml.Node) kit.Scope, descend func(string, *yaml.Node) error) error {
	if m.Kind == yaml.AliasNode {
		return fmt.Errorf("%s:%d: YAML aliases are not supported for %s", f.file, m.Line, strings.TrimSuffix(prefix, "."))
	}
	if m.Kind != yaml.MappingNode {
		return nil
	}
	kept := m.Content[:0]
	for i := 0; i+1 < len(m.Content); i += 2 {
		k, v := m.Content[i], m.Content[i+1]
		// The scope is looked up by the key's text, so that text must be what
		// the decoder sees: aliases, merge keys and tags such as !!binary
		// would make the two differ.
		if k.Kind != yaml.ScalarNode || k.ShortTag() != "!!str" {
			return fmt.Errorf("%s:%d: only plain string keys are supported here (no YAML aliases, merge keys or tags)", f.file, k.Line)
		}
		if scopeOf != nil {
			if s := scopeOf(k.Value, v); !f.layer.accepts(s) {
				f.reject(prefix+k.Value, k, v, s)
				continue
			}
		}
		if descend != nil {
			if err := descend(k.Value, v); err != nil {
				return err
			}
		}
		kept = append(kept, k, v)
	}
	m.Content = kept
	return nil
}

func (f *scopeFilter) reject(path string, key, value *yaml.Node, s kit.Scope) {
	f.rejected = append(f.rejected, Rejection{
		File:    f.file,
		Line:    key.Line,
		Path:    path,
		Snippet: snippet(path, value),
		Scope:   s,
	})
}

// snippet renders value nested under every segment of path, so it can be
// pasted into another config file as is.
func snippet(path string, value *yaml.Node) string {
	node := value
	segments := strings.Split(path, ".")
	for i := len(segments) - 1; i >= 0; i-- {
		node = &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: segments[i]}, node,
		}}
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(node); err != nil {
		return path
	}
	return strings.TrimSuffix(buf.String(), "\n")
}
