package middleware

import (
	"net/http"

	"github.com/lfkeitel/registry-webui/src/models"
	"github.com/lfkeitel/registry-webui/src/models/stores"
	"github.com/lfkeitel/registry-webui/src/utils"
	"github.com/lfkeitel/verbose"
)

func SetSessionInfo(e *utils.Environment, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := e.Sessions.GetSession(r)
		sessionUser, err := stores.GetUserStore(e.DB).GetUserByName(session.GetString("username"))
		if err != nil {
			e.Log.WithFields(verbose.Fields{
				"error":    err,
				"package":  "middleware:session",
				"username": session.GetString("username"),
			}).Error("Error getting session user")
		}
		r = utils.SetSessionToContext(r, session)
		r = utils.SetEnvironmentToContext(r, e)
		r = models.SetUserToContext(r, sessionUser)

		// If running behind a proxy, set the RemoteAddr to the real address
		if r.Header.Get("X-Real-IP") != "" {
			r.RemoteAddr = r.Header.Get("X-Real-IP")
		}
		r = utils.SetIPToContext(r)

		next.ServeHTTP(w, r)
	})
}
