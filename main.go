package main

import (
	"encoding/json"
	"fmt"
	"github.com/marni/goigc"
	"google.golang.org/appengine"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var startTime time.Time
var results []int
var igcs []string
var idCount int

type Response struct {
	Id	int		`json:"id"`
}

type Post struct {
	Url string	`json:"url"`
}

type Point struct {
	Longitude	float64	`json:"longitude"`
	Latitude	float64	`json:"latitude"`
}

type IgcInfo struct {
	H_date			string	`json:"h_date"`
	Pilot			string	`json:"pilot"`
	Glider			string	`json:"glider"`
	Gider_id		string	`json:"glider_id"`
	Track_length	float64		`json:"track_length"`
}

type ApiInfo struct {
	Uptime	string	`json:"uptime"`
	Info	string	`json:"info"`
	Version	string	`json:"version"`
}

func uptime() time.Duration {
	return time.Since(startTime)
}

func init() {
	startTime = time.Now()
	idCount = 0
}
func main() {


	http.HandleFunc("/igcinfo/api", infoHandler)
	http.HandleFunc("/igcinfo/api/igc", igcHandler)
	http.HandleFunc("/igcinfo/api/igc/", igcInfoHandler)

	appengine.Main() // Om det står her så funker det

	err := http.ListenAndServe(":" +os.Getenv("PORT"), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	api := ApiInfo{uptime().String(), "Service for IGC tracks.", "v1"}
	err := json.NewEncoder(w).Encode(api)
	if err != nil {
		panic(err)
	}
}

func igcHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		var p Post
		err = json.Unmarshal(body, &p)
		if err != nil{
			panic(err)
		}
		igcs = append(igcs, p.Url)
		results = append(results, idCount)

		r := Response{idCount}
		err = json.NewEncoder(w).Encode(r)
		if err != nil {
			panic(err)
		}
		idCount++
	} else if r.Method == "GET" {
		err := json.NewEncoder(w).Encode(results)
		if err != nil {
			panic(err)
		}
	}
}

func igcInfoHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	if len(url) == 5 {
		id, err := strconv.Atoi(url[4])
		if err != nil {
			panic(err)
		}
		if id <= len(results) {
			var i IgcInfo
			data, err := igc.ParseLocation(igcs[id])
			if err != nil {
				fmt.Errorf("Problem reading the track", err)
			}
			i.H_date = data.Date.String()
			i.Pilot = data.Pilot
			i.Glider = data.GliderType
			i.Gider_id = data.GliderID
			i.Track_length = distOfTrack(data.Points)
			err = json.NewEncoder(w).Encode(i)
			if err != nil {
				panic(err)
			}
		}
	}
}


func distOfTrack(p []igc.Point) float64 {
	totaldist := 0.0
	for i := 1; i < len(p); i++ {
		totaldist += (*igc.Point).Distance(&p[i], p[i-1])
	}
	return totaldist
}
