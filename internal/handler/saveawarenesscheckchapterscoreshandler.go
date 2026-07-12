package handler

import (
	"net/http"

	"api/internal/logic"
	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func SaveAwarenessCheckChapterScoresHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq types.AwarenessCheckChapterPathReq
		if err := httpx.ParsePath(r, &pathReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		var req types.AwarenessCheckScoreSaveReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		req.ChapterId = pathReq.ChapterId

		l := logic.NewSaveAwarenessCheckChapterScoresLogic(r.Context(), svcCtx)
		resp, err := l.SaveAwarenessCheckChapterScores(pathReq.ChapterId, &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
