package controllers

import (
	"net/http"

	"github.com/lfkeitel/registry-webui/src/models"
	"github.com/lfkeitel/registry-webui/src/models/stores"
	"github.com/lfkeitel/registry-webui/src/utils"
)

type Home struct {
	e *utils.Environment
}

func NewHomeController(e *utils.Environment) *Home {
	return &Home{e: e}
}

func (c *Home) Home(w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(r)

	orgs, err := stores.GetOrganizationStore(c.e.DB).GetOrgForUser(user)
	if err != nil {
		c.e.Log.WithField("err", err).Error("Failed to get user org membership")
		return
	}

	repoNamespace := r.URL.Query().Get("ns")
	if repoNamespace == "" {
		repoNamespace = user.Name
	}

	repos, err := stores.GetRepoStore(c.e.DB).GetReposByNamespace(repoNamespace)
	if err != nil {
		c.e.Log.WithField("err", err).Error("Failed to get user org membership")
		return
	}

	data := map[string]interface{}{
		"user":  user,
		"orgs":  orgs,
		"repos": repos,
	}
	c.e.Views.NewView("home", r).Render(w, data)
}
