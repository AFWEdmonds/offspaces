package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Offspace struct {
	Id          int64
	Name        string
	Bio         string
	Street      string
	Postcode    string
	City        Tag
	Website     string
	SocialMedia string
	Photo       string
	Published   bool
	EditKey     string
	Tags        []Tag
}

type Tag struct {
	Id        int64
	Name      string
	IsCity    bool
	Published bool
}

func (o Offspace) String() string {
	return fmt.Sprintf("%d, %s, %s, %s, %s, %s, %s, %s, %d, %s", o.Id, o.Name, o.Bio, o.Street, o.Postcode, o.City, o.Website, o.SocialMedia, o.Photo, o.Published, o.EditKey)
}

// DB implement namespace
type DB struct {
	Db *sqlx.DB
}

var dbAdapter DB

func connectDb(username *string, password *string) {
	newDb, err := sqlx.Open("mysql", fmt.Sprintf("%s:%s@/offspaces", *username, *password))
	if err != nil {
		log.Fatal(err)
	}
	newDb.SetConnMaxLifetime(time.Minute * 3)
	newDb.SetMaxOpenConns(10)
	newDb.SetMaxIdleConns(10)

	pingErr := newDb.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")
	dbAdapter.Db = newDb
}

func (DB) queryOffspaces(checkPublished bool, tags []int) ([]Offspace, error) {
	var offspaces []Offspace
	var query string
	var args []any
	var err error
	if tags == nil || len(tags) < 1 {
		// No tag filtering
		query = `
            SELECT *
            FROM offspace
            WHERE (? = FALSE OR published = TRUE)`
		args = []any{checkPublished}
	} else {
		// Query with tag filtering
		query, args, err = sqlx.In(`
            SELECT DISTINCT o.*
            FROM offspace o
            LEFT JOIN offspace_tag ot ON o.id = ot.offspace
            WHERE (? = FALSE OR o.published = TRUE)
              AND (o.city IN (?) OR ot.tag IN (?))`,
			checkPublished, tags, tags,
		)
		if err != nil {
			return nil, fmt.Errorf("offspace: %v", err)
		}
		query = dbAdapter.Db.Rebind(query)
	}
	err = dbAdapter.Db.Select(&offspaces, query, args...)
	if err != nil {
		return nil, fmt.Errorf("offspace: %v", err)
	}
	return offspaces, nil
}

func (DB) queryTags(checkPublished bool) ([]Tag, error) {
	var tags []Tag
	var rows *sql.Rows
	var err error
	if checkPublished {
		rows, err = dbAdapter.Db.Query("SELECT * FROM tag WHERE published=false")
	} else {
		rows, err = dbAdapter.Db.Query("SELECT * FROM tag")
	}
	if err != nil {
		return nil, fmt.Errorf("taq: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.Id, &tag.Name, &tag.Name); err != nil {
			return nil, fmt.Errorf("offspace: %v", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (DB) createOffspace(o OffspaceRest) error {
	editUuid, err := uuid.NewRandom()
	rows, err := dbAdapter.Db.Exec("INSERT INTO offspace (name, bio, street, postcode, city, website, social_media, photo, published, edit_key) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		o.Name, o.Bio, o.Street, o.Postcode, o.City, o.Website, o.SocialMedia, o.Photo, false, editUuid)
	if err != nil {
		return err
	}
	lastInsertId, err := rows.LastInsertId()
	fmt.Println(fmt.Sprintf("Added row with id: %d, and contents %d %s %s %s %s %s %s %s", lastInsertId, o.Id, o.Name, o.Bio, o.Street, o.City, o.Postcode, o.Website, o.SocialMedia))
	return err
}

func (DB) updateOffspace(o Offspace, admin bool) error {
	// published gets reset to false everytime the listing is edited by a non-admin
	if !admin {
		o.Published = false
	}
	rows, err := dbAdapter.Db.Exec("UPDATE offspace SET name = ?, bio = ?, street = ?, postcode = ?, city = ?, website = ?, social_media = ?, photo = ?, published = ?, edit_key = ? WHERE edit_key = ?",
		o.Name, o.Bio, o.Street, o.Postcode, o.City, o.Website, o.SocialMedia, o.Photo, o.Published, o.EditKey, o.EditKey)
	if err != nil {
		return err
	}
	lastInsertId, err := rows.LastInsertId()
	fmt.Println(fmt.Sprintf("Added row with id: %d, and contents %s", lastInsertId, o.String()))
	return err
}
