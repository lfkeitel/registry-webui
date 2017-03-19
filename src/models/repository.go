package models

type RepoStore interface {
	Save(r *Repo) error
	Delete(r *Repo) error
}

type Repo struct {
	store   RepoStore
	ID      int
	Name    string
	Private bool
}

func NewRepo(store RepoStore) *Repo {
	return &Repo{store: store}
}

func (r *Repo) Save() error {
	return r.store.Save(r)
}

func (r *Repo) Delete() error {
	return r.store.Delete(r)
}
