package middleware

import (
	"net/http"

	"github.com/lfkeitel/registry-webui/src/auth"
	"github.com/lfkeitel/registry-webui/src/utils"
)

func Auth(e *utils.Environment, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsLoggedIn(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
