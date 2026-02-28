package models

import (
	"os/exec"
	"sort"
	"strings"
)

// Fetch 执行 `opencode models` 并返回排序后的可用模型列表。
func Fetch() ([]string, error) {
	out, err := exec.Command("opencode", "models").Output()
	if err != nil {
		return nil, err
	}

	var result []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	sort.Strings(result)
	return result, nil
}
