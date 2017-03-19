package auth

import (
	"net/http"
	"strings"

	"github.com/lfkeitel/registry-webui/src/utils"
	"github.com/lfkeitel/verbose"
)

func LoginUser(w http.ResponseWriter, r *http.Request) bool {
	if r.FormValue("password") == "" || r.FormValue("username") == "" {
		return false
	}

	e := utils.GetEnvironmentFromContext(r)
	username := strings.ToLower(r.FormValue("username"))

	if checkLogin(username, r.FormValue("password"), r) {
		sess := utils.GetSessionFromContext(r)
		sess.Set("loggedin", true)
		sess.Set("username", username)
		sess.Save(r, w)
		e.Log.WithFields(verbose.Fields{
			"username": username,
			"action":   "login",
			"package":  "auth",
		}).Info("Logged in user")
		return true
	}

	e.Log.WithFields(verbose.Fields{
		"username": username,
		"package":  "auth",
	}).Info("Failed login")
	return false
}

func IsLoggedIn(r *http.Request) bool {
	return utils.GetSessionFromContext(r).GetBool("loggedin")
}

func LogoutUser(r *http.Request, w http.ResponseWriter) {
	sess := utils.GetSessionFromContext(r)
	if !sess.GetBool("loggedin") {
		return
	}
	username := sess.GetString("username")

	sess.Set("loggedin", false)
	sess.Set("username", "")
	sess.Save(r, w)

	e := utils.GetEnvironmentFromContext(r)
	e.Log.WithFields(verbose.Fields{
		"username": username,
		"action":   "logout",
		"package":  "auth",
	}).Info("Logged out user")
}
