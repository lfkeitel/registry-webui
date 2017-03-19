package controllers

import (
	"net/http"

	"github.com/lfkeitel/registry-webui/src/auth"
	"github.com/lfkeitel/registry-webui/src/utils"
)

type Auth struct {
	e *utils.Environment
}

func NewAuthController(e *utils.Environment) *Auth {
	return &Auth{e: e}
}

func (c *Auth) Login(w http.ResponseWriter, r *http.Request) {
	if auth.IsLoggedIn(r) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	c.e.Views.NewView("login", r).Render(w, nil)
}

func (c *Auth) LoginPost(w http.ResponseWriter, r *http.Request) {
	if auth.LoginUser(w, r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	sess := utils.GetSessionFromContext(r)
	sess.AddFlash("Incorrect username or password")

	c.e.Views.NewView("login", r).Render(w, nil)
}

func (c *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	auth.LogoutUser(r, w)
	if _, ok := r.URL.Query()["noredirect"]; ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}
