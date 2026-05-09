package hypnosis

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func TestStandardizeDocxReplacesTermsAndSpeakers(t *testing.T) {
	input := buildTestDocx(t, []string{
		"星岩(00:00:01): 请连接功放系统和零和。",
		"瑞祥(00:00:09): 看到20，也有70%的真实度。",
	})

	output, err := StandardizeDocx(input, StandardizeOptions{
		Topic:       "潜意识探索",
		Date:        "2026年02月09日",
		Duration:    "约1分钟",
		HostName:    "星岩",
		SubjectName: "瑞祥",
		Rules: ReplacementRules{
			Terms: map[string][]string{
				"灵核":   {"零和"},
				"攻防系统": {"功放系统"},
			},
			RegexTerms: map[string][]string{
				"二灵": {"20"},
				"七灵": {"70"},
			},
		},
	})
	if err != nil {
		t.Fatalf("StandardizeDocx returned error: %v", err)
	}

	text := docxDocumentXMLText(t, output)
	for _, want := range []string{
		"互催主题：", "潜意识探索",
		"互催日期：", "2026年02月09日",
		"互催时长：", "约1分钟",
		"主催名称：", "星岩",
		"被催名称：", "瑞祥",
		"主催复盘：",
		"被催复盘：",
		"主催(00:00:01): 请连接攻防系统和灵核。",
		"被催(00:00:09): 看到二灵，也有70%的真实度。",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("document xml text missing %q in %q", want, text)
		}
	}

	for _, want := range []string{
		`<w:tblW w:w="5000" w:type="pct"/>`,
		`<w:jc w:val="left"/>`,
		`<w:tcW w:w="1600" w:type="dxa"/>`,
		`<w:tcW w:w="7800" w:type="dxa"/>`,
		`<w:color w:val="0485B0"/>`,
		`<w:color w:val="C26A1B"/>`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("document xml style missing %q in %q", want, text)
		}
	}
}

func TestAnalyzeDocxReturnsSpeakerOptionsAndEstimatedDuration(t *testing.T) {
	input := buildTestDocx(t, []string{
		"星岩(00:00:01): 放松。",
		"瑞祥(00:00:08): 好的。",
		"星岩(01:12:10): 结束。",
	})

	analysis, err := AnalyzeDocx(input)
	if err != nil {
		t.Fatalf("AnalyzeDocx returned error: %v", err)
	}

	if analysis.Speakers[0].Name != "星岩" || analysis.Speakers[0].Count != 2 {
		t.Fatalf("expected top speaker 星岩 count 2, got %#v", analysis.Speakers)
	}
	if analysis.Speakers[1].Name != "瑞祥" || analysis.Speakers[1].Count != 1 {
		t.Fatalf("expected second speaker 瑞祥 count 1, got %#v", analysis.Speakers)
	}
	if analysis.Duration != "约1小时13分钟" {
		t.Fatalf("expected estimated duration, got %q", analysis.Duration)
	}
}

func TestDetectSpeakersCountsTimestampedParagraphs(t *testing.T) {
	input := buildTestDocx(t, []string{
		"星岩(00:00:01): 放松。",
		"瑞祥(00:00:08): 好的。",
		"星岩(00:00:10): 继续。",
		"这不是说话人行。",
	})

	speakers, err := DetectSpeakers(input)
	if err != nil {
		t.Fatalf("DetectSpeakers returned error: %v", err)
	}

	if speakers["星岩"] != 2 {
		t.Fatalf("expected 星岩 count 2, got %d", speakers["星岩"])
	}
	if speakers["瑞祥"] != 1 {
		t.Fatalf("expected 瑞祥 count 1, got %d", speakers["瑞祥"])
	}
}

func buildTestDocx(t *testing.T, paragraphs []string) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	addZipFile(t, zw, "[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`)
	addZipFile(t, zw, "_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`)
	addZipFile(t, zw, "word/_rels/document.xml.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"/>`)
	addZipFile(t, zw, "word/document.xml", buildDocumentXML(paragraphs))

	if err := zw.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}

	return buf.Bytes()
}

func addZipFile(t *testing.T, zw *zip.Writer, name string, content string) {
	t.Helper()

	w, err := zw.Create(name)
	if err != nil {
		t.Fatalf("create zip entry %s: %v", name, err)
	}
	if _, err = w.Write([]byte(content)); err != nil {
		t.Fatalf("write zip entry %s: %v", name, err)
	}
}

func buildDocumentXML(paragraphs []string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	for _, paragraph := range paragraphs {
		b.WriteString(`<w:p><w:r><w:t>`)
		b.WriteString(paragraph)
		b.WriteString(`</w:t></w:r></w:p>`)
	}
	b.WriteString(`<w:sectPr/></w:body></w:document>`)
	return b.String()
}

func docxDocumentXMLText(t *testing.T, data []byte) string {
	t.Helper()

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open output docx: %v", err)
	}

	for _, file := range zr.File {
		if file.Name != "word/document.xml" {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			t.Fatalf("open document.xml: %v", err)
		}
		defer rc.Close()

		var buf bytes.Buffer
		if _, err = buf.ReadFrom(rc); err != nil {
			t.Fatalf("read document.xml: %v", err)
		}
		return buf.String()
	}

	t.Fatal("word/document.xml not found")
	return ""
}
