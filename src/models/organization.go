package models

type OrganizationStore interface {
	Save(o *Organization) error
	Delete(o *Organization) error
}

type Organization struct {
	store       OrganizationStore
	ID          int
	Namespace   string
	DisplayName string
}

func NewOrganization(store OrganizationStore) *Organization {
	return &Organization{store: store}
}

func (o *Organization) Save() error {
	return o.store.Save(o)
}

func (o *Organization) Delete() error {
	return o.store.Delete(o)
}
