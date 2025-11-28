package apl

import "gopkg.in/yaml.v3"

// File represents one rotation YAML file.
type File struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Imports     []string           `yaml:"imports"`
	Variables   map[string]any     `yaml:"variables"`
	Rotation    []ActionDefinition `yaml:"rotation"`
}

// ActionDefinition describes one entry in the priority list.
type ActionDefinition struct {
	Action          string             `yaml:"action"`
	Spell           string             `yaml:"spell,omitempty"`
	Item            string             `yaml:"item,omitempty"`
	DurationSeconds float64            `yaml:"duration_seconds,omitempty"`
	Steps           []ActionDefinition `yaml:"steps,omitempty"`
	Tags            []string           `yaml:"tags,omitempty"`
	When            *ConditionNode     `yaml:"when,omitempty"`
}

// ConditionNode captures the raw YAML tree for conditions.
// We keep the node so the compiler can interpret it later.
type ConditionNode struct {
	raw *yaml.Node
}

// Node exposes the underlying YAML node.
func (c *ConditionNode) Node() *yaml.Node {
	if c == nil {
		return nil
	}
	return c.raw
}

// UnmarshalYAML stores the condition tree verbatim.
func (c *ConditionNode) UnmarshalYAML(value *yaml.Node) error {
	c.raw = value
	return nil
}

// NewConditionNode wraps a YAML condition node.
func NewConditionNode(node *yaml.Node) *ConditionNode {
	return &ConditionNode{raw: node}
}
