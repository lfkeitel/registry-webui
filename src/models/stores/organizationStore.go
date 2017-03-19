package stores

import (
	"strings"

	"github.com/lfkeitel/registry-webui/src/models"
	"github.com/lfkeitel/registry-webui/src/utils"
)

type OrganizationStore struct {
	db *utils.DatabaseAccessor
}

var orgStore *OrganizationStore

func newOrganizationStore(db *utils.DatabaseAccessor) *OrganizationStore {
	return &OrganizationStore{db: db}
}

func GetOrganizationStore(db *utils.DatabaseAccessor) *OrganizationStore {
	if orgStore == nil {
		orgStore = newOrganizationStore(db)
	}
	return orgStore
}

func (s *OrganizationStore) GetOrgByName(name string) (*models.Organization, error) {
	if name == "" {
		return nil, nil
	}

	name = strings.ToLower(name)

	sql := `WHERE "namespace" = ?`
	orgs, err := s.query(sql, name)
	if err != nil || orgs == nil || len(orgs) == 0 {
		return nil, err
	}
	return orgs[0], nil
}

func (s *OrganizationStore) GetOrgForUser(u *models.User) ([]*models.Organization, error) {
	sql := `WHERE "id" in (SELECT "orgid" FROM "user_team_org" WHERE "userid" = ?)`
	return s.query(sql, u.ID)
}

func (s *OrganizationStore) query(where string, values ...interface{}) ([]*models.Organization, error) {
	sql := `SELECT "id", "namespace", "displayname" FROM "organization" ` + where
	rows, err := s.db.Query(sql, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.Organization
	for rows.Next() {
		var id int
		var namespace string
		var displayname string

		err := rows.Scan(
			&id,
			&namespace,
			&displayname,
		)
		if err != nil {
			continue
		}

		org := models.NewOrganization(s)
		org.ID = id
		org.Namespace = namespace
		org.DisplayName = displayname
		results = append(results, org)
	}
	return results, nil
}

func (s *OrganizationStore) Save(u *models.Organization) error {
	return nil
}

func (s *OrganizationStore) Delete(u *models.Organization) error {
	return nil
}
