package hypnosis

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReplacementRulesIgnoresMetadataAndLoadsRegex(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rules.json")
	if err := os.WriteFile(path, []byte(`{
		"_info": {"主题": "潜意识探索"},
		"灵核": ["零和", "林河"],
		"_speaker_replace": {"被催": ["瑞祥"]},
		"_regex": {
			"二灵": ["20"]
		}
	}`), 0600); err != nil {
		t.Fatalf("write rules file: %v", err)
	}

	rules, err := LoadReplacementRules(path)
	if err != nil {
		t.Fatalf("LoadReplacementRules returned error: %v", err)
	}

	if got := rules.Terms["灵核"]; len(got) != 2 || got[0] != "零和" || got[1] != "林河" {
		t.Fatalf("unexpected normal rules: %#v", got)
	}
	if _, ok := rules.Terms["_info"]; ok {
		t.Fatal("metadata should not be loaded as normal rule")
	}
	if got := rules.RegexTerms["二灵"]; len(got) != 1 || got[0] != "20" {
		t.Fatalf("unexpected regex rules: %#v", got)
	}
}
