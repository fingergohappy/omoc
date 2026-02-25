package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const configPath = ".config/opencode/oh-my-opencode.json"

// AgentEntry represents an agent or category model configuration.
type AgentEntry struct {
	Model   string `json:"model,omitempty"`
	Variant string `json:"variant,omitempty"`
}

// Config represents the oh-my-opencode.json structure.
// We use a raw map to preserve unknown fields on save.
type Config struct {
	raw        map[string]json.RawMessage
	Agents     map[string]*AgentEntry
	Categories map[string]*AgentEntry
}

// KnownAgents is the ordered list of agent names from the schema.
var KnownAgents = []string{
	"sisyphus",
	"prometheus",
	"metis",
	"atlas",
	"oracle",
	"momus",
	"librarian",
	"explore",
	"multimodal-looker",
}

// KnownCategories is the ordered list of task categories.
var KnownCategories = []string{
	"visual-engineering",
	"ultrabrain",
	"deep",
	"artistry",
	"quick",
	"unspecified-high",
	"unspecified-low",
	"writing",
}

// KnownVariants lists the valid variant values.
var KnownVariants = []string{"", "low", "high", "xhigh", "max"}

func configFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, configPath)
}

// Load reads and parses the config file.
func Load() (*Config, error) {
	data, err := os.ReadFile(configFilePath())
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	c := &Config{
		raw:        raw,
		Agents:     make(map[string]*AgentEntry),
		Categories: make(map[string]*AgentEntry),
	}

	if agentsRaw, ok := raw["agents"]; ok {
		var agents map[string]*AgentEntry
		if err := json.Unmarshal(agentsRaw, &agents); err != nil {
			return nil, fmt.Errorf("parse agents: %w", err)
		}
		c.Agents = agents
	}

	if catsRaw, ok := raw["categories"]; ok {
		var cats map[string]*AgentEntry
		if err := json.Unmarshal(catsRaw, &cats); err != nil {
			return nil, fmt.Errorf("parse categories: %w", err)
		}
		c.Categories = cats
	}

	for _, name := range KnownAgents {
		if c.Agents[name] == nil {
			c.Agents[name] = &AgentEntry{}
		}
	}
	for _, name := range KnownCategories {
		if c.Categories[name] == nil {
			c.Categories[name] = &AgentEntry{}
		}
	}

	return c, nil
}

// Save creates a backup then writes the config back to disk, preserving unknown fields.
func (c *Config) Save() error {
	cfgPath := configFilePath()
	if err := backupFile(cfgPath); err != nil {
		return fmt.Errorf("backup: %w", err)
	}

	agentsJSON, err := json.Marshal(c.Agents)
	if err != nil {
		return fmt.Errorf("marshal agents: %w", err)
	}
	c.raw["agents"] = agentsJSON

	catsJSON, err := json.Marshal(c.Categories)
	if err != nil {
		return fmt.Errorf("marshal categories: %w", err)
	}
	c.raw["categories"] = catsJSON

	data, err := json.MarshalIndent(c.raw, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')

	return os.WriteFile(configFilePath(), data, 0644)
}

func backupFile(path string) error {
	src, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer src.Close()

	backupPath := path + "." + time.Now().Format("20060102-150405") + ".bak"
	dst, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
