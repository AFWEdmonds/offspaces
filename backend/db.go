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
	RequireExhibOn bool
	SearchName     bool
	SearchAddress  bool
	SearchExhib    bool
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

func (DB) queryOffspaces(showUnpublished bool, q Query) (int, []Offspace, error) {
	// language=SQL
	base := `SELECT 
                id, name, street, postcode, city,
                website, social_media, photo, published,
                edit_key, opening_times
             FROM offspace`

	conditions := []string{}
	args := []interface{}{}

	if !showUnpublished {
		conditions = append(conditions, "published = 1")
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

		if q.SearchExhib {
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

	if q.RequireExhibOn {
		// placeholder until "show" exists
		conditions = append(conditions, "social_media <> ''")
	}

	if q.RequireOpenNow {
		now := time.Now()
		wd := now.Weekday().String()[:3] // Mon, Tue, Wed...

		nowStr := now.Format("15:04") // HH:MM

		conditions = append(conditions,
			fmt.Sprintf(
				"JSON_UNQUOTE(JSON_EXTRACT(opening_times, '$.%s.Start')) <= ? AND "+
					"JSON_UNQUOTE(JSON_EXTRACT(opening_times, '$.%s.End')) >= ?",
				wd, wd,
			),
		)

		args = append(args, nowStr, nowStr)
	}

	sqlCountQuery := "SELECT COUNT(*) FROM offspace"
	sqlQuery := base
	if len(conditions) > 0 {
		sqlQuery += " WHERE " + strings.Join(conditions, " AND ")
		sqlCountQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

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

	if q.DisplayAmount <= 0 {
		q.DisplayAmount = 50
	}
	offset := q.Index * q.DisplayAmount

	sqlQuery += " LIMIT ? OFFSET ?"
	args = append(args, q.DisplayAmount, offset)

	rows, err := dbAdapter.Db.Query(sqlQuery, args...)
	if err != nil {
		return 0, nil, fmt.Errorf("queryOffspaces: %v", err)
	}
	defer rows.Close()

	var result []Offspace

	var total int
	err = dbAdapter.Db.QueryRow(sqlCountQuery).Scan(&total)
	if err != nil {
		return 0, nil, err
	}

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
			return 0, nil, fmt.Errorf("scan: %v", err)
		}

		if len(openingJSON) > 0 {
			if err := json.Unmarshal(openingJSON, &off.OpeningTimes); err != nil {
				return 0, nil, fmt.Errorf("opening_times JSON: %v", err)
			}
		}

		result = append(result, off)
	}

	return total, result, nil
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

func (DB) createOffspace(o OffspaceRest) (string, error) {
	editUuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	openingJSON, err := json.Marshal(o.Opening)
	if err != nil {
		return "", fmt.Errorf("marshal opening_times: %v", err)
	}

	res, err := dbAdapter.Db.Exec(`
        INSERT INTO offspace 
            (name, street, postcode, city, website, social_media, photo, published, edit_key, opening_times)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `,
		o.Name,
		o.Street,
		o.Postcode,
		o.City,
		o.Website,
		o.SocialMedia,
		o.Photo,
		false,       // published
		editUuid,    // edit_key
		openingJSON, // JSON data
	)
	if err != nil {
		return "", err
	}
	id, _ := res.LastInsertId()
	fmt.Printf("Created offspace %d (%s)\n", id, o.Name)
	return editUuid.String(), nil
}

func (DB) updateOffspace(o Offspace, admin bool) error {
	if !admin {
		o.Published = false
	}
	openingJSON, err := json.Marshal(o.OpeningTimes)
	if err != nil {
		return fmt.Errorf("marshal opening_times: %v", err)
	}
	_, err = dbAdapter.Db.Exec(`
        UPDATE offspace SET 
            name = ?, 
            street = ?, 
            postcode = ?, 
            city = ?, 
            website = ?, 
            social_media = ?, 
            photo = ?, 
            published = ?, 
            opening_times = ?
        WHERE edit_key = ?
    `,
		o.Name,
		o.Street,
		o.Postcode,
		o.City,
		o.Website,
		o.SocialMedia,
		o.Photo,
		o.Published,
		openingJSON,
		o.EditKey, // WHERE clause
	)
	if err != nil {
		return err
	}
	fmt.Printf("Updated offspace %s (%d)\n", o.EditKey, o.Id)
	return nil
}

func (o OpeningDay) Value() (driver.Value, error) {
	return json.Marshal(o) // returns []byte containing JSON
}

func (o *OpeningDay) Scan(src any) error {
	if src == nil {
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("openingTimes.Scan: expected []byte, got %T", src)
	}

	return json.Unmarshal(b, o)
}
