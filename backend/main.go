package main

import (
	"flag"
	_ "github.com/go-sql-driver/mysql"
)

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
	return dbAdapter.createOffspace(offspace)
}

func updateOffspace(offspace Offspace, password string) error {
	return dbAdapter.updateOffspace(offspace, password == *adminPassword)
}

func deleteOffspace() {
	//todo
}

func getOffspaces(checkPublished bool) ([]OffspaceRest, error) {
	offspaces, err := dbAdapter.queryOffspaces(checkPublished)
	return mapOffspaceToRestArray(offspaces), err
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
		Id:          offspace.Id,
		Name:        offspace.Name,
		Street:      offspace.Street,
		Postcode:    offspace.Postcode,
		City:        offspace.City,
		Website:     offspace.Website,
		SocialMedia: offspace.SocialMedia,
		Photo:       offspace.Photo,
	}
}
