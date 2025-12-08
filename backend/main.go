package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var maxFileSize = 1 * 1024 * 1024 // 1MB
var adminPassword *string

func main() {
	adminPassword = flag.String("adminpass", "password", "admin password for offspace editing")
	dbUser := flag.String("dbuser", "user", "db username ")
	dbPassword := flag.String("dbpass", "password", "db password ")
	flag.Parse()
	connectDb(dbUser, dbPassword)
	startServer()
}

func createOffspace(offspace OffspaceRest) (string, error) {
	imgBytes, err := decodeBase64Image(offspace.Photo)
	if err != nil || len(imgBytes) > maxFileSize || len(imgBytes) == 0 || !isJPEG(imgBytes) {
	}
	if len(imgBytes) > maxFileSize {
		return "", errors.New("image exceeds size limit")
	}

	if !isJPEG(imgBytes) {
		return "", errors.New("image is not valid JPEG")

	}

	// Optional: Try decoding to ensure it's a valid JPEG image
	_, err = jpeg.Decode(strings.NewReader(string(imgBytes)))
	if err != nil {
		return "", errors.New("invalid JPEG image")
	}
	return dbAdapter.createOffspace(offspace)
}

func updateOffspace(offspace Offspace, password string) error {
	return dbAdapter.updateOffspace(offspace, password == *adminPassword)
}

func deleteOffspace() {
	//todo
}

func getOffspaces(params QueryRest) ([]OffspaceRest, error) {
	var err error
	var offspaces []Offspace
	if params.AdminKey == *adminPassword {
		offspaces, err = dbAdapter.queryOffspaces(true, mapQueryRestToQuery(params))
	} else {
		offspaces, err = dbAdapter.queryOffspaces(false, mapQueryRestToQuery(params))
	}
	return mapOffspaceToRestArray(offspaces), err
}

func getOffspaceByKey(key string) (OffspaceRest, error) {
	offspace, err := dbAdapter.getOffspaceByKey(key)
	return mapOffspaceToRest(offspace), err
}

func mapOffspaceToRestArray(offspace []Offspace) []OffspaceRest {
	offspaces := make([]OffspaceRest, len(offspace))
	for i := 0; i < len(offspace); i++ {
		offspaces[i] = mapOffspaceToRest(offspace[i])
	}
	return offspaces
}

func mapOffspaceToRest(offspace Offspace) OffspaceRest {
	return OffspaceRest{
		ID:          offspace.Id,
		Name:        offspace.Name,
		Street:      offspace.Street,
		Postcode:    offspace.Postcode,
		City:        offspace.City,
		Website:     offspace.Website,
		SocialMedia: offspace.SocialMedia,
		Photo:       offspace.Photo,
		Opening:     offspace.OpeningTimes,
	}
}

func mapQueryRestToQuery(query QueryRest) Query {
	return Query{
		Text:           query.Text,
		Index:          query.Index,
		DisplayAmount:  query.DisplayAmount,
		RequireOpenNow: query.RequireOpenNow,
		RequireExhibOn: query.RequireExhibOn,
		SearchName:     query.SearchName,
		SearchAddress:  query.SearchAddress,
		SearchExhib:    query.SearchExhib,
		SortBy:         query.SortBy,
	}
}

func isJPEG(imgBytes []byte) bool {
	_, format, err := image.DecodeConfig(strings.NewReader(string(imgBytes)))
	return err == nil && format == "jpeg"
}

func decodeBase64Image(data string) ([]byte, error) {
	// Remove the data:image/jpeg;base64, header
	re := regexp.MustCompile(`^data:image\/jpeg;base64,`)
	if !re.MatchString(data) {
		return nil, fmt.Errorf("image must be base64-encoded JPEG")
	}
	base64Data := re.ReplaceAllString(data, "")
	return base64.StdEncoding.DecodeString(base64Data)
}
