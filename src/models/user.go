package models

import "time"

type UserStore interface {
	Save(u *User) error
	SavePassword(u *User, password string) error
	Delete(u *User) error
}

type User struct {
	store       UserStore
	ID          int
	Name        string
	DisplayName string
	Created     time.Time
}

func NewUser(store UserStore) *User {
	return &User{store: store}
}

func (u *User) Save() error {
	return u.store.Save(u)
}

func (u *User) SavePassword(password string) error {
	return u.store.SavePassword(u, password)
}

func (u *User) Delete() error {
	return u.store.Delete(u)
}
