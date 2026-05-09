package hypnosis

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStandardizeRealDocxWhenSamplePathIsProvided(t *testing.T) {
	samplePath := os.Getenv("ZERO_HYPNOSIS_SAMPLE_DOCX")
	if samplePath == "" {
		t.Skip("ZERO_HYPNOSIS_SAMPLE_DOCX is not set")
	}

	content, err := os.ReadFile(samplePath)
	if err != nil {
		t.Fatalf("read sample docx: %v", err)
	}

	rules, err := LoadReplacementRules(filepath.Join("..", "..", "etc", "hypnosis-replacements.json"))
	if err != nil {
		t.Fatalf("load replacement rules: %v", err)
	}

	output, err := StandardizeDocx(content, StandardizeOptions{
		Topic:         "潜意识探索",
		Date:          "2026年02月09日",
		Duration:      "约2小时",
		HostName:      "星岩",
		SubjectName:   "瑞祥",
		HostReview:    "主催复盘",
		SubjectReview: "被催复盘",
		Rules:         rules,
	})
	if err != nil {
		t.Fatalf("standardize sample docx: %v", err)
	}
	if len(output) == 0 {
		t.Fatal("expected output docx bytes")
	}
}
