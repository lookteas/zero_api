package hypnosis

import (
	"encoding/json"
	"fmt"
	"os"
)

type rawReplacementRules map[string]json.RawMessage

func LoadReplacementRules(path string) (ReplacementRules, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return ReplacementRules{}, fmt.Errorf("read replacement rules: %w", err)
	}

	var raw rawReplacementRules
	if err = json.Unmarshal(content, &raw); err != nil {
		return ReplacementRules{}, fmt.Errorf("parse replacement rules: %w", err)
	}

	rules := ReplacementRules{
		Terms:      map[string][]string{},
		RegexTerms: map[string][]string{},
	}

	for key, value := range raw {
		if key == "_regex" {
			var regexRules map[string][]string
			if err = json.Unmarshal(value, &regexRules); err != nil {
				return ReplacementRules{}, fmt.Errorf("parse regex replacement rules: %w", err)
			}
			rules.RegexTerms = regexRules
			continue
		}

		if len(key) > 0 && key[0] == '_' {
			continue
		}

		var sources []string
		if err = json.Unmarshal(value, &sources); err != nil {
			return ReplacementRules{}, fmt.Errorf("parse replacement rule %s: %w", key, err)
		}
		rules.Terms[key] = sources
	}

	return rules, nil
}
