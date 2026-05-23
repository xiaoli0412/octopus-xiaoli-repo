package llmname

import "testing"

func TestParseNormalizedModelInfoExtractsVendorVersionAndSuffix(t *testing.T) {
	info := ParseNormalizedModelInfo("DeepSeek V3.1 Turbo")

	if info.Vendor != "deepseek" {
		t.Fatalf("Vendor = %q, want deepseek", info.Vendor)
	}
	if info.Family != "deepseek" {
		t.Fatalf("Family = %q, want deepseek", info.Family)
	}
	if info.Version != "3.1" {
		t.Fatalf("Version = %q, want 3.1", info.Version)
	}
	if info.Suffix != "turbo" {
		t.Fatalf("Suffix = %q, want turbo", info.Suffix)
	}
	if len(info.Variants) != 1 || info.Variants[0] != "turbo" {
		t.Fatalf("Variants = %#v, want [turbo]", info.Variants)
	}
}

func TestCandidateModelKeysIncludeCanonicalDotVersionForms(t *testing.T) {
	keys := CandidateModelKeys("GLM 4.7 Pro")
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		seen[key] = struct{}{}
	}

	for _, want := range []string{
		"glm-4.7-pro",
		"glm-pro-4.7",
		"glm-4.7",
	} {
		if _, ok := seen[want]; !ok {
			t.Fatalf("CandidateModelKeys() missing %q in %#v", want, keys)
		}
	}
}

