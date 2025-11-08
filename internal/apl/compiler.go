package apl

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// CompiledRotation is the runtime representation of an APL file.
type CompiledRotation struct {
	Name        string
	Description string
	Variables   map[string]any
	Actions     []*Action
}

// ActionType enumerates supported rotation actions.
type ActionType int

const (
	ActionCastSpell ActionType = iota
	ActionUseItem
	ActionWait
	ActionMacro
)

// Action is a compiled, ready-to-evaluate rotation entry.
type Action struct {
	Type      ActionType
	Spell     string
	Item      string
	Duration  time.Duration
	Steps     []*Action
	Condition Condition
	Tags      []string
}

// Compile turns a parsed File into a CompiledRotation.
func Compile(file *File) (*CompiledRotation, error) {
	if file == nil {
		return nil, fmt.Errorf("nil rotation file")
	}
	var actions []*Action
	for idx, def := range file.Rotation {
		action, err := compileAction(&def, file.Variables)
		if err != nil {
			return nil, fmt.Errorf("rotation entry %d: %w", idx, err)
		}
		actions = append(actions, action)
	}
	return &CompiledRotation{
		Name:        file.Name,
		Description: file.Description,
		Variables:   file.Variables,
		Actions:     actions,
	}, nil
}

func compileAction(def *ActionDefinition, vars map[string]any) (*Action, error) {
	if def == nil {
		return nil, fmt.Errorf("nil action")
	}
	action := &Action{
		Tags: def.Tags,
	}
	var err error
	action.Condition, err = compileCondition(def.When, vars)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(def.Action) {
	case "cast_spell", "cast":
		if def.Spell == "" {
			return nil, fmt.Errorf("cast_spell action requires 'spell'")
		}
		spellName, err := validateSpellName(def.Spell)
		if err != nil {
			return nil, err
		}
		action.Type = ActionCastSpell
		action.Spell = spellName
	case "use_item":
		if def.Item == "" {
			return nil, fmt.Errorf("use_item action requires 'item'")
		}
		// TODO: validate item names once we support them
		action.Type = ActionUseItem
		action.Item = normalizeName(def.Item)
	case "wait":
		if def.DurationSeconds <= 0 {
			return nil, fmt.Errorf("wait action requires duration_seconds > 0")
		}
		action.Type = ActionWait
		action.Duration = time.Duration(def.DurationSeconds * float64(time.Second))
	case "macro":
		action.Type = ActionMacro
		for stepIdx := range def.Steps {
			step, err := compileAction(&def.Steps[stepIdx], vars)
			if err != nil {
				return nil, fmt.Errorf("macro step %d: %w", stepIdx, err)
			}
			action.Steps = append(action.Steps, step)
		}
	default:
		return nil, fmt.Errorf("unsupported action '%s'", def.Action)
	}

	return action, nil
}

func compileCondition(node *ConditionNode, vars map[string]any) (Condition, error) {
	if node == nil || node.Node() == nil {
		return trueCondition{}, nil
	}
	return parseConditionNode(node.Node(), vars)
}

func parseConditionNode(node *yaml.Node, vars map[string]any) (Condition, error) {
	switch node.Kind {
	case yaml.MappingNode:
		return parseConditionMapping(node, vars)
	case yaml.SequenceNode:
		// Treat bare sequences as implicit "all"
		children, err := parseConditionSequence(node, vars)
		if err != nil {
			return nil, err
		}
		return allCondition{children: children}, nil
	case yaml.ScalarNode:
		var boolVal bool
		if err := node.Decode(&boolVal); err == nil {
			if boolVal {
				return trueCondition{}, nil
			}
			return falseCondition{}, nil
		}
		return nil, fmt.Errorf("unsupported scalar condition: %s", node.Value)
	default:
		return nil, fmt.Errorf("unsupported YAML node kind %d", node.Kind)
	}
}

func parseConditionMapping(node *yaml.Node, vars map[string]any) (Condition, error) {
	if len(node.Content)%2 != 0 || len(node.Content) == 0 {
		return nil, fmt.Errorf("condition mapping must have key/value pairs")
	}
	if len(node.Content) != 2 {
		return nil, fmt.Errorf("condition mapping must have exactly one entry")
	}

	key := node.Content[0].Value
	val := node.Content[1]

	switch key {
	case "all":
		children, err := parseConditionSequence(val, vars)
		if err != nil {
			return nil, fmt.Errorf("all: %w", err)
		}
		return allCondition{children: children}, nil
	case "any":
		children, err := parseConditionSequence(val, vars)
		if err != nil {
			return nil, fmt.Errorf("any: %w", err)
		}
		return anyCondition{children: children}, nil
	case "not":
		child, err := parseConditionNode(val, vars)
		if err != nil {
			return nil, fmt.Errorf("not: %w", err)
		}
		return notCondition{child: child}, nil
	case "true":
		return trueCondition{}, nil
	case "false":
		return falseCondition{}, nil
	case "debuff_active":
		params, err := nodeToMap(val)
		if err != nil {
			return nil, err
		}
		nameRaw, err := stringField(params, "debuff", true, vars)
		if err != nil {
			return nil, err
		}
		name, err := validateDebuffName(nameRaw)
		if err != nil {
			return nil, err
		}
		minDur, err := durationField(params, "min_remaining", vars)
		if err != nil {
			return nil, err
		}
		maxDur, err := durationField(params, "max_remaining", vars)
		if err != nil {
			return nil, err
		}
		return debuffActiveCondition{name: strings.ToLower(name), minRemaining: minDur, maxRemaining: maxDur}, nil
	case "dot_remaining":
		params, err := nodeToMap(val)
		if err != nil {
			return nil, err
		}
		spellRaw, err := stringField(params, "spell", true, vars)
		if err != nil {
			return nil, err
		}
		spell, err := validateDebuffName(spellRaw)
		if err != nil {
			return nil, err
		}
		cond := dotRemainingCondition{spell: spell}
		if cond.lt, err = durationField(params, "lt_seconds", vars); err != nil {
			return nil, err
		}
		if cond.lte, err = durationField(params, "lte_seconds", vars); err != nil {
			return nil, err
		}
		if cond.gt, err = durationField(params, "gt_seconds", vars); err != nil {
			return nil, err
		}
		if cond.gte, err = durationField(params, "gte_seconds", vars); err != nil {
			return nil, err
		}
		return cond, nil
	case "buff_active":
		params, err := nodeToMap(val)
		if err != nil {
			return nil, err
		}
		rawName, err := stringField(params, "buff", true, vars)
		if err != nil {
			return nil, err
		}
		name, err := validateBuffName(rawName)
		if err != nil {
			return nil, err
		}
		cond := buffActiveCondition{name: name}
		if cond.minRemaining, err = durationField(params, "min_remaining", vars); err != nil {
			return nil, err
		}
		if cond.maxRemaining, err = durationField(params, "max_remaining", vars); err != nil {
			return nil, err
		}
		return cond, nil
	case "resource_percent":
		params, err := nodeToMap(val)
		if err != nil {
			return nil, err
		}
		resRaw, err := stringField(params, "resource", true, vars)
		if err != nil {
			return nil, err
		}
		res, err := validateResourceName(resRaw)
		if err != nil {
			return nil, err
		}
		cond := resourcePercentCondition{resource: res}
		if cond.lt, err = floatField(params, "lt", vars); err != nil {
			return nil, err
		}
		if cond.lte, err = floatField(params, "lte", vars); err != nil {
			return nil, err
		}
		if cond.gt, err = floatField(params, "gt", vars); err != nil {
			return nil, err
		}
		if cond.gte, err = floatField(params, "gte", vars); err != nil {
			return nil, err
		}
		return cond, nil
	case "cooldown_ready":
		params, err := nodeToMap(val)
		if err != nil {
			return nil, err
		}
		name, err := stringField(params, "spell", false, vars)
		if err != nil {
			return nil, err
		}
		isItem := false
		if name == "" {
			name, err = stringField(params, "item", true, vars)
			if err != nil {
				return nil, err
			}
			isItem = true
		}
		name = normalizeName(name)
		if !isItem {
			if name, err = validateCooldownName(name); err != nil {
				return nil, err
			}
		}
		return cooldownReadyCondition{name: name}, nil
	case "cooldown_remaining":
		params, err := nodeToMap(val)
		if err != nil {
			return nil, err
		}
		name, err := stringField(params, "spell", false, vars)
		if err != nil {
			return nil, err
		}
		isItem := false
		if name == "" {
			name, err = stringField(params, "item", true, vars)
			if err != nil {
				return nil, err
			}
			isItem = true
		}
		name = normalizeName(name)
		if !isItem {
			if name, err = validateCooldownName(name); err != nil {
				return nil, err
			}
		}
		cond := cooldownRemainingCondition{name: name}
		if cond.lt, err = durationField(params, "lt_seconds", vars); err != nil {
			return nil, err
		}
		if cond.lte, err = durationField(params, "lte_seconds", vars); err != nil {
			return nil, err
		}
		if cond.gt, err = durationField(params, "gt_seconds", vars); err != nil {
			return nil, err
		}
		if cond.gte, err = durationField(params, "gte_seconds", vars); err != nil {
			return nil, err
		}
		return cond, nil
	case "charges":
		params, err := nodeToMap(val)
		if err != nil {
			return nil, err
		}
		buffRaw, err := stringField(params, "buff", true, vars)
		if err != nil {
			return nil, err
		}
		buff, err := validateBuffName(buffRaw)
		if err != nil {
			return nil, err
		}
		cond := chargesCondition{buff: buff}
		if cond.lt, err = intField(params, "lt", vars); err != nil {
			return nil, err
		}
		if cond.lte, err = intField(params, "lte", vars); err != nil {
			return nil, err
		}
		if cond.gt, err = intField(params, "gt", vars); err != nil {
			return nil, err
		}
		if cond.gte, err = intField(params, "gte", vars); err != nil {
			return nil, err
		}
		return cond, nil
	default:
		return nil, fmt.Errorf("unknown condition '%s'", key)
	}
}

func parseConditionSequence(node *yaml.Node, vars map[string]any) ([]Condition, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("expected sequence, got %d", node.Kind)
	}
	children := make([]Condition, 0, len(node.Content))
	for idx, childNode := range node.Content {
		child, err := parseConditionNode(childNode, vars)
		if err != nil {
			return nil, fmt.Errorf("condition %d: %w", idx, err)
		}
		children = append(children, child)
	}
	return children, nil
}

func nodeToMap(node *yaml.Node) (map[string]*yaml.Node, error) {
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected mapping node, got %d", node.Kind)
	}
	result := make(map[string]*yaml.Node, len(node.Content)/2)
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		result[key] = node.Content[i+1]
	}
	return result, nil
}

func stringField(fields map[string]*yaml.Node, key string, required bool, vars map[string]any) (string, error) {
	node, ok := fields[key]
	if !ok {
		if required {
			return "", fmt.Errorf("missing field '%s'", key)
		}
		return "", nil
	}
	val, err := resolveScalar(node, vars)
	if err != nil {
		return "", err
	}
	switch v := val.(type) {
	case string:
		return v, nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func durationField(fields map[string]*yaml.Node, key string, vars map[string]any) (*time.Duration, error) {
	val, err := floatField(fields, key, vars)
	if err != nil || val == nil {
		return nil, err
	}
	d := time.Duration(*val * float64(time.Second))
	return &d, nil
}

func floatField(fields map[string]*yaml.Node, key string, vars map[string]any) (*float64, error) {
	node, ok := fields[key]
	if !ok {
		return nil, nil
	}
	val, err := resolveScalar(node, vars)
	if err != nil {
		return nil, err
	}
	switch v := val.(type) {
	case float64:
		return &v, nil
	case int:
		f := float64(v)
		return &f, nil
	case int64:
		f := float64(v)
		return &f, nil
	case uint64:
		f := float64(v)
		return &f, nil
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return nil, err
		}
		return &parsed, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to float for key '%s'", v, key)
	}
}

func intField(fields map[string]*yaml.Node, key string, vars map[string]any) (*int, error) {
	node, ok := fields[key]
	if !ok {
		return nil, nil
	}
	val, err := resolveScalar(node, vars)
	if err != nil {
		return nil, err
	}
	switch v := val.(type) {
	case int:
		return &v, nil
	case int64:
		c := int(v)
		return &c, nil
	case uint64:
		c := int(v)
		return &c, nil
	case float64:
		c := int(v)
		return &c, nil
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return nil, err
		}
		return &parsed, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to int for key '%s'", v, key)
	}
}

func resolveScalar(node *yaml.Node, vars map[string]any) (interface{}, error) {
	if node == nil {
		return nil, fmt.Errorf("nil scalar")
	}
	var out interface{}
	if err := node.Decode(&out); err != nil {
		return nil, err
	}
	if str, ok := out.(string); ok {
		str = strings.TrimSpace(str)
		if strings.HasPrefix(str, "${") && strings.HasSuffix(str, "}") {
			name := strings.TrimSpace(str[2 : len(str)-1])
			if vars == nil {
				return nil, fmt.Errorf("variable '%s' not defined", name)
			}
			val, ok := vars[name]
			if !ok {
				return nil, fmt.Errorf("variable '%s' not defined", name)
			}
			return val, nil
		}
	}
	return out, nil
}
