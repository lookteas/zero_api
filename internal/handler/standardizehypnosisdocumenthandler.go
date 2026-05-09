package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"api/internal/hypnosis"
	"api/internal/svc"
)

func StandardizeHypnosisDocumentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, ok := readUploadedDocx(w, r, svcCtx)
		if !ok {
			return
		}

		rulesPath := svcCtx.Config.Hypnosis.ReplacementRulesPath
		if rulesPath == "" {
			rulesPath = "etc/hypnosis-replacements.json"
		}
		rules, err := hypnosis.LoadReplacementRules(rulesPath)
		if err != nil {
			http.Error(w, "读取替换规则失败", http.StatusInternalServerError)
			return
		}

		output, err := hypnosis.StandardizeDocx(content, hypnosis.StandardizeOptions{
			Topic:       strings.TrimSpace(r.FormValue("topic")),
			Date:        strings.TrimSpace(r.FormValue("date")),
			Duration:    strings.TrimSpace(r.FormValue("duration")),
			HostName:    strings.TrimSpace(r.FormValue("hostName")),
			SubjectName: strings.TrimSpace(r.FormValue("subjectName")),
			Rules:       rules,
		})
		if err != nil {
			http.Error(w, "处理 docx 文件失败", http.StatusBadRequest)
			return
		}

		filename := fmt.Sprintf("hypnosis-standardized-%s.docx", time.Now().Format("20060102150405"))
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(output)))
		_, _ = w.Write(output)
	}
}
