package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"image"
	"image/jpeg"
	"regexp"
	"strings"
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

func createOffspace(offspace OffspaceRest) error {
	imgBytes, err := decodeBase64Image(offspace.Photo)
	if err != nil || len(imgBytes) > maxFileSize || len(imgBytes) == 0 || !isJPEG(imgBytes) {
	}
	if len(imgBytes) > maxFileSize {
		return errors.New("image exceeds size limit")
	}

	if !isJPEG(imgBytes) {
		return errors.New("image is not valid JPEG")

	}

	// Optional: Try decoding to ensure it's a valid JPEG image
	_, err = jpeg.Decode(strings.NewReader(string(imgBytes)))
	if err != nil {
		return errors.New("invalid JPEG image")
	}
	return dbAdapter.createOffspace(offspace)
}

func updateOffspace(offspace Offspace, password string) error {
	return dbAdapter.updateOffspace(offspace, password == *adminPassword)
}

func deleteOffspace() {
	//todo
}

func getOffspaces(checkPublished bool, tags []int) ([]OffspaceRest, error) {
	offspaces, err := dbAdapter.queryOffspaces(checkPublished, tags)
	return mapOffspaceToRestArray(offspaces), err
}

func getTags(checkPublished bool) ([]TagRest, error) {
	tags, err := dbAdapter.queryTags(checkPublished)
	return mapTagToRestArray(tags), err
}

func mapOffspaceToRestArray(offspace []Offspace) []OffspaceRest {
	offspaces := make([]OffspaceRest, len(offspace))
	for i := 0; i < len(offspace); i++ {
		offspaces[i] = mapOffspaceToRest(offspace[i])
	}
	return offspaces
}

func mapTagToRestArray(tag []Tag) []TagRest {
	tags := make([]TagRest, len(tag))
	for i := 0; i < len(tag); i++ {
		tags[i] = mapTagToRest(tag[i])
	}
	return tags
}

func mapTagToRest(tag Tag) TagRest {
	return TagRest{
		Id:        tag.Id,
		Name:      tag.Name,
		IsCity:    tag.IsCity,
		Published: tag.Published,
	}
}

func mapOffspaceToRest(offspace Offspace) OffspaceRest {
	return OffspaceRest{
		Id:          offspace.Id,
		Name:        offspace.Name,
		Bio:         offspace.Bio,
		Street:      offspace.Street,
		Postcode:    offspace.Postcode,
		City:        mapTagToRest(offspace.City),
		Website:     offspace.Website,
		SocialMedia: offspace.SocialMedia,
		Photo:       offspace.Photo,
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
