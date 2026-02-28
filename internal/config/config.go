package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	ActiveProfileLinkFile = "oh-my-opencode.json"
	DefaultProfileFile    = "oh-my-opencode.default.json"
	TestProfileFile       = "oh-my-opencode.test.json"

	profileFilePrefix = "oh-my-opencode."
	profileFileSuffix = ".json"
)

// AgentEntry 表示 agent 或 category 的模型配置。
type AgentEntry struct {
	Model   string `json:"model,omitempty"`
	Variant string `json:"variant,omitempty"`
}

// Config 表示 oh-my-opencode.json 的结构。
// 使用 raw map 以便在保存时保留未知字段。
type Config struct {
	raw                map[string]json.RawMessage
	Agents             map[string]*AgentEntry
	Categories         map[string]*AgentEntry
	ActiveProfileFile  string
	ProfileLoadWarning string
}

// KnownAgents 是 schema 中按顺序定义的 agent 名称列表。
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

// KnownCategories 是任务分类的有序列表。
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

// KnownVariants 列出所有合法的 variant 值。
var KnownVariants = []string{"", "low", "high", "xhigh", "max"}

func configFilePath() string {
	return configFilePathFor(ActiveProfileLinkFile)
}

func configDirPath() string {
	xdgConfigHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME"))
	if xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "opencode")
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "opencode")
}

func configFilePathFor(profileFile string) string {
	return filepath.Join(configDirPath(), profileFile)
}

func profilePathFor(profileFile string) (string, error) {
	if !isValidProfileFileName(profileFile) {
		return "", fmt.Errorf("invalid profile filename: %q", profileFile)
	}

	configDir := configDirPath()
	profilePath := filepath.Clean(filepath.Join(configDir, profileFile))
	rel, err := filepath.Rel(configDir, profilePath)
	if err != nil {
		return "", fmt.Errorf("resolve profile path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("profile path escapes config dir: %q", profileFile)
	}

	return profilePath, nil
}

func activeProfileLinkPath() string {
	return filepath.Join(configDirPath(), ActiveProfileLinkFile)
}

func isValidProfileFileName(name string) bool {
	if filepath.Base(name) != name {
		return false
	}

	if !strings.HasPrefix(name, profileFilePrefix) || !strings.HasSuffix(name, profileFileSuffix) {
		return false
	}

	profileName := strings.TrimSuffix(strings.TrimPrefix(name, profileFilePrefix), profileFileSuffix)
	if profileName == "" {
		return false
	}

	for _, r := range profileName {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return false
	}

	return true
}

func profileFileNameForName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", fmt.Errorf("profile name is required")
	}

	fileName := profileFilePrefix + trimmed + profileFileSuffix
	if !isValidProfileFileName(fileName) {
		return "", fmt.Errorf("invalid profile name %q: allowed [a-z0-9_-]", trimmed)
	}

	return fileName, nil
}

// ListProfiles 以确定性顺序返回已发现的 profile 文件名。
func ListProfiles() ([]string, error) {
	entries, err := os.ReadDir(configDirPath())
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("read config dir: %w", err)
	}

	hasDefault := false
	hasTest := false
	others := make([]string, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if name == ActiveProfileLinkFile {
			continue
		}
		if !isValidProfileFileName(name) {
			continue
		}

		switch name {
		case DefaultProfileFile:
			hasDefault = true
		case TestProfileFile:
			hasTest = true
		default:
			others = append(others, name)
		}
	}

	sort.Strings(others)

	profiles := make([]string, 0, 2+len(others))
	if hasDefault {
		profiles = append(profiles, DefaultProfileFile)
	}
	profiles = append(profiles, others...)
	if hasTest {
		profiles = append(profiles, TestProfileFile)
	}

	return profiles, nil
}

// LoadProfile 读取并解析指定的 profile 文件。
func LoadProfile(filename string) (*Config, error) {
	path, err := profilePathFor(filename)
	if err != nil {
		return nil, err
	}
	cfg, err := loadFromPath(path)
	if err != nil {
		return nil, err
	}
	cfg.ActiveProfileFile = filename
	return cfg, nil
}

// Load 读取并解析配置文件。
func Load() (*Config, error) {
	warnings := make([]string, 0, 2)

	activeProfileFile, err := resolveActiveProfileFile()
	if err != nil {
		warnings = append(warnings, err.Error())
	}

	if activeProfileFile != "" {
		activePath, pathErr := profilePathFor(activeProfileFile)
		if pathErr != nil {
			warnings = append(warnings, fmt.Sprintf("active profile %q could not be resolved (%v)", activeProfileFile, pathErr))
		} else {
			cfg, loadErr := loadFromPath(activePath)
			if loadErr == nil {
				cfg.ActiveProfileFile = activeProfileFile
				cfg.ProfileLoadWarning = strings.Join(warnings, "; ")
				return cfg, nil
			}
			warnings = append(warnings, fmt.Sprintf("active profile %q could not be loaded (%v)", activeProfileFile, loadErr))
		}
	}

	defaultPath, defaultPathErr := profilePathFor(DefaultProfileFile)
	if defaultPathErr == nil {
		defaultCfg, defaultErr := loadFromPath(defaultPath)
		if defaultErr == nil {
			defaultCfg.ActiveProfileFile = DefaultProfileFile
			if linkErr := writeActiveProfileLink(activeProfileLinkPath(), DefaultProfileFile); linkErr != nil {
				warnings = append(warnings, fmt.Sprintf("failed to update active profile link: %v", linkErr))
			}
			defaultCfg.ProfileLoadWarning = strings.Join(warnings, "; ")
			return defaultCfg, nil
		}
	}

	cfg := defaultConfig()
	cfg.ActiveProfileFile = DefaultProfileFile
	cfg.ProfileLoadWarning = strings.Join(warnings, "; ")
	return cfg, nil
}

func loadFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
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

	ensureKnownEntries(c)

	return c, nil
}

func defaultConfig() *Config {
	c := &Config{
		raw:        make(map[string]json.RawMessage),
		Agents:     make(map[string]*AgentEntry),
		Categories: make(map[string]*AgentEntry),
	}
	ensureKnownEntries(c)
	return c
}

func (c *Config) clone() *Config {
	clonedRaw := make(map[string]json.RawMessage, len(c.raw))
	for key, value := range c.raw {
		copyBytes := make(json.RawMessage, len(value))
		copy(copyBytes, value)
		clonedRaw[key] = copyBytes
	}

	clonedAgents := make(map[string]*AgentEntry, len(c.Agents))
	for key, value := range c.Agents {
		if value == nil {
			clonedAgents[key] = &AgentEntry{}
			continue
		}
		copyEntry := *value
		clonedAgents[key] = &copyEntry
	}

	clonedCategories := make(map[string]*AgentEntry, len(c.Categories))
	for key, value := range c.Categories {
		if value == nil {
			clonedCategories[key] = &AgentEntry{}
			continue
		}
		copyEntry := *value
		clonedCategories[key] = &copyEntry
	}

	cloned := &Config{
		raw:                clonedRaw,
		Agents:             clonedAgents,
		Categories:         clonedCategories,
		ActiveProfileFile:  c.ActiveProfileFile,
		ProfileLoadWarning: c.ProfileLoadWarning,
	}
	ensureKnownEntries(cloned)
	return cloned
}

func (c *Config) activeProfileFileOrLegacy() string {
	activeProfileFile := strings.TrimSpace(c.ActiveProfileFile)
	if activeProfileFile == "" {
		return DefaultProfileFile
	}
	return activeProfileFile
}

// ActiveProfilePath 返回当前激活 profile 文件的绝对路径。
func (c *Config) ActiveProfilePath() (string, error) {
	return profilePathFor(c.activeProfileFileOrLegacy())
}

// CloneToNewProfile 通过克隆当前配置创建新 profile 文件，并将其设为激活状态。
func (c *Config) CloneToNewProfile(profileName string) (string, error) {
	newProfileFile, err := profileFileNameForName(profileName)
	if err != nil {
		return "", err
	}

	newProfilePath, err := profilePathFor(newProfileFile)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(newProfilePath); err == nil {
		return "", fmt.Errorf("profile %q already exists", newProfileFile)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("check profile file: %w", err)
	}

	clone := c.clone()
	clone.ActiveProfileFile = newProfileFile
	if err := clone.Save(); err != nil {
		return "", err
	}

	*c = *clone
	return newProfileFile, nil
}

func ensureKnownEntries(c *Config) {
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
}

func readActiveProfileLink(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return "", fmt.Errorf("active profile link is not a symlink")
	}

	target, err := os.Readlink(path)
	if err != nil {
		return "", fmt.Errorf("read active profile link: %w", err)
	}
	target = strings.TrimSpace(target)
	if target == "" {
		return "", fmt.Errorf("active profile link target is empty")
	}
	if filepath.Base(target) != target || filepath.IsAbs(target) {
		return "", fmt.Errorf("active profile link target must be a basename profile filename: %q", target)
	}
	if !isValidProfileFileName(target) {
		return "", fmt.Errorf("invalid active profile filename: %q", target)
	}

	return target, nil
}

func writeActiveProfileLink(path, activeProfileFile string) error {
	if !isValidProfileFileName(activeProfileFile) {
		return fmt.Errorf("invalid active profile filename: %q", activeProfileFile)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if filepath.Base(activeProfileFile) != activeProfileFile {
		return fmt.Errorf("active profile filename must be basename: %q", activeProfileFile)
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("replace active profile link: %w", err)
	}

	if err := os.Symlink(activeProfileFile, path); err != nil {
		return fmt.Errorf("create active profile symlink %q -> %q: %w", filepath.Base(path), activeProfileFile, err)
	}

	return nil
}

func resolveActiveProfileFile() (string, error) {
	linkPath := activeProfileLinkPath()
	linkInfo, err := os.Lstat(linkPath)
	if err == nil {
		if linkInfo.Mode()&os.ModeSymlink != 0 {
			activeProfileFile, readErr := readActiveProfileLink(linkPath)
			if readErr != nil {
				return "", fmt.Errorf("active profile link is invalid (%v)", readErr)
			}
			return activeProfileFile, nil
		}
		if !linkInfo.Mode().IsRegular() {
			return "", fmt.Errorf("active profile link path exists but is not a regular file or symlink")
		}

		defaultPath, pathErr := profilePathFor(DefaultProfileFile)
		if pathErr != nil {
			return "", fmt.Errorf("default profile path could not be resolved (%v)", pathErr)
		}
		if renameErr := os.Rename(linkPath, defaultPath); renameErr != nil {
			return "", fmt.Errorf("legacy config migration failed (rename %q -> %q): %v", ActiveProfileLinkFile, DefaultProfileFile, renameErr)
		}
		if writeErr := writeActiveProfileLink(linkPath, DefaultProfileFile); writeErr != nil {
			return DefaultProfileFile, fmt.Errorf("legacy config migrated but active profile link update failed (%v)", writeErr)
		}
		return DefaultProfileFile, nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("active profile link could not be inspected (%v)", err)
	}

	defaultPath, pathErr := profilePathFor(DefaultProfileFile)
	if pathErr != nil {
		return "", fmt.Errorf("default profile path could not be resolved (%v)", pathErr)
	}
	if _, statErr := os.Stat(defaultPath); statErr == nil {
		if writeErr := writeActiveProfileLink(linkPath, DefaultProfileFile); writeErr != nil {
			return DefaultProfileFile, fmt.Errorf("active profile link was missing and could not be recreated (%v)", writeErr)
		}
		return DefaultProfileFile, nil
	} else if !os.IsNotExist(statErr) {
		return "", fmt.Errorf("default profile could not be inspected (%v)", statErr)
	}

	return "", nil
}

// Save 先创建备份，再将配置写回磁盘，同时保留未知字段。
func (c *Config) Save() error {
	activeProfileFile := c.activeProfileFileOrLegacy()
	if !isValidProfileFileName(activeProfileFile) {
		return fmt.Errorf("invalid active profile filename: %q", activeProfileFile)
	}

	cfgPath, err := profilePathFor(activeProfileFile)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		return fmt.Errorf("mkdir config dir: %w", err)
	}

	if c.raw == nil {
		c.raw = make(map[string]json.RawMessage)
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

	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		return err
	}

	if err := writeActiveProfileLink(activeProfileLinkPath(), activeProfileFile); err != nil {
		return fmt.Errorf("write active profile link: %w", err)
	}

	c.ActiveProfileFile = activeProfileFile
	return nil
}
