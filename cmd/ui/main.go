package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"wotlk-destro-sim/internal/apl"
	"wotlk-destro-sim/internal/config"
	"wotlk-destro-sim/internal/runes"
)

//go:embed static/index.html
var content embed.FS

type playerResponse struct {
	Player  config.Player `json:"player"`
	Options struct {
		Rotations []string                   `json:"rotations"`
		Pets      []string                   `json:"pets"`
		Runes     map[string][]string        `json:"runes"`
		Limits    config.MysticEnchantConfig `json:"limits"`
	} `json:"options"`
}

type saveRequest struct {
	Player config.Player `json:"player"`
}

type identifiersResponse struct {
	Spells    []string `json:"spells"`
	Buffs     []string `json:"buffs"`
	Debuffs   []string `json:"debuffs"`
	Resources []string `json:"resources"`
}

type rotationResponse struct {
	Filename string      `json:"filename"`
	File     rotationDTO `json:"file"`
}

type rotationDTO struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Imports     []string       `json:"imports"`
	Variables   map[string]any `json:"variables"`
	Rotation    []actionDTO    `json:"rotation"`
}

type actionDTO struct {
	Action          string        `json:"action"`
	Spell           string        `json:"spell,omitempty"`
	Item            string        `json:"item,omitempty"`
	DurationSeconds float64       `json:"duration_seconds,omitempty"`
	Tags            []string      `json:"tags,omitempty"`
	Steps           []actionDTO   `json:"steps,omitempty"`
	When            *conditionDTO `json:"when,omitempty"`
}

type conditionDTO struct {
	Type string `json:"type"` // all, any, not, buff_active, debuff_active, dot_remaining, cooldown_ready, cooldown_remaining, resource_percent, charges, true, false

	Children []conditionDTO `json:"children,omitempty"` // for all/any/not

	Buff   string `json:"buff,omitempty"`
	Debuff string `json:"debuff,omitempty"`
	Spell  string `json:"spell,omitempty"` // dot_remaining or cooldown names

	Resource string `json:"resource,omitempty"`

	MinRemainingSeconds *float64 `json:"min_remaining_seconds,omitempty"`
	MaxRemainingSeconds *float64 `json:"max_remaining_seconds,omitempty"`

	LtSeconds  *float64 `json:"lt_seconds,omitempty"`
	LteSeconds *float64 `json:"lte_seconds,omitempty"`
	GtSeconds  *float64 `json:"gt_seconds,omitempty"`
	GteSeconds *float64 `json:"gte_seconds,omitempty"`

	LtCharges  *int `json:"lt_charges,omitempty"`
	LteCharges *int `json:"lte_charges,omitempty"`
	GtCharges  *int `json:"gt_charges,omitempty"`
	GteCharges *int `json:"gte_charges,omitempty"`
}

func main() {
	configDir := flag.String("config-dir", "./configs", "Path to config directory")
	addr := flag.String("addr", ":8080", "Listen address (e.g., :8080)")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := content.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "missing index.html", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	http.HandleFunc("/api/player", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetPlayer(w, r, *configDir)
		case http.MethodPost:
			handleSavePlayer(w, r, *configDir)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/identifiers", handleIdentifiers)
	http.HandleFunc("/api/rotations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleListRotations(w, r, *configDir)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	http.HandleFunc("/api/rotations/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/rotations/")
		if name == "" {
			http.Error(w, "rotation name required", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			handleGetRotation(w, r, *configDir, name)
		case http.MethodPost:
			handleSaveRotation(w, r, *configDir, name)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	listenAddr := normalizeAddr(*addr)
	log.Printf("UI server listening on %s (config dir: %s)\n", listenAddr, *configDir)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

// normalizeAddr ensures the listen address includes a colon.
func normalizeAddr(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return addr
	}
	return ":" + addr
}

func handleGetPlayer(w http.ResponseWriter, r *http.Request, configDir string) {
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("load config: %v", err), http.StatusInternalServerError)
		return
	}

	resp := playerResponse{
		Player: cfg.Player,
	}
	resp.Options.Rotations, err = listRotationFiles(filepath.Join(configDir, "rotations"))
	if err != nil {
		http.Error(w, fmt.Sprintf("list rotations: %v", err), http.StatusInternalServerError)
		return
	}
	resp.Options.Pets = []string{"imp", "none"}
	resp.Options.Runes = groupRunesByRarity()
	resp.Options.Limits = cfg.Player.MysticEnchants

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleSavePlayer(w http.ResponseWriter, r *http.Request, configDir string) {
	var req saveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := validatePlayerPayload(req.Player, configDir); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	out, err := yaml.Marshal(&req.Player)
	if err != nil {
		http.Error(w, fmt.Sprintf("marshal yaml: %v", err), http.StatusInternalServerError)
		return
	}

	dest := filepath.Join(configDir, "player.yaml")
	if err := os.WriteFile(dest, out, 0644); err != nil {
		http.Error(w, fmt.Sprintf("write player.yaml: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func validatePlayerPayload(p config.Player, configDir string) error {
	allowedPets := []string{"imp", "none"}
	if p.Pet.Summon != "" && !slices.Contains(allowedPets, strings.ToLower(p.Pet.Summon)) {
		return fmt.Errorf("invalid pet: %s", p.Pet.Summon)
	}
	p.Pet.Summon = strings.ToLower(p.Pet.Summon)

	rotations, err := listRotationFiles(filepath.Join(configDir, "rotations"))
	if err != nil {
		return fmt.Errorf("list rotations: %w", err)
	}
	if p.Rotation != "" && !slices.Contains(rotations, p.Rotation) {
		return fmt.Errorf("rotation not found: %s", p.Rotation)
	}

	cfg := config.Config{Player: p}
	if err := cfg.Player.Validate(); err != nil {
		return err
	}
	return nil
}

func listRotationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			files = append(files, name)
		}
	}
	slices.Sort(files)
	return files, nil
}

func groupRunesByRarity() map[string][]string {
	out := map[string][]string{
		string(runes.RarityLegendary): {},
		string(runes.RarityEpic):      {},
		string(runes.RarityRare):      {},
	}
	for name := range runes.KnownRunes() {
		rarity, _ := runes.RarityOf(name)
		out[string(rarity)] = append(out[string(rarity)], name)
	}
	for key := range out {
		slices.Sort(out[key])
	}
	return out
}

func handleIdentifiers(w http.ResponseWriter, r *http.Request) {
	resp := identifiersResponse{
		Spells:    collectKeys(apl.KnownSpells()),
		Buffs:     collectKeys(apl.KnownBuffs()),
		Debuffs:   collectKeys(apl.KnownDebuffs()),
		Resources: collectKeys(apl.KnownResources()),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func collectKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func handleListRotations(w http.ResponseWriter, r *http.Request, configDir string) {
	files, err := listRotationFiles(filepath.Join(configDir, "rotations"))
	if err != nil {
		http.Error(w, fmt.Sprintf("list rotations: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func handleGetRotation(w http.ResponseWriter, r *http.Request, configDir, name string) {
	path := filepath.Join(configDir, "rotations", name)
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("read rotation: %v", err), http.StatusInternalServerError)
		return
	}
	var file apl.File
	if err := yaml.Unmarshal(data, &file); err != nil {
		http.Error(w, fmt.Sprintf("parse rotation: %v", err), http.StatusBadRequest)
		return
	}
	dto, err := rotationFileToDTO(&file)
	if err != nil {
		http.Error(w, fmt.Sprintf("convert rotation: %v", err), http.StatusInternalServerError)
		return
	}
	resp := rotationResponse{
		Filename: name,
		File:     *dto,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleSaveRotation(w http.ResponseWriter, r *http.Request, configDir, name string) {
	var dto rotationDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	file, err := dtoToRotationFile(&dto)
	if err != nil {
		http.Error(w, fmt.Sprintf("build rotation: %v", err), http.StatusBadRequest)
		return
	}
	out, err := yaml.Marshal(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("marshal yaml: %v", err), http.StatusInternalServerError)
		return
	}
	if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
		name += ".yaml"
	}
	dest := filepath.Join(configDir, "rotations", name)
	if err := os.WriteFile(dest, out, 0644); err != nil {
		http.Error(w, fmt.Sprintf("write rotation: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func rotationFileToDTO(f *apl.File) (*rotationDTO, error) {
	dto := &rotationDTO{
		Name:        f.Name,
		Description: f.Description,
		Imports:     f.Imports,
		Variables:   f.Variables,
	}
	for _, act := range f.Rotation {
		a, err := actionToDTO(act)
		if err != nil {
			return nil, err
		}
		dto.Rotation = append(dto.Rotation, *a)
	}
	return dto, nil
}

func actionToDTO(a apl.ActionDefinition) (*actionDTO, error) {
	dto := &actionDTO{
		Action:          a.Action,
		Spell:           a.Spell,
		Item:            a.Item,
		DurationSeconds: a.DurationSeconds,
		Tags:            a.Tags,
	}
	if a.When != nil {
		c, err := conditionToDTO(a.When.Node())
		if err != nil {
			return nil, err
		}
		dto.When = c
	}
	for _, step := range a.Steps {
		child, err := actionToDTO(step)
		if err != nil {
			return nil, err
		}
		dto.Steps = append(dto.Steps, *child)
	}
	return dto, nil
}

func conditionToDTO(node *yaml.Node) (*conditionDTO, error) {
	if node == nil {
		return nil, nil
	}
	if node.Kind == yaml.ScalarNode {
		switch strings.ToLower(node.Value) {
		case "true":
			return &conditionDTO{Type: "true"}, nil
		case "false":
			return &conditionDTO{Type: "false"}, nil
		}
	}
	if node.Kind != yaml.MappingNode || len(node.Content) == 0 {
		return nil, fmt.Errorf("unsupported condition node")
	}
	key := node.Content[0].Value
	switch key {
	case "all", "any":
		childrenNode := node.Content[1]
		if childrenNode.Kind != yaml.SequenceNode {
			return nil, fmt.Errorf("%s expects sequence", key)
		}
		dto := &conditionDTO{Type: key}
		for _, child := range childrenNode.Content {
			c, err := conditionToDTO(child)
			if err != nil {
				return nil, err
			}
			dto.Children = append(dto.Children, *c)
		}
		return dto, nil
	case "not":
		child, err := conditionToDTO(node.Content[1])
		if err != nil {
			return nil, err
		}
		return &conditionDTO{Type: "not", Children: []conditionDTO{*child}}, nil
	case "buff_active":
		m := mapNodeToMap(node.Content[1])
		dto := &conditionDTO{Type: "buff_active", Buff: m["buff"]}
		dto.MinRemainingSeconds = parseOptFloat(m, "min_remaining")
		dto.MaxRemainingSeconds = parseOptFloat(m, "max_remaining")
		return dto, nil
	case "debuff_active":
		m := mapNodeToMap(node.Content[1])
		dto := &conditionDTO{Type: "debuff_active", Debuff: m["debuff"]}
		dto.MinRemainingSeconds = parseOptFloat(m, "min_remaining")
		dto.MaxRemainingSeconds = parseOptFloat(m, "max_remaining")
		return dto, nil
	case "dot_remaining":
		m := mapNodeToMap(node.Content[1])
		dto := &conditionDTO{Type: "dot_remaining", Spell: m["spell"]}
		dto.LtSeconds = parseOptFloat(m, "lt_seconds")
		dto.LteSeconds = parseOptFloat(m, "lte_seconds")
		dto.GtSeconds = parseOptFloat(m, "gt_seconds")
		dto.GteSeconds = parseOptFloat(m, "gte_seconds")
		return dto, nil
	case "cooldown_ready":
		m := mapNodeToMap(node.Content[1])
		dto := &conditionDTO{Type: "cooldown_ready", Spell: m["spell"]}
		return dto, nil
	case "cooldown_remaining":
		m := mapNodeToMap(node.Content[1])
		dto := &conditionDTO{Type: "cooldown_remaining", Spell: m["spell"]}
		dto.LtSeconds = parseOptFloat(m, "lt_seconds")
		dto.LteSeconds = parseOptFloat(m, "lte_seconds")
		dto.GtSeconds = parseOptFloat(m, "gt_seconds")
		dto.GteSeconds = parseOptFloat(m, "gte_seconds")
		return dto, nil
	case "resource_percent":
		m := mapNodeToMap(node.Content[1])
		dto := &conditionDTO{Type: "resource_percent", Resource: m["resource"]}
		dto.LtSeconds = parseOptFloat(m, "lt")
		dto.LteSeconds = parseOptFloat(m, "lte")
		dto.GtSeconds = parseOptFloat(m, "gt")
		dto.GteSeconds = parseOptFloat(m, "gte")
		return dto, nil
	case "charges":
		m := mapNodeToMap(node.Content[1])
		dto := &conditionDTO{Type: "charges", Buff: m["buff"]}
		dto.LtCharges = parseOptInt(m, "lt")
		dto.LteCharges = parseOptInt(m, "lte")
		dto.GtCharges = parseOptInt(m, "gt")
		dto.GteCharges = parseOptInt(m, "gte")
		return dto, nil
	default:
		return nil, fmt.Errorf("unsupported condition key: %s", key)
	}
}

func parseOptFloat(m map[string]string, key string) *float64 {
	if v, ok := m[key]; ok {
		val, err := time.ParseDuration(v + "s")
		if err == nil {
			f := val.Seconds()
			return &f
		}
		if f, err2 := strconv.ParseFloat(v, 64); err2 == nil {
			return &f
		}
	}
	return nil
}

func parseOptInt(m map[string]string, key string) *int {
	if v, ok := m[key]; ok {
		if iv, err := strconv.Atoi(v); err == nil {
			return &iv
		}
	}
	return nil
}

func mapNodeToMap(n *yaml.Node) map[string]string {
	out := map[string]string{}
	if n == nil || n.Kind != yaml.MappingNode {
		return out
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		k := n.Content[i].Value
		v := n.Content[i+1].Value
		out[k] = v
	}
	return out
}

func dtoToRotationFile(dto *rotationDTO) (*apl.File, error) {
	file := &apl.File{
		Name:        dto.Name,
		Description: dto.Description,
		Imports:     dto.Imports,
		Variables:   dto.Variables,
	}
	for _, a := range dto.Rotation {
		act, err := dtoToAction(a)
		if err != nil {
			return nil, err
		}
		file.Rotation = append(file.Rotation, *act)
	}
	return file, nil
}

func dtoToAction(a actionDTO) (*apl.ActionDefinition, error) {
	act := &apl.ActionDefinition{
		Action:          a.Action,
		Spell:           a.Spell,
		Item:            a.Item,
		DurationSeconds: a.DurationSeconds,
		Tags:            a.Tags,
	}
	if a.When != nil {
		node, err := conditionDTOToNode(a.When)
		if err != nil {
			return nil, err
		}
		act.When = apl.NewConditionNode(node)
	}
	for _, step := range a.Steps {
		child, err := dtoToAction(step)
		if err != nil {
			return nil, err
		}
		act.Steps = append(act.Steps, *child)
	}
	return act, nil
}

func conditionDTOToNode(c *conditionDTO) (*yaml.Node, error) {
	if c == nil {
		return nil, nil
	}
	switch c.Type {
	case "true":
		return &yaml.Node{Kind: yaml.ScalarNode, Value: "true"}, nil
	case "false":
		return &yaml.Node{Kind: yaml.ScalarNode, Value: "false"}, nil
	case "all", "any":
		seq := &yaml.Node{Kind: yaml.SequenceNode}
		for _, child := range c.Children {
			n, err := conditionDTOToNode(&child)
			if err != nil {
				return nil, err
			}
			seq.Content = append(seq.Content, n)
		}
		return mapToNode(c.Type, seq), nil
	case "not":
		if len(c.Children) == 0 {
			return mapToNode("not", &yaml.Node{Kind: yaml.ScalarNode, Value: "true"}), nil
		}
		n, err := conditionDTOToNode(&c.Children[0])
		if err != nil {
			return nil, err
		}
		return mapToNode("not", n), nil
	case "buff_active":
		m := map[string]any{"buff": c.Buff}
		if c.MinRemainingSeconds != nil {
			m["min_remaining"] = *c.MinRemainingSeconds
		}
		if c.MaxRemainingSeconds != nil {
			m["max_remaining"] = *c.MaxRemainingSeconds
		}
		return mapToNode("buff_active", mapAnyToNode(m)), nil
	case "debuff_active":
		m := map[string]any{"debuff": c.Debuff}
		if c.MinRemainingSeconds != nil {
			m["min_remaining"] = *c.MinRemainingSeconds
		}
		if c.MaxRemainingSeconds != nil {
			m["max_remaining"] = *c.MaxRemainingSeconds
		}
		return mapToNode("debuff_active", mapAnyToNode(m)), nil
	case "dot_remaining":
		m := map[string]any{"spell": c.Spell}
		addComparators(m, c)
		return mapToNode("dot_remaining", mapAnyToNode(m)), nil
	case "cooldown_ready":
		return mapToNode("cooldown_ready", mapAnyToNode(map[string]any{"spell": c.Spell})), nil
	case "cooldown_remaining":
		m := map[string]any{"spell": c.Spell}
		addComparators(m, c)
		return mapToNode("cooldown_remaining", mapAnyToNode(m)), nil
	case "resource_percent":
		m := map[string]any{"resource": c.Resource}
		addComparators(m, c)
		return mapToNode("resource_percent", mapAnyToNode(m)), nil
	case "charges":
		m := map[string]any{"buff": c.Buff}
		if c.LtCharges != nil {
			m["lt"] = *c.LtCharges
		}
		if c.LteCharges != nil {
			m["lte"] = *c.LteCharges
		}
		if c.GtCharges != nil {
			m["gt"] = *c.GtCharges
		}
		if c.GteCharges != nil {
			m["gte"] = *c.GteCharges
		}
		return mapToNode("charges", mapAnyToNode(m)), nil
	default:
		return nil, fmt.Errorf("unsupported condition type %s", c.Type)
	}
}

func addComparators(m map[string]any, c *conditionDTO) {
	if c.LtSeconds != nil {
		m["lt_seconds"] = *c.LtSeconds
	}
	if c.LteSeconds != nil {
		m["lte_seconds"] = *c.LteSeconds
	}
	if c.GtSeconds != nil {
		m["gt_seconds"] = *c.GtSeconds
	}
	if c.GteSeconds != nil {
		m["gte_seconds"] = *c.GteSeconds
	}
}

func mapAnyToNode(m map[string]any) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode}
	for k, v := range m {
		node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: k})
		node.Content = append(node.Content, scalarFor(v))
	}
	return node
}

func scalarFor(v any) *yaml.Node {
	switch t := v.(type) {
	case string:
		return &yaml.Node{Kind: yaml.ScalarNode, Value: t}
	case int:
		return &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", t)}
	case float64:
		return &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%v", t)}
	default:
		return &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%v", t)}
	}
}

func mapToNode(key string, value *yaml.Node) *yaml.Node {
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: key},
			value,
		},
	}
}
