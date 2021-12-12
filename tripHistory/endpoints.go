package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var db *sql.DB

// --------------
// Structures and common function
// --------------

// A regular REST JSON response.
type RegularResponse struct {
	Status      bool   `json:"status"`
	Description string `json:"description"`
}

// Writes a regular JSON error response
func writeError(w http.ResponseWriter, r *http.Request, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(RegularResponse{
		Status:      false,
		Description: description,
	})
}

// Ensures that the request is a json and converts it
func ensureJson(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if r.Header.Get("Content-type") != "application/json" {
		writeError(w, r, "Expected Content-type = application/json")
		return errors.New("Expected Content-type = application/json")
	}

	// read the string sent to the service
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, r, "Could not read request body")
		return err
	}

	// convert JSON to object
	err = json.Unmarshal(reqBody, &v)
	if err != nil {
		writeError(w, r, "Could not decode request JSON: "+err.Error())
		return err
	}

	// All good
	return nil
}

// --------------
// Main endpoint callbacks
// --------------

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world!")
}

// --------------

// Describes the structure of a trip history add & get request, which almost
// directly correspond to the actual DB table structure
type TripHistoryInfo struct {
	Id          int64  `json:"id"`
	PostalCode  string `json:"postalCode"`
	PassengerId int64  `json:"passengerId"`
	DriverId    int64  `json:"driverId"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
}

type GetPasssengerTripsResponse struct {
	Trips []TripHistoryInfo `json:"trips"`
}

func getPasssengerTrips(w http.ResponseWriter, r *http.Request) {
	reqPassengerId := mux.Vars(r)["passengerId"]

	stmt, err := db.Prepare(`SELECT
		id, postalCode, passengerId, driverId, startTime, endTIme
		FROM trip_history
		WHERE passengerId = ?
		ORDER BY startTime DESC`)
	if err != nil {
		writeError(w, r, "DB err 1")
		return
	}
	rows, err := stmt.Query(reqPassengerId)
	if err != nil {
		writeError(w, r, "DB err 2")
		return
	}

	// Save a list of trips made, latest startTime first
	var resp GetPasssengerTripsResponse
	for rows.Next() {
		var info TripHistoryInfo
		err = rows.Scan(
			&info.Id, &info.PostalCode, &info.PassengerId,
			&info.DriverId, &info.StartTime, &info.EndTime,
		)
		if err != nil {
			writeError(w, r, "DB err 3")
			return
		}
		resp.Trips = append(resp.Trips, info)
	}

	json.NewEncoder(w).Encode(resp)

}

type AddTripResponse struct {
	EndTime int64 `json:"endTime"`
}

func addTripLog(w http.ResponseWriter, r *http.Request) {
	var info TripHistoryInfo
	if ensureJson(w, r, &info) != nil {
		return
	}

	timestamp := time.Now().Unix()

	stmt, err := db.Prepare(`INSERT INTO trip_history
		(id, postalCode, passengerId, driverId, startTime, endTIme)
		VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		writeError(w, r, "DB err 1")
		return
	}
	_, err = stmt.Exec(info.Id, info.PostalCode, info.PassengerId,
		info.DriverId, info.StartTime, timestamp)
	if err != nil {
		writeError(w, r, "DB err 2")
		log.Println("addTripLog: Error in exec" + err.Error())
		return
	}

	json.NewEncoder(w).Encode(AddTripResponse{
		EndTime: timestamp,
	})

}

// --------------
// Main endpoint registry
// --------------

func registerEndpoints(db1 *sql.DB) *mux.Router {
	// Set db object
	db = db1

	// Register routes
	router := mux.NewRouter()

	router.HandleFunc("/api/v1", home)

	// Gets all history of trips of a passenger
	router.HandleFunc("/api/v1/passengerTrips/{passengerId}", getPasssengerTrips).Methods("GET")
	// Adds a new trip log
	// TODO: This could be an RPC call instead.
	router.HandleFunc("/api/v1/tripsLog", addTripLog).Methods("POST")

	return router
}
