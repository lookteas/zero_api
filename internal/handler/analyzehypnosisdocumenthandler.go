package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"api/internal/hypnosis"
	"api/internal/svc"
)

type hypnosisAnalysisResp struct {
	Code    int64                `json:"code"`
	Message string               `json:"message"`
	Data    hypnosisAnalysisData `json:"data"`
}

type hypnosisAnalysisData struct {
	Speakers []hypnosis.SpeakerOption `json:"speakers"`
	Date     string                   `json:"date"`
	Duration string                   `json:"duration"`
}

func AnalyzeHypnosisDocumentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, ok := readUploadedDocx(w, r, svcCtx)
		if !ok {
			return
		}

		analysis, err := hypnosis.AnalyzeDocx(content)
		if err != nil {
			http.Error(w, "解析 docx 文件失败", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(hypnosisAnalysisResp{
			Code:    0,
			Message: "ok",
			Data: hypnosisAnalysisData{
				Speakers: analysis.Speakers,
				Date:     time.Now().Format("2006年01月02日"),
				Duration: analysis.Duration,
			},
		})
	}
}

func readUploadedDocx(w http.ResponseWriter, r *http.Request, svcCtx *svc.ServiceContext) ([]byte, bool) {
	maxUploadBytes := svcCtx.Config.Hypnosis.MaxUploadBytes
	if maxUploadBytes <= 0 {
		maxUploadBytes = 20 << 20
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		http.Error(w, "上传文件过大或表单格式不正确", http.StatusBadRequest)
		return nil, false
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "请上传 docx 文件", http.StatusBadRequest)
		return nil, false
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".docx") {
		http.Error(w, "仅支持 docx 文件", http.StatusBadRequest)
		return nil, false
	}

	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "读取上传文件失败", http.StatusBadRequest)
		return nil, false
	}

	return content, true
}
