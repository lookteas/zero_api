package handler

import (
	"net/http"

	"api/internal/logic"
	"api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateAwarenessCheckHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewCreateAwarenessCheckLogic(r.Context(), svcCtx)
		resp, err := l.CreateAwarenessCheck()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
