package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type OffspaceRest struct {
	ID          int              `json:"id"`
	Name        string           `json:"name"`
	Street      string           `json:"street"`
	Postcode    string           `json:"postcode"`
	City        string           `json:"city"`
	Website     string           `json:"website"`
	SocialMedia string           `json:"social_media"`
	Photo       string           `json:"photo"` // base64 encoded by default if in JSON
	Published   bool             `json:"published"`
	Opening     OpeningHoursRest `json:"opening_hours"`
}

type QueryRest struct {
	Text           string `json:"text"`
	Index          int    `json:"index"`
	DisplayAmount  int    `json:"display_amount"`
	RequireOpenNow bool   `json:"require_open_now"`
	RequireShowOn  bool   `json:"require_show_on"`
	SearchName     bool   `json:"search_name"`
	SearchAddress  bool   `json:"search_address"`
	SearchShow     bool   `json:"search_show"`
	SortBy         string `json:"sort_by"`
	AdminKey       string `json:"admin_key"`
}

type OpeningHoursRest struct {
	Mon [24]bool `json:"mon"`
	Tue [24]bool `json:"tue"`
	Wed [24]bool `json:"wed"`
	Thu [24]bool `json:"thu"`
	Fri [24]bool `json:"fri"`
	Sat [24]bool `json:"sat"`
	Sun [24]bool `json:"sun"`
}

func (o OffspaceRest) String() string {
	return fmt.Sprintf("%d, %s, %s, %s, %s, %s, %s, %s, %s", o.ID, o.Name, o.Street, o.Postcode, o.City, o.Website, o.SocialMedia, o.Photo)
}

func startServer() {
	http.HandleFunc("/", getRoot)
	http.HandleFunc("/create/", postOffspace)
	http.HandleFunc("/update/", putOffspace)
	http.HandleFunc("/get/", getOffspace)
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
	offspaces, err := getOffspaces(queryToStruct(url.ParseQuery(r.URL.RawQuery)))
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

func getOffspace(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	fmt.Printf("got / get request\n")
	values, err := url.ParseQuery(r.URL.RawQuery)
	offspace, err := getOffspaceByKey(getString(values, "editKey", ""))
	if err != nil {
		fmt.Errorf("read error: %v", err)
		io.WriteString(w, fmt.Sprintf("read error: %s", err))
		return
	}
	response, err := json.Marshal(offspace)
	if err != nil {
		fmt.Errorf("read error: %v", err)
		io.WriteString(w, fmt.Sprintf("read error: %s", err))
		return
	}
	io.WriteString(w, string(response))
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

func queryToStruct(values map[string][]string, err error) QueryRest {
	return QueryRest{
		Text:           getString(values, "text", ""),
		Index:          getInt(values, "index", 0),
		DisplayAmount:  getInt(values, "displayAmount", 10),
		RequireOpenNow: getBool(values, "requireOpenNow", false),
		RequireShowOn:  getBool(values, "requireShowOn", true),
		SearchAddress:  getBool(values, "searchAddress", false),
		SearchShow:     getBool(values, "searchShow", false),
		SortBy:         getString(values, "sortBy", "date"),
		AdminKey:       getString(values, "adminKey", ""),
	}
}

func getString(values map[string][]string, key, def string) string {
	if v, ok := values[key]; ok && len(v) > 0 {
		return v[0]
	}
	return def
}

func getBool(values map[string][]string, key string, def bool) bool {
	if v, ok := values[key]; ok && len(v) > 0 {
		val := strings.ToLower(v[0])
		return val == "true" || val == "1" || val == "on"
	}
	return def
}

func getInt(values map[string][]string, key string, def int) int {
	if v, ok := values[key]; ok && len(v) > 0 {
		if i, err := strconv.Atoi(v[0]); err == nil {
			return i
		}
	}
	return def
}
