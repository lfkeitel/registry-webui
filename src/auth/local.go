package auth

import (
	"net/http"

	"github.com/lfkeitel/registry-webui/src/models/stores"
	"github.com/lfkeitel/registry-webui/src/utils"
	"github.com/lfkeitel/verbose"
	"golang.org/x/crypto/bcrypt"
)

func checkLogin(username, password string, r *http.Request) bool {
	e := utils.GetEnvironmentFromContext(r)
	hashedPassword, err := stores.GetUserStore(e.DB).GetUserPassword(username)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "auth:local",
		}).Errorf("Error getting user")
		return false
	}

	if hashedPassword == "" { // User doesn't have a local password
		return false
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		if err != bcrypt.ErrMismatchedHashAndPassword {
			e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "auth:local",
			}).Debug("Bcrypt failed")
		}
		return false
	}

	return true
}
