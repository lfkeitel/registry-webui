package models

import (
	"context"
	"net/http"

	"github.com/lfkeitel/registry-webui/src/utils"
)

func GetUserFromContext(r *http.Request) *User {
	if rv := r.Context().Value(utils.SessionUserKey); rv != nil {
		return rv.(*User)
	}
	return nil
}

func SetUserToContext(r *http.Request, u *User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), utils.SessionUserKey, u))
}
