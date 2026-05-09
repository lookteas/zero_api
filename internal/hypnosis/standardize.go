package hypnosis

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"regexp"
	"sort"
	"strings"
)

type ReplacementRules struct {
	Terms      map[string][]string
	RegexTerms map[string][]string
}

type StandardizeOptions struct {
	Topic       string
	Date        string
	Duration    string
	HostName    string
	SubjectName string
	Rules       ReplacementRules
}

var speakerLinePattern = regexp.MustCompile(`^([^(]+)\(([0-9:]+)\):`)

type SpeakerOption struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type DocxAnalysis struct {
	Speakers []SpeakerOption `json:"speakers"`
	Duration string          `json:"duration"`
}

func AnalyzeDocx(docx []byte) (DocxAnalysis, error) {
	documentXML, err := readDocumentXML(docx)
	if err != nil {
		return DocxAnalysis{}, err
	}

	counts := map[string]int{}
	lastSeconds := 0
	for _, paragraph := range documentParagraphTexts(documentXML) {
		matches := speakerLinePattern.FindStringSubmatch(paragraph)
		if len(matches) == 0 {
			continue
		}
		speaker := strings.TrimSpace(matches[1])
		if speaker != "" {
			counts[speaker]++
		}
		if seconds, ok := parseTimestampSeconds(matches[2]); ok && seconds > lastSeconds {
			lastSeconds = seconds
		}
	}

	speakers := make([]SpeakerOption, 0, len(counts))
	for name, count := range counts {
		speakers = append(speakers, SpeakerOption{Name: name, Count: count})
	}
	sort.Slice(speakers, func(i, j int) bool {
		if speakers[i].Count == speakers[j].Count {
			return speakers[i].Name < speakers[j].Name
		}
		return speakers[i].Count > speakers[j].Count
	})

	return DocxAnalysis{
		Speakers: speakers,
		Duration: formatEstimatedDuration(lastSeconds),
	}, nil
}

func DetectSpeakers(docx []byte) (map[string]int, error) {
	analysis, err := AnalyzeDocx(docx)
	if err != nil {
		return nil, err
	}

	speakers := map[string]int{}
	for _, speaker := range analysis.Speakers {
		speakers[speaker.Name] = speaker.Count
	}

	return speakers, nil
}

func StandardizeDocx(docx []byte, options StandardizeOptions) ([]byte, error) {
	documentXML, err := readDocumentXML(docx)
	if err != nil {
		return nil, err
	}

	standardizedXML := replaceTextNodes(documentXML, func(text string) string {
		return standardizeText(text, options)
	})
	standardizedXML = colorSpeakerParagraphs(standardizedXML)
	standardizedXML = prependInfoParagraphs(standardizedXML, options)

	return replaceDocumentXML(docx, standardizedXML)
}

func standardizeText(text string, options StandardizeOptions) string {
	text = standardizeSpeaker(text, options.HostName, "主催")
	text = standardizeSpeaker(text, options.SubjectName, "被催")

	for old, newValue := range flattenedTerms(options.Rules.Terms) {
		text = strings.ReplaceAll(text, old, newValue)
	}

	for _, rule := range regexRules(options.Rules.RegexTerms) {
		text = rule.replace(text)
	}

	return text
}

func standardizeSpeaker(text string, name string, role string) string {
	name = strings.TrimSpace(name)
	if name == "" || name == role {
		return text
	}

	pattern := regexp.MustCompile(`^` + regexp.QuoteMeta(name) + `\(([0-9:]+)\):`)
	return pattern.ReplaceAllString(text, role+`($1):`)
}

func flattenedTerms(rules map[string][]string) map[string]string {
	result := map[string]string{}
	targets := make([]string, 0, len(rules))
	for target := range rules {
		targets = append(targets, target)
	}
	sort.Strings(targets)

	for _, target := range targets {
		sources := append([]string(nil), rules[target]...)
		sort.Strings(sources)
		for _, source := range sources {
			if source != "" {
				result[source] = target
			}
		}
	}
	return result
}

type compiledRegexRule struct {
	target  string
	pattern *regexp.Regexp
}

func regexRules(rules map[string][]string) []compiledRegexRule {
	result := []compiledRegexRule{}
	targets := make([]string, 0, len(rules))
	for target := range rules {
		targets = append(targets, target)
	}
	sort.Strings(targets)

	for _, target := range targets {
		sources := append([]string(nil), rules[target]...)
		sort.Strings(sources)
		for _, source := range sources {
			if source == "" {
				continue
			}
			result = append(result, compiledRegexRule{
				target:  target,
				pattern: regexp.MustCompile(`(^|[^:\d])` + regexp.QuoteMeta(source) + `($|[^%\d:])`),
			})
		}
	}
	return result
}

func (rule compiledRegexRule) replace(text string) string {
	return rule.pattern.ReplaceAllString(text, `${1}`+rule.target+`${2}`)
}

func readDocumentXML(docx []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if err != nil {
		return "", fmt.Errorf("open docx: %w", err)
	}

	for _, file := range zr.File {
		if file.Name != "word/document.xml" {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("open word/document.xml: %w", err)
		}
		defer rc.Close()

		content, err := io.ReadAll(rc)
		if err != nil {
			return "", fmt.Errorf("read word/document.xml: %w", err)
		}
		return string(content), nil
	}

	return "", fmt.Errorf("word/document.xml not found")
}

func replaceDocumentXML(docx []byte, documentXML string) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if err != nil {
		return nil, fmt.Errorf("open docx: %w", err)
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	for _, file := range zr.File {
		header := file.FileHeader
		w, err := zw.CreateHeader(&header)
		if err != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("create zip entry %s: %w", file.Name, err)
		}

		if file.Name == "word/document.xml" {
			if _, err = w.Write([]byte(documentXML)); err != nil {
				_ = zw.Close()
				return nil, fmt.Errorf("write word/document.xml: %w", err)
			}
			continue
		}

		rc, err := file.Open()
		if err != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("open zip entry %s: %w", file.Name, err)
		}
		if _, err = io.Copy(w, rc); err != nil {
			_ = rc.Close()
			_ = zw.Close()
			return nil, fmt.Errorf("copy zip entry %s: %w", file.Name, err)
		}
		_ = rc.Close()
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("close docx: %w", err)
	}

	return buf.Bytes(), nil
}

func replaceTextNodes(documentXML string, transform func(string) string) string {
	return regexp.MustCompile(`<w:t([^>]*)>(.*?)</w:t>`).ReplaceAllStringFunc(documentXML, func(node string) string {
		matches := regexp.MustCompile(`(?s)<w:t([^>]*)>(.*?)</w:t>`).FindStringSubmatch(node)
		if len(matches) != 3 {
			return node
		}

		decoded := html.UnescapeString(matches[2])
		return `<w:t` + matches[1] + `>` + xmlEscape(transform(decoded)) + `</w:t>`
	})
}

func colorSpeakerParagraphs(documentXML string) string {
	paragraphPattern := regexp.MustCompile(`(?s)<w:p\b[^>]*>.*?</w:p>`)
	return paragraphPattern.ReplaceAllStringFunc(documentXML, func(paragraph string) string {
		text := paragraphText(paragraph)
		switch {
		case strings.HasPrefix(text, "主催("):
			return colorParagraphRuns(paragraph, "1F6F8B")
		case strings.HasPrefix(text, "被催("):
			return colorParagraphRuns(paragraph, "C26A1B")
		default:
			return paragraph
		}
	})
}

func colorParagraphRuns(paragraph string, color string) string {
	runPattern := regexp.MustCompile(`(?s)<w:r>(.*?)</w:r>`)
	return runPattern.ReplaceAllStringFunc(paragraph, func(run string) string {
		inner := strings.TrimSuffix(strings.TrimPrefix(run, "<w:r>"), "</w:r>")
		if strings.Contains(inner, "<w:rPr>") {
			return strings.Replace(run, "<w:rPr>", `<w:rPr><w:color w:val="`+color+`"/>`, 1)
		}
		return `<w:r><w:rPr><w:color w:val="` + color + `"/></w:rPr>` + inner + `</w:r>`
	})
}

func paragraphText(paragraphXML string) string {
	textPattern := regexp.MustCompile(`(?s)<w:t[^>]*>(.*?)</w:t>`)
	var b strings.Builder
	for _, textMatch := range textPattern.FindAllStringSubmatch(paragraphXML, -1) {
		b.WriteString(html.UnescapeString(textMatch[1]))
	}
	return b.String()
}

func documentParagraphTexts(documentXML string) []string {
	paragraphPattern := regexp.MustCompile(`(?s)<w:p\b[^>]*>(.*?)</w:p>`)
	textPattern := regexp.MustCompile(`(?s)<w:t[^>]*>(.*?)</w:t>`)

	paragraphs := []string{}
	for _, paragraphMatch := range paragraphPattern.FindAllStringSubmatch(documentXML, -1) {
		var b strings.Builder
		for _, textMatch := range textPattern.FindAllStringSubmatch(paragraphMatch[1], -1) {
			b.WriteString(html.UnescapeString(textMatch[1]))
		}
		if text := b.String(); strings.TrimSpace(text) != "" {
			paragraphs = append(paragraphs, text)
		}
	}

	return paragraphs
}

func prependInfoParagraphs(documentXML string, options StandardizeOptions) string {
	rows := [][2]string{
		{"互催主题：", options.Topic},
		{"互催日期：", options.Date},
		{"互催时长：", options.Duration},
		{"主催名称：", options.HostName},
		{"被催名称：", options.SubjectName},
	}

	var b strings.Builder
	b.WriteString(`<w:tbl><w:tblPr><w:tblW w:w="5000" w:type="pct"/><w:tblLayout w:type="fixed"/><w:jc w:val="left"/><w:tblBorders><w:top w:val="single" w:sz="4" w:color="000000"/><w:left w:val="single" w:sz="4" w:color="000000"/><w:bottom w:val="single" w:sz="4" w:color="000000"/><w:right w:val="single" w:sz="4" w:color="000000"/><w:insideH w:val="single" w:sz="4" w:color="000000"/><w:insideV w:val="single" w:sz="4" w:color="000000"/></w:tblBorders></w:tblPr><w:tblGrid><w:gridCol w:w="1600"/><w:gridCol w:w="7800"/></w:tblGrid>`)
	for _, row := range rows {
		b.WriteString(`<w:tr><w:tc><w:tcPr><w:tcW w:w="1600" w:type="dxa"/><w:vAlign w:val="center"/></w:tcPr><w:p><w:pPr><w:jc w:val="left"/></w:pPr><w:r><w:rPr><w:b/></w:rPr><w:t>`)
		b.WriteString(xmlEscape(row[0]))
		b.WriteString(`</w:t></w:r></w:p></w:tc><w:tc><w:tcPr><w:tcW w:w="7800" w:type="dxa"/><w:vAlign w:val="center"/></w:tcPr><w:p><w:pPr><w:jc w:val="left"/></w:pPr><w:r><w:t>`)
		b.WriteString(xmlEscape(row[1]))
		b.WriteString(`</w:t></w:r></w:p></w:tc></w:tr>`)
	}
	b.WriteString(`</w:tbl>`)
	b.WriteString(`<w:p/>`)
	b.WriteString(`<w:p><w:r><w:t>======================================================================</w:t></w:r></w:p>`)
	b.WriteString(`<w:p/>`)

	bodyStart := strings.Index(documentXML, "<w:body>")
	if bodyStart < 0 {
		return documentXML
	}

	insertAt := bodyStart + len("<w:body>")
	return documentXML[:insertAt] + b.String() + documentXML[insertAt:]
}

func parseTimestampSeconds(value string) (int, bool) {
	parts := strings.Split(value, ":")
	if len(parts) != 3 {
		return 0, false
	}

	total := 0
	for _, part := range parts {
		number := 0
		for _, char := range part {
			if char < '0' || char > '9' {
				return 0, false
			}
			number = number*10 + int(char-'0')
		}
		total = total*60 + number
	}
	return total, true
}

func formatEstimatedDuration(seconds int) string {
	if seconds <= 0 {
		return ""
	}

	minutes := seconds / 60
	if seconds%60 > 0 {
		minutes++
	}
	if minutes <= 0 {
		minutes = 1
	}

	hours := minutes / 60
	remainingMinutes := minutes % 60
	if hours == 0 {
		return fmt.Sprintf("约%d分钟", remainingMinutes)
	}
	if remainingMinutes == 0 {
		return fmt.Sprintf("约%d小时", hours)
	}
	return fmt.Sprintf("约%d小时%d分钟", hours, remainingMinutes)
}

func xmlEscape(value string) string {
	var b strings.Builder
	if err := xml.EscapeText(&b, []byte(value)); err != nil {
		return value
	}
	return b.String()
}
