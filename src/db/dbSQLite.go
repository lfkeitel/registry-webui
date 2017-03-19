// +build dbsqlite dball

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/lfkeitel/registry-webui/src/utils"
	"github.com/lfkeitel/verbose"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func init() {
	RegisterDatabaseAccessor("sqlite", newSQLiteDBInit())
}

type sqliteDB struct {
	createFuncs  map[string]func(*utils.DatabaseAccessor) error
	migrateFuncs []func(*utils.DatabaseAccessor) error
}

func newSQLiteDBInit() *sqliteDB {
	s := &sqliteDB{}

	s.createFuncs = map[string]func(*utils.DatabaseAccessor) error{
		"settings":      s.createSettingTable,
		"user":          s.createUserTable,
		"organization":  s.createOrgTable,
		"team":          s.createTeamTable,
		"repo":          s.createRepoTable,
		"repo_acl":      s.createRepoACLTable,
		"user_team_org": s.createUserTeamOrgTable,
	}

	s.migrateFuncs = []func(*utils.DatabaseAccessor) error{}

	return s
}

func (s *sqliteDB) connect(d *utils.DatabaseAccessor, c *utils.Config) error {
	var err error
	if err = os.MkdirAll(path.Dir(c.Database.Address), os.ModePerm); err != nil {
		return fmt.Errorf("Failed to create directories: %v", err)
	}
	d.DB, err = sql.Open("sqlite3", c.Database.Address)
	if err != nil {
		return err
	}

	err = d.DB.Ping()
	if err != nil {
		return err
	}

	_, err = d.Exec("PRAGMA foreign_keys = ON")
	return err
}

func (s *sqliteDB) createTables(d *utils.DatabaseAccessor) error {
	rows, err := d.DB.Query(`SELECT name FROM sqlite_master WHERE type='table'`)
	if err != nil {
		return err
	}
	defer rows.Close()
	tables := make(map[string]bool)
	for _, table := range utils.DatabaseTableNames {
		tables[table] = false
	}

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		tables[tableName] = true
	}

	for table, create := range s.createFuncs {
		if !tables[table] {
			if err := create(d); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *sqliteDB) migrateTables(d *utils.DatabaseAccessor) error {
	var currDBVer int
	verRow := d.DB.QueryRow(`SELECT "value" FROM "settings" WHERE "id" = 'db_version'`)
	if verRow == nil {
		return errors.New("Failed to get database version")
	}
	verRow.Scan(&currDBVer)

	utils.SystemLogger.WithFields(verbose.Fields{
		"current-version": currDBVer,
		"active-version":  dbVersion,
	}).Debug("Database Versions")

	// No migration needed
	if currDBVer == dbVersion {
		return nil
	}

	neededMigrations := s.migrateFuncs[currDBVer:dbVersion]
	for _, migrate := range neededMigrations {
		if migrate == nil {
			continue
		}
		if err := migrate(d); err != nil {
			return err
		}
	}

	_, err := d.DB.Exec(`UPDATE "settings" SET "value" = ? WHERE "id" = 'db_version'`, dbVersion)
	return err
}

func (s *sqliteDB) init(d *utils.DatabaseAccessor, c *utils.Config) error {
	if err := s.connect(d, c); err != nil {
		return err
	}

	d.Driver = "sqlite"

	if err := s.createTables(d); err != nil {
		return err
	}

	return s.migrateTables(d)
}

func (s *sqliteDB) createSettingTable(d *utils.DatabaseAccessor) error {
	sql := `CREATE TABLE "settings" (
	    "id" TEXT PRIMARY KEY NOT NULL,
	    "value" TEXT DEFAULT ''
	)`

	if _, err := d.DB.Exec(sql); err != nil {
		return err
	}

	_, err := d.DB.Exec(`INSERT INTO "settings" ("id", "value") VALUES ('db_version', ?)`, dbVersion)
	return err
}

func (s *sqliteDB) createUserTable(d *utils.DatabaseAccessor) error {
	sql := `CREATE TABLE "user" (
		"id" INTEGER PRIMARY KEY AUTOINCREMENT,
		"name" TEXT NOT NULL UNIQUE ON CONFLICT ROLLBACK,
		"displayname" TEXT DEFAULT '',
		"password" TEXT NOT NULL,
		"created" INTEGER NOT NULL
	)`

	if _, err := d.DB.Exec(sql); err != nil {
		return err
	}

	_, err := d.DB.Exec(`INSERT INTO "user" ("name","displayname","password","created") VALUES ("admin","Admin","$2a$10$fjjjYKwy/Y6cWWZBLhqG6OfUcxq3EVwHm/beQHT.S8K2AVhnjeG.G",?)`, time.Now().Unix())
	return err
}

func (s *sqliteDB) createOrgTable(d *utils.DatabaseAccessor) error {
	sql := `CREATE TABLE "organization" (
		"id"  INTEGER PRIMARY KEY AUTOINCREMENT,
		"namespace" TEXT NOT NULL UNIQUE ON CONFLICT ROLLBACK,
		"displayname" TEXT DEFAULT ''
	)`

	if _, err := d.DB.Exec(sql); err != nil {
		return err
	}

	_, err := d.DB.Exec(`INSERT INTO "organization" ("namespace","displayname") VALUES ("","Admin"),("_","Library")`)
	return err
}

func (s *sqliteDB) createTeamTable(d *utils.DatabaseAccessor) error {
	sql := `CREATE TABLE "team" (
		"id"  INTEGER PRIMARY KEY AUTOINCREMENT,
		"orgid" INTEGER NOT NULL,
		"name" TEXT NOT NULL DEFAULT ''
	)`

	if _, err := d.DB.Exec(sql); err != nil {
		return err
	}

	_, err := d.DB.Exec(`INSERT INTO "team" ("orgid","name") VALUES (1,"owners"),(2,"owners")`)
	return err
}

func (s *sqliteDB) createRepoTable(d *utils.DatabaseAccessor) error {
	sql := `CREATE TABLE "repo" (
		"id"  INTEGER PRIMARY KEY AUTOINCREMENT,
		"name" TEXT NOT NULL DEFAULT '',
		"private" INTEGER DEFAULT 0
	)`
	_, err := d.DB.Exec(sql)
	return err
}

func (s *sqliteDB) createRepoACLTable(d *utils.DatabaseAccessor) error {
	sql := `CREATE TABLE "repo_acl" (
		"id"  INTEGER PRIMARY KEY AUTOINCREMENT,
		"repoid" INTEGER NOT NULL,
		"teamid" INTEGER NOT NULL,
		"access" INTEGER NOT NULL DEFAULT 0
	)`
	_, err := d.DB.Exec(sql)
	return err
}

func (s *sqliteDB) createUserTeamOrgTable(d *utils.DatabaseAccessor) error {
	sql := `CREATE TABLE "user_team_org" (
		"userid" INTEGER NOT NULL,
		"teamid" INTEGER NOT NULL,
		"orgid" INTEGER NOT NULL
	)`

	if _, err := d.DB.Exec(sql); err != nil {
		return err
	}

	_, err := d.DB.Exec(`INSERT INTO "user_team_org" VALUES (1,1,1), (1,2,2)`)
	return err
}
