package stores

import (
	"strings"

	"github.com/lfkeitel/registry-webui/src/models"
	"github.com/lfkeitel/registry-webui/src/utils"
)

type TeamStore struct {
	db *utils.DatabaseAccessor
}

var teamStore *TeamStore

func newTeamStore(db *utils.DatabaseAccessor) *TeamStore {
	return &TeamStore{db: db}
}

func GetTeamStore(db *utils.DatabaseAccessor) *TeamStore {
	if teamStore == nil {
		teamStore = newTeamStore(db)
	}
	return teamStore
}

func (s *TeamStore) GetTeamByName(name string) (*models.Team, error) {
	if name == "" {
		return nil, nil
	}

	name = strings.ToLower(name)

	sql := `WHERE "name" = ?`
	teams, err := s.query(sql, name)
	if err != nil || teams == nil || len(teams) == 0 {
		return nil, err
	}
	return teams[0], nil
}

func (s *TeamStore) query(where string, values ...interface{}) ([]*models.Team, error) {
	sql := `SELECT "id", "orgid", "name" FROM "team" ` + where
	rows, err := s.db.Query(sql, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.Team
	for rows.Next() {
		var id int
		var orgid int
		var name string

		err := rows.Scan(
			&id,
			&orgid,
			&name,
		)
		if err != nil {
			continue
		}

		team := models.NewTeam(s)
		team.ID = id
		team.OrgID = orgid
		team.Name = name
		results = append(results, team)
	}
	return results, nil
}

func (s *TeamStore) Save(u *models.Team) error {
	return nil
}

func (s *TeamStore) Delete(u *models.Team) error {
	return nil
}
