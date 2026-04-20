// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"api/internal/logic"
	"api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetCurrentWeeklyVoteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetCurrentWeeklyVoteLogic(r.Context(), svcCtx)
		resp, err := l.GetCurrentWeeklyVote()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
