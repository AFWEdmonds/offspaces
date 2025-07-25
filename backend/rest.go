package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

type OffspaceRest struct {
	Id          int64
	Name        string
	Bio         string
	Street      string
	Postcode    string
	City        string
	Website     string
	SocialMedia string
	Photo       string
}

func (o OffspaceRest) String() string {
	return fmt.Sprintf("%d, %s, %s, %s, %s, %s, %s, %s, %s", o.Id, o.Name, o.Bio, o.Street, o.Postcode, o.City, o.Website, o.SocialMedia, o.Photo)
}

func startServer() {
	http.HandleFunc("/", getRoot)
	http.HandleFunc("/publish/", getPublish)
	http.HandleFunc("/create/", postOffspace)
	http.HandleFunc("/update/", putOffspace)
	err2 := http.ListenAndServe(":3333", nil)
	if errors.Is(err2, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err2 != nil {
		fmt.Printf("error starting server: %s\n", err2)
		os.Exit(1)
	}
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	fmt.Printf("got / get request\n")
	var offspaces []OffspaceRest
	offspaces, err := getOffspaces(true)
	if err != nil {
		fmt.Errorf("read error: %v", err)
		io.WriteString(w, fmt.Sprintf("read error: %s", err))
		return
	}
	response, err := json.Marshal(offspaces)
	if err != nil {
		fmt.Errorf("read error: %v", err)
		io.WriteString(w, fmt.Sprintf("read error: %s", err))
		return
	}
	io.WriteString(w, string(response))

}

func getPublish(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	fmt.Printf("got /publish get request\n")
	var offspaces []OffspaceRest
	offspaces, err := getOffspaces(false)
	if err != nil {
		fmt.Errorf("read error: %v", err)
		io.WriteString(w, fmt.Sprintf("read error: %s", err))
		return
	}
	for _, offspace := range offspaces {
		io.WriteString(w, offspace.String())
	}
	w.Header().Add("Content-Type", "application/json")
}

func postOffspace(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	data := OffspaceRest{}
	fmt.Printf("got post request\n")
	body, err := io.ReadAll(r.Body)
	if err == nil {
		err = json.Unmarshal(body, &data)
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		if createOffspace(data) != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func putOffspace(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	data := Offspace{}
	fmt.Printf("got put request\n")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		err = json.Unmarshal(body, &data)
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		err = updateOffspace(data, r.URL.Query().Get("password"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:63342")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}
