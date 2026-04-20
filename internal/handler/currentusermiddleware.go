package handler

import (
	"net/http"

	"api/internal/logic"

	"github.com/zeromicro/go-zero/rest"
)

const currentUserIDHeader = "X-User-Id"
const currentAdminIDHeader = "X-Admin-Id"

func CurrentUserMiddleware() rest.Middleware {
	return rest.ToMiddleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := logic.ParseCurrentUserID(r.Header.Get(currentUserIDHeader))
			if userID > 0 {
				r = r.WithContext(logic.WithCurrentUserID(r.Context(), userID))
			}

			adminID := logic.ParseCurrentAdminID(r.Header.Get(currentAdminIDHeader))
			if adminID > 0 {
				r = r.WithContext(logic.WithCurrentAdminID(r.Context(), adminID))
			}

			next.ServeHTTP(w, r)
		})
	})
}
