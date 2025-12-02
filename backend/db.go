package main

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

type OpeningDay struct {
	Start string
	End   string
}

type OpeningTimes struct {
	Mon OpeningDay
	Tue OpeningDay
	Wed OpeningDay
	Thu OpeningDay
	Fri OpeningDay
	Sat OpeningDay
	Sun OpeningDay
}

type Offspace struct {
	Id           int
	Name         string
	Street       string
	Postcode     string
	City         string
	Website      string
	SocialMedia  string
	Photo        string
	Published    bool
	EditKey      string
	OpeningTimes OpeningTimes
}

type Query struct {
	Text           string
	Index          int
	DisplayAmount  int
	RequireOpenNow bool
	RequireShowOn  bool
	SearchName     bool
	SearchAddress  bool
	SearchShow     bool
	SortBy         string
}

func (o Offspace) String() string {
	return fmt.Sprintf("%d, %s, %s, %s, %s, %s, %s, %s, %d, %s", o.Id, o.Name, o.Street, o.Postcode, o.City, o.Website, o.SocialMedia, o.Photo, o.Published, o.EditKey)
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

func (DB) queryOffspaces(showUnpublished bool, q Query) ([]Offspace, error) {
	// language=SQL
	base := `SELECT 
                id, name, street, postcode, city,
                website, social_media, photo, published,
                edit_key, opening_times
             FROM offspace`

	conditions := []string{}
	args := []interface{}{}

	if !showUnpublished {
		conditions = append(conditions, "published = TRUE")
	}

	if q.Text != "" {
		like := "%" + q.Text + "%"

		var searchParts []string

		if q.SearchName {
			searchParts = append(searchParts, "name LIKE ?")
			args = append(args, like)
		}

		if q.SearchAddress {
			searchParts = append(searchParts, "(street LIKE ? OR city LIKE ? OR postcode LIKE ?)")
			args = append(args, like, like, like)
		}

		if q.SearchShow {
			// placeholder until "show" exists
			searchParts = append(searchParts, "0 = 1") // always false for now
		}

		// default → name search
		if len(searchParts) == 0 {
			searchParts = append(searchParts, "name LIKE ?")
			args = append(args, like)
		}

		conditions = append(conditions, "("+strings.Join(searchParts, " OR ")+")")
	}

	// -------------------------------------------------------
	// 3. requireShowOn
	// -------------------------------------------------------
	if q.RequireShowOn {
		// placeholder until "show" exists
		conditions = append(conditions, "social_media <> ''")
	}

	// -------------------------------------------------------
	// 4. Open-now filtering
	// -------------------------------------------------------
	if q.RequireOpenNow {
		now := time.Now()
		weekday := strings.ToLower(now.Weekday().String()[:3]) // mon/tue/wed...
		hour := now.Hour()

		conditions = append(conditions,
			fmt.Sprintf("JSON_EXTRACT(opening_times, '$.%s[%d]') = true", weekday, hour),
		)
	}

	// -------------------------------------------------------
	// WHERE clause
	// -------------------------------------------------------
	sqlQuery := base
	if len(conditions) > 0 {
		sqlQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	// -------------------------------------------------------
	// Sorting
	// -------------------------------------------------------
	switch q.SortBy {
	case "name":
		sqlQuery += " ORDER BY name ASC"
	case "city":
		sqlQuery += " ORDER BY city ASC"
	case "newest":
		sqlQuery += " ORDER BY id DESC"
	default:
		sqlQuery += " ORDER BY id ASC"
	}

	// -------------------------------------------------------
	// Pagination
	// -------------------------------------------------------
	if q.DisplayAmount <= 0 {
		q.DisplayAmount = 50
	}
	offset := q.Index * q.DisplayAmount

	sqlQuery += " LIMIT ? OFFSET ?"
	args = append(args, q.DisplayAmount, offset)

	// -------------------------------------------------------
	// Execute
	// -------------------------------------------------------
	rows, err := dbAdapter.Db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("queryOffspaces: %v", err)
	}
	defer rows.Close()

	var result []Offspace

	for rows.Next() {
		var off Offspace
		var openingJSON []byte

		err := rows.Scan(
			&off.Id,
			&off.Name,
			&off.Street,
			&off.Postcode,
			&off.City,
			&off.Website,
			&off.SocialMedia,
			&off.Photo,
			&off.Published,
			&off.EditKey,
			&openingJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("scan: %v", err)
		}

		if len(openingJSON) > 0 {
			if err := json.Unmarshal(openingJSON, &off.openingTimes); err != nil {
				return nil, fmt.Errorf("opening_times JSON: %v", err)
			}
		}

		result = append(result, off)
	}

	return result, nil
}

func (DB) getOffspaceByKey(key string) (Offspace, error) {
	var off Offspace
	rows, err := dbAdapter.Db.Query("SELECT * FROM offspace WHERE edit_key=?", key)
	if err != nil {
		return Offspace{}, fmt.Errorf("offspace: %v", err)
	}
	defer rows.Close()
	rows.Next()
	if err := rows.Scan(&off.Id, &off.Name, &off.Street, &off.Postcode, &off.City, &off.Website, &off.SocialMedia, &off.Photo, &off.Published, &off.EditKey); err != nil {
		return Offspace{}, fmt.Errorf("offspace: %v", err)
	}
	return off, nil
}

func (DB) createOffspace(o OffspaceRest) error {
	editUuid, err := uuid.NewRandom()
	rows, err := dbAdapter.Db.Exec("INSERT INTO offspace (name, street, postcode, city, website, social_media, photo, published, edit_key) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		o.Name, o.Street, o.Postcode, o.City, o.Website, o.SocialMedia, o.Photo, false, editUuid)
	if err != nil {
		return err
	}
	lastInsertId, err := rows.LastInsertId()
	fmt.Println(fmt.Sprintf("Added row with id: %d, and contents %d %s %s %s %s %s %s %s", lastInsertId, o.ID, o.Name, o.Street, o.City, o.Postcode, o.Website, o.SocialMedia))
	return err
}

func (DB) updateOffspace(o Offspace, admin bool) error {
	// published gets reset to false everytime the listing is edited by a non-admin
	if !admin {
		o.Published = false
	}
	rows, err := dbAdapter.Db.Exec("UPDATE offspace SET name = ?, street = ?, postcode = ?, city = ?, website = ?, social_media = ?, photo = ?, published = ?, edit_key = ? WHERE edit_key = ?",
		o.Name, o.Street, o.Postcode, o.City, o.Website, o.SocialMedia, o.Photo, o.Published, o.EditKey, o.EditKey)
	if err != nil {
		return err
	}
	lastInsertId, err := rows.LastInsertId()
	fmt.Println(fmt.Sprintf("Added row with id: %d, and contents %s", lastInsertId, o.String()))
	return err
}

func (o openingTimes) Value() (driver.Value, error) {
	return json.Marshal(o) // returns []byte containing JSON
}

func (o *openingTimes) Scan(src any) error {
	if src == nil {
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("openingTimes.Scan: expected []byte, got %T", src)
	}

	return json.Unmarshal(b, o)
}
