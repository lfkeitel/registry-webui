/*
This file was originally part of github.com/michaeljs1990/sqlitestore which is
licensed under the MIT license. Copyright (c) 2013 Contributors.
*/

package utils

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

type dbStore struct {
	db         *sql.DB
	stmtInsert *sql.Stmt
	stmtDelete *sql.Stmt
	stmtUpdate *sql.Stmt
	stmtSelect *sql.Stmt

	Codecs  []securecookie.Codec
	Options *Options
}

type Options struct {
	Path      string
	Domain    string
	MaxAge    int
	Secure    bool
	HttpOnly  bool
	TableName string
}

type sessionRow struct {
	id         string
	data       string
	createdOn  int64
	modifiedOn int64
}

func init() {
	gob.Register(time.Time{})
}

func newDBStore(db *DatabaseAccessor, options *Options, keyPairs ...[]byte) (*dbStore, error) {
	cTableQ := getTableCreate(db.Driver, options.TableName)
	if _, err := db.Exec(cTableQ); err != nil {
		return nil, err
		// switch err.(type) {
		// case *mysql.MySQLError:
		// 	// Error 1142 means permission denied for create command
		// 	if err.(*mysql.MySQLError).Number == 1142 {
		// 		break
		// 	}
		// 	return nil, err
		// case mysql.MySQLWarnings:
		// 	// Warning 1050 means table already exists
		// 	if err.(mysql.MySQLWarnings)[0].Code == "1050" {
		// 		break
		// 	}
		// default:
		// 	return nil, err
		// }
	}

	insQ := "INSERT INTO " + options.TableName +
		"(id, session_data, created_on, modified_on) VALUES (NULL, ?, ?, ?)"
	stmtInsert, stmtErr := db.Prepare(insQ)
	if stmtErr != nil {
		return nil, stmtErr
	}

	delQ := "DELETE FROM " + options.TableName + " WHERE id = ?"
	stmtDelete, stmtErr := db.Prepare(delQ)
	if stmtErr != nil {
		return nil, stmtErr
	}

	updQ := "UPDATE " + options.TableName + " SET session_data = ?, created_on = ?, modified_on = ? " +
		"WHERE id = ?"
	stmtUpdate, stmtErr := db.Prepare(updQ)
	if stmtErr != nil {
		return nil, stmtErr
	}

	selQ := "SELECT id, session_data, created_on, modified_on from " +
		options.TableName + " WHERE id = ?"
	stmtSelect, stmtErr := db.Prepare(selQ)
	if stmtErr != nil {
		return nil, stmtErr
	}

	return &dbStore{
		db:         db.DB,
		stmtInsert: stmtInsert,
		stmtDelete: stmtDelete,
		stmtUpdate: stmtUpdate,
		stmtSelect: stmtSelect,
		Codecs:     securecookie.CodecsFromPairs(keyPairs...),
		Options:    options,
	}, nil
}

func getTableCreate(typ, tableName string) string {
	switch typ {
	case "sqlite":
		return "CREATE TABLE IF NOT EXISTS " +
			tableName + " (id INTEGER PRIMARY KEY, " +
			"session_data LONGBLOB, " +
			"created_on BIGINT NOT NULL DEFAULT 0, " +
			"modified_on BIGINT NOT NULL DEFAULT 0);"
	}
	return ""
}

func (m *dbStore) close() {
	m.stmtSelect.Close()
	m.stmtUpdate.Close()
	m.stmtDelete.Close()
	m.stmtInsert.Close()
	m.db.Close()
}

func (m *dbStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(m, name)
}

func (m *dbStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(m, name)
	session.Options = &sessions.Options{
		Path:   m.Options.Path,
		MaxAge: m.Options.MaxAge,
	}
	session.IsNew = true
	var err error
	if cook, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, cook.Value, &session.ID, m.Codecs...)
		if err == nil {
			err = m.load(session)
			if err == nil {
				session.IsNew = false
			} else {
				err = nil
			}
		}
	}
	return session, err
}

func (m *dbStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	var err error
	if session.ID == "" {
		if err = m.insert(session); err != nil {
			return err
		}
	} else if err = m.save(session); err != nil {
		return err
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, m.Codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

func (m *dbStore) insert(session *sessions.Session) error {
	var createdOn time.Time
	crOn := session.Values["created_on"]
	if crOn == nil {
		createdOn = time.Now()
	} else {
		createdOn = crOn.(time.Time)
	}
	modifiedOn := time.Now()

	session.Values["created_on"] = createdOn
	session.Values["modified_on"] = modifiedOn

	encoded, encErr := securecookie.EncodeMulti(session.Name(), session.Values, m.Codecs...)
	if encErr != nil {
		return encErr
	}
	res, insErr := m.stmtInsert.Exec(encoded, createdOn.Unix(), modifiedOn.Unix())
	if insErr != nil {
		return insErr
	}
	lastInserted, lInsErr := res.LastInsertId()
	if lInsErr != nil {
		return lInsErr
	}
	session.ID = fmt.Sprintf("%d", lastInserted)
	return nil
}

func (m *dbStore) Delete(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	// Set cookie to expire.
	options := *session.Options
	options.MaxAge = -1
	http.SetCookie(w, sessions.NewCookie(session.Name(), "", &options))
	// Clear session values.
	for k := range session.Values {
		delete(session.Values, k)
	}

	_, delErr := m.stmtDelete.Exec(session.ID)
	if delErr != nil {
		return delErr
	}
	return nil
}

func (m *dbStore) save(session *sessions.Session) error {
	if session.IsNew {
		return m.insert(session)
	}
	modifiedOn := time.Now()
	var createdOn time.Time
	crOn := session.Values["created_on"]
	if crOn == nil {
		createdOn = time.Now()
	} else {
		createdOn = crOn.(time.Time)
	}

	session.Values["created_on"] = createdOn
	session.Values["modified_on"] = modifiedOn
	encoded, encErr := securecookie.EncodeMulti(session.Name(), session.Values, m.Codecs...)
	if encErr != nil {
		return encErr
	}
	_, updErr := m.stmtUpdate.Exec(encoded, createdOn.Unix(), modifiedOn.Unix(), session.ID)
	if updErr != nil {
		return updErr
	}
	return nil
}

func (m *dbStore) load(session *sessions.Session) error {
	row := m.stmtSelect.QueryRow(session.ID)
	sess := sessionRow{}
	scanErr := row.Scan(&sess.id, &sess.data, &sess.createdOn, &sess.modifiedOn)
	if scanErr != nil {
		return scanErr
	}
	err := securecookie.DecodeMulti(session.Name(), sess.data, &session.Values, m.Codecs...)
	if err != nil {
		return err
	}
	session.Values["created_on"] = time.Unix(sess.createdOn, 0)
	session.Values["modified_on"] = time.Unix(sess.modifiedOn, 0)
	return nil
}
