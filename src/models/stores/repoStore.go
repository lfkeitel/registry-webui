package stores

import (
	"strings"

	"github.com/lfkeitel/registry-webui/src/models"
	"github.com/lfkeitel/registry-webui/src/utils"
)

type RepoStore struct {
	db *utils.DatabaseAccessor
}

var repoStore *RepoStore

func newRepoStore(db *utils.DatabaseAccessor) *RepoStore {
	return &RepoStore{db: db}
}

func GetRepoStore(db *utils.DatabaseAccessor) *RepoStore {
	if repoStore == nil {
		repoStore = newRepoStore(db)
	}
	return repoStore
}

func (s *RepoStore) GetRepoByName(name string) (*models.Repo, error) {
	if name == "" {
		return nil, nil
	}

	name = strings.ToLower(name)

	sql := `WHERE "name" = ?`
	repos, err := s.query(sql, name)
	if err != nil || repos == nil || len(repos) == 0 {
		return nil, err
	}
	return repos[0], nil
}

func (s *RepoStore) GetReposByNamespace(name string) ([]*models.Repo, error) {
	name = strings.ToLower(name) + "/%"
	sql := `WHERE "name" LIKE ?`
	return s.query(sql, name)
}

func (s *RepoStore) query(where string, values ...interface{}) ([]*models.Repo, error) {
	sql := `SELECT "id", "name", "private" FROM "repo" ` + where
	rows, err := s.db.Query(sql, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.Repo
	for rows.Next() {
		var id int
		var name string
		var private bool

		err := rows.Scan(
			&id,
			&name,
			&private,
		)
		if err != nil {
			continue
		}

		repo := models.NewRepo(s)
		repo.ID = id
		repo.Name = name
		repo.Private = private
		results = append(results, repo)
	}
	return results, nil
}

func (s *RepoStore) Save(u *models.Repo) error {
	return nil
}

func (s *RepoStore) Delete(u *models.Repo) error {
	return nil
}
