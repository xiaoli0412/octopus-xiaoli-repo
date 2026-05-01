package llmname

import (
	"regexp"
	"strings"
)

var normalizeSeparators = regexp.MustCompile(`[^a-z0-9]+`)
var versionLike = regexp.MustCompile(`^[0-9]+(-[0-9]+)*$`)

var knownVariants = map[string]struct{}{
	"pro":       {},
	"mini":      {},
	"nano":      {},
	"turbo":     {},
	"preview":   {},
	"latest":    {},
	"flash":     {},
	"lite":      {},
	"reasoning": {},
	"thinking":  {},
	"instruct":  {},
	"chat":      {},
}

var manualAliases = map[string]string{
	"gpt54":          "gpt-5.4",
	"gpt54pro":       "gpt-5.4-pro",
	"gpt54mini":      "gpt-5.4-mini",
	"glm45":          "glm-4.5",
	"glm45pro":       "glm-4.5-pro",
	"deepseekv32":    "deepseek-v3.2",
	"deepseekv321":   "deepseek-v3.21",
	"claude35sonnet": "claude-3.5-sonnet",
	"claude37sonnet": "claude-3.7-sonnet",
	"gemini25pro":    "gemini-2.5-pro",
	"gemini25flash":  "gemini-2.5-flash",
	"qwenmax":        "qwen-max",
}

type NormalizedModelInfo struct {
	Normalized string
	Compact    string
	Family     string
	Version    string
	Variant    string
	Parts      []string
}

func NormalizeModelName(input string) string {
	v := strings.ToLower(strings.TrimSpace(input))
	if v == "" {
		return ""
	}
	v = strings.ReplaceAll(v, "_", "-")
	v = strings.ReplaceAll(v, " ", "-")
	v = strings.ReplaceAll(v, ".", "-")
	v = normalizeSeparators.ReplaceAllString(v, "-")
	v = strings.Trim(v, "-")
	for strings.Contains(v, "--") {
		v = strings.ReplaceAll(v, "--", "-")
	}
	return v
}

func compactModelName(input string) string {
	v := NormalizeModelName(input)
	v = strings.ReplaceAll(v, "-", "")
	return v
}

func ParseNormalizedModelInfo(input string) NormalizedModelInfo {
	norm := NormalizeModelName(input)
	parts := strings.Split(norm, "-")
	info := NormalizedModelInfo{Normalized: norm, Compact: compactModelName(input), Parts: parts}
	if len(parts) == 0 {
		return info
	}
	info.Family = parts[0]
	for _, part := range parts[1:] {
		switch {
		case info.Version == "" && versionLike.MatchString(part):
			info.Version = part
		case info.Variant == "":
			info.Variant = part
		}
	}
	for _, part := range parts[1:] {
		if _, ok := knownVariants[part]; ok {
			info.Variant = part
			break
		}
	}
	return info
}

func removeKnownSuffixes(input string) string {
	parts := strings.Split(input, "-")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if _, ok := knownVariants[part]; ok {
			continue
		}
		filtered = append(filtered, part)
	}
	return strings.Join(filtered, "-")
}

func CandidateModelKeys(input string) []string {
	info := ParseNormalizedModelInfo(input)
	if info.Normalized == "" {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, 12)
	add := func(v string) {
		if v == "" {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	add(info.Normalized)
	if alias, ok := manualAliases[info.Compact]; ok {
		add(alias)
	}
	if info.Family != "" && info.Version != "" && info.Variant != "" {
		add(strings.Join([]string{info.Family, info.Version, info.Variant}, "-"))
		add(strings.Join([]string{info.Family, info.Variant, info.Version}, "-"))
	}
	if info.Family != "" && info.Version != "" {
		add(strings.Join([]string{info.Family, info.Version}, "-"))
	}
	if info.Family != "" && info.Variant != "" {
		add(strings.Join([]string{info.Family, info.Variant}, "-"))
	}
	add(strings.ReplaceAll(info.Normalized, "-preview", ""))
	add(strings.ReplaceAll(info.Normalized, "-latest", ""))
	add(strings.ReplaceAll(info.Normalized, "-turbo", ""))
	add(strings.ReplaceAll(info.Normalized, "-reasoning", ""))
	add(strings.ReplaceAll(info.Normalized, "-thinking", ""))
	add(removeKnownSuffixes(info.Normalized))
	add(strings.ReplaceAll(info.Normalized, "-preview", "-"))
	add(strings.ReplaceAll(info.Normalized, "-latest", "-"))
	add(info.Family)
	return out
}

func CanonicalModelName(input string) string {
	info := ParseNormalizedModelInfo(input)
	if info.Normalized == "" {
		return ""
	}
	if alias, ok := manualAliases[info.Compact]; ok {
		return alias
	}
	return info.Normalized
}
