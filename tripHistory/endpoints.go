package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

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

// Describes the structure of a trip history add request, which almost directly
// correspond to the actual DB table structure
type TripHistoryInfo struct {
	Id          int64  `json:"id"`
	PostalCode  string `json:"postalCode"`
	PassengerId int64  `json:passengerId"`
	DriverId    int64  `json:"driverId"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
}

// 1. Sets driver available
// 2. Rmv ongoing_trip record
// 3. Call tripHistory to archive trip
func getPasssengerTrips(w http.ResponseWriter, r *http.Request) {
	//reqPassengerId := mux.Vars(r)["passengerId"]

	var info TripHistoryInfo
	if ensureJson(w, r, &info) != nil {
		return
	}

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
