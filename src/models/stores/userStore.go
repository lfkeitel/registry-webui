package stores

import (
	"strings"
	"time"

	"github.com/lfkeitel/registry-webui/src/models"
	"github.com/lfkeitel/registry-webui/src/utils"
)

type UserStore struct {
	db *utils.DatabaseAccessor
}

var userStore *UserStore

func newUserStore(db *utils.DatabaseAccessor) *UserStore {
	return &UserStore{db: db}
}

func GetUserStore(db *utils.DatabaseAccessor) *UserStore {
	if userStore == nil {
		userStore = newUserStore(db)
	}
	return userStore
}

func (s *UserStore) GetUserByName(name string) (*models.User, error) {
	if name == "" {
		return nil, nil
	}

	name = strings.ToLower(name)

	sql := `WHERE "name" = ?`
	users, err := s.query(sql, name)
	if err != nil || users == nil || len(users) == 0 {
		return nil, err
	}
	return users[0], nil
}

func (s *UserStore) GetUserPassword(username string) (string, error) {
	sql := `SELECT "password" FROM "user" WHERE "name" = ?`
	var password string
	return password, s.db.QueryRow(sql, username).Scan(&password)
}

func (s *UserStore) query(where string, values ...interface{}) ([]*models.User, error) {
	sql := `SELECT "id", "name", "displayname", "created" FROM "user" ` + where
	rows, err := s.db.Query(sql, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.User
	for rows.Next() {
		var id int
		var name string
		var displayname string
		var created int64

		err := rows.Scan(
			&id,
			&name,
			&displayname,
			&created,
		)
		if err != nil {
			continue
		}

		user := models.NewUser(s)
		user.ID = id
		user.Name = name
		user.DisplayName = displayname
		user.Created = time.Unix(created, 0)
		results = append(results, user)
	}
	return results, nil
}

func (s *UserStore) Save(u *models.User) error {
	return nil
}

func (s *UserStore) SavePassword(u *models.User, password string) error {
	return nil
}

func (s *UserStore) Delete(u *models.User) error {
	return nil
}
