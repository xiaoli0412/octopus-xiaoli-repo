package llmname

import (
	"regexp"
	"strings"
)

var normalizeSeparators = regexp.MustCompile(`[^a-z0-9]+`)
var numericToken = regexp.MustCompile(`^[0-9]+$`)
var versionLike = regexp.MustCompile(`^v?[0-9]+([.-][0-9]+)*[a-z]*$`)
var compactVersionLike = regexp.MustCompile(`^[0-9]+[a-z]+$`)
var datestampLike = regexp.MustCompile(`^[0-9]{6,}$`)

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
	"max":       {},
	"ultra":     {},
	"plus":      {},
	"exp":       {},
	"beta":      {},
	"alpha":     {},
	"vision":    {},
	"audio":     {},
	"image":     {},
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

var vendorAliases = map[string]string{
	"gpt":        "openai",
	"text":       "openai",
	"o1":         "openai",
	"o3":         "openai",
	"o4":         "openai",
	"codex":      "openai",
	"claude":     "anthropic",
	"gemini":     "google",
	"deepseek":   "deepseek",
	"glm":        "glm",
	"zhipu":      "glm",
	"qwen":       "qwen",
	"kimi":       "moonshot",
	"moonshot":   "moonshot",
	"moonshotai": "moonshot",
	"grok":       "xai",
	"xai":        "xai",
	"minimax":    "minimax",
	"doubao":     "doubao",
	"hunyuan":    "hunyuan",
	"llama":      "llama",
	"mistral":    "mistral",
	"yi":         "yi",
	"amazon":     "amazon",
	"nova":       "amazon",
	"bedrock":    "amazon",
}

type NormalizedModelInfo struct {
	Normalized string
	Compact    string
	Vendor     string
	Family     string
	Version    string
	Variant    string
	Variants   []string
	Suffix     string
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

func detectFamily(parts []string) (string, int) {
	if len(parts) == 0 {
		return "", 0
	}
	if len(parts) >= 2 {
		joined := parts[0] + "-" + parts[1]
		switch joined {
		case "text-embedding", "amazon-nova":
			return joined, 2
		}
	}
	return parts[0], 1
}

func normalizeVersionToken(token string) string {
	token = strings.TrimSpace(strings.TrimPrefix(token, "v"))
	token = strings.ReplaceAll(token, "-", ".")
	return token
}

func joinVersionParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ".")
}

func extractVersion(parts []string) (string, map[int]struct{}) {
	used := make(map[int]struct{})
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		switch {
		case datestampLike.MatchString(part):
			continue
		case versionLike.MatchString(part):
			used[i] = struct{}{}
			version := normalizeVersionToken(part)
			if numericToken.MatchString(strings.TrimPrefix(part, "v")) {
				versionParts := []string{normalizeVersionToken(part)}
				for j := i + 1; j < len(parts); j++ {
					next := parts[j]
					if !numericToken.MatchString(next) || datestampLike.MatchString(next) {
						break
					}
					used[j] = struct{}{}
					versionParts = append(versionParts, next)
				}
				version = joinVersionParts(versionParts)
			}
			return version, used
		case compactVersionLike.MatchString(part):
			used[i] = struct{}{}
			return strings.ToLower(part), used
		}
	}
	return "", used
}

func vendorForFamily(family string) string {
	if vendor, ok := vendorAliases[family]; ok {
		return vendor
	}
	if strings.HasPrefix(family, "text-embedding") {
		return "openai"
	}
	return family
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func versionDash(version string) string {
	return strings.ReplaceAll(version, ".", "-")
}

func versionWithV(version string) string {
	if version == "" || strings.HasPrefix(version, "v") || compactVersionLike.MatchString(version) {
		return version
	}
	return "v" + version
}

func joinModelParts(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(part, "- ")
		if part == "" {
			continue
		}
		filtered = append(filtered, part)
	}
	return strings.Join(filtered, "-")
}

func ParseNormalizedModelInfo(input string) NormalizedModelInfo {
	norm := NormalizeModelName(input)
	if norm == "" {
		return NormalizedModelInfo{}
	}
	parts := strings.Split(norm, "-")
	info := NormalizedModelInfo{Normalized: norm, Compact: compactModelName(input), Parts: parts}
	if len(parts) == 0 {
		return info
	}

	info.Family, _ = detectFamily(parts)
	info.Vendor = vendorForFamily(info.Family)
	startIndex := 1
	if info.Family == "text-embedding" || info.Family == "amazon-nova" {
		startIndex = 2
	}
	var usedVersion map[int]struct{}
	info.Version, usedVersion = extractVersion(parts[startIndex:])
	variants := make([]string, 0, len(parts))
	for offset, part := range parts[startIndex:] {
		if _, ok := usedVersion[offset]; ok {
			continue
		}
		if part == "" || datestampLike.MatchString(part) {
			continue
		}
		variants = append(variants, part)
	}
	info.Variants = uniqueStrings(variants)
	if len(info.Variants) > 0 {
		info.Variant = info.Variants[0]
		info.Suffix = strings.Join(info.Variants, "-")
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
	if info.Family != "" && info.Version != "" {
		versionDot := info.Version
		versionDashed := versionDash(info.Version)
		versionPrefixed := versionWithV(info.Version)
		if info.Suffix != "" {
			add(joinModelParts(info.Family, versionDot, info.Suffix))
			add(joinModelParts(info.Family, versionDashed, info.Suffix))
			add(joinModelParts(info.Family, versionPrefixed, info.Suffix))
			add(joinModelParts(info.Family, versionDash(versionPrefixed), info.Suffix))
			add(joinModelParts(info.Family, info.Suffix, versionDot))
			add(joinModelParts(info.Family, info.Suffix, versionDashed))
		}
		add(joinModelParts(info.Family, versionDot))
		add(joinModelParts(info.Family, versionDashed))
		add(joinModelParts(info.Family, versionPrefixed))
		add(joinModelParts(info.Family, versionDash(versionPrefixed)))
	}
	if info.Family != "" && info.Suffix != "" {
		add(joinModelParts(info.Family, info.Suffix))
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
	if info.Family != "" && info.Version != "" {
		if info.Suffix != "" {
			return joinModelParts(info.Family, info.Version, info.Suffix)
		}
		return joinModelParts(info.Family, info.Version)
	}
	return info.Normalized
}
