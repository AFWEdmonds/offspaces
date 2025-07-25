package main

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"
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
	Id     int64
	Name   string
	IsCity bool
}

func (o Offspace) String() string {
	return fmt.Sprintf("%d, %s, %s, %s, %s, %s, %s, %s, %d, %s", o.Id, o.Name, o.Bio, o.Street, o.Postcode, o.City, o.Website, o.SocialMedia, o.Photo, o.Published, o.EditKey)
}

// DB implement namespace
type DB struct {
	Db *sql.DB
}

var dbAdapter DB

func connectDb(username *string, password *string) {
	newDb, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/offspaces", *username, *password))
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

func (DB) queryOffspaces(checkPublished bool) ([]Offspace, error) {
	var offspaces []Offspace
	var rows *sql.Rows
	var err error
	if checkPublished {
		rows, err = dbAdapter.Db.Query("SELECT * FROM offspace WHERE published=false")
	} else {
		rows, err = dbAdapter.Db.Query("SELECT * FROM offspace")
	}
	if err != nil {
		return nil, fmt.Errorf("offspace: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var off Offspace
		if err := rows.Scan(&off.Id, &off.Name, &off.Name, &off.Street, &off.Postcode, &off.City, &off.Website, &off.SocialMedia, &off.Photo, &off.Published, &off.EditKey); err != nil {
			return nil, fmt.Errorf("offspace: %v", err)
		}
		offspaces = append(offspaces, off)
	}
	return offspaces, nil
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
