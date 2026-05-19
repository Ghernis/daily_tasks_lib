package store

import (
	"errors"
	"strconv"
	"strings"
)

var ErrEmptyHours = errors.New("empty hours")

// MergeSubtaskRows builds ordered checkbox rows: defaults first (checked), then past-only names.
func MergeSubtaskRows(defaults, fromPast []string) (ordered []string, defaultSet map[string]bool) {
	defaultSet = make(map[string]bool)
	seen := make(map[string]bool)
	for _, d := range defaults {
		if d == "" || seen[d] {
			continue
		}
		seen[d] = true
		defaultSet[d] = true
		ordered = append(ordered, d)
	}
	for _, p := range fromPast {
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		ordered = append(ordered, p)
	}
	return ordered, defaultSet
}

func ParseHours(raw string) (float64, error) {
	text := strings.TrimSpace(strings.ReplaceAll(raw, ",", "."))
	if text == "" {
		return 0, ErrEmptyHours
	}
	return strconv.ParseFloat(text, 64)
}

func ParseSubtaskLines(raw string) []string {
	var parts []string
	for _, line := range strings.Split(strings.ReplaceAll(raw, ",", "\n"), "\n") {
		t := strings.TrimSpace(line)
		if t != "" {
			parts = append(parts, t)
		}
	}
	return parts
}
