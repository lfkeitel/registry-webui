package models

type TeamStore interface {
	Save(t *Team) error
	Delete(t *Team) error
}

type Team struct {
	store TeamStore
	ID    int
	Name  string
	OrgID int
}

func NewTeam(store TeamStore) *Team {
	return &Team{store: store}
}

func (t *Team) Save() error {
	return t.store.Save(t)
}

func (t *Team) Delete() error {
	return t.store.Delete(t)
}
