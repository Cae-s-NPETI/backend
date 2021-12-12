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

type CreateTripInfo struct {
	// Note: normally this would be retrieved or verified from an authentication
	// service based on the client's auth token, but for the sake of simplicity
	// we'll trust this info directly from the client
	PassengerId int64  `json:"passengerId"`
	PostalCode  string `json:"postalCode"`
}

type CreateTripResponse struct {
	Id int64 `json:"id"`
}

func createTrip(w http.ResponseWriter, r *http.Request) {
	var info CreateTripInfo
	if ensureJson(w, r, &info) != nil {
		return
	}

	log.Println(info)

	// Get random available driver
	var driverId int64
	err := db.QueryRow(`
		SELECT driverId FROM available_driver
		ORDER BY RAND()
		LIMIT 1
	`).Scan(&driverId)
	if err != nil {
		writeError(w, r, "No available driver for your trip.")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Insert ongoing_trip table
	stmt, err := db.Prepare(`
		INSERT INTO
		ongoing_trip (postalCode, passengerId, driverId)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		writeError(w, r, "DB err 1")
		return
	}

	res, err := stmt.Exec(info.PostalCode, info.PassengerId, driverId)
	if err != nil {
		writeError(w, r, "DB err 2")
		log.Println("createTrip: Error in exec" + err.Error())
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		writeError(w, r, "DB err 3")
		log.Println("createTrip: Error in last id" + err.Error())
		return
	}

	json.NewEncoder(w).Encode(CreateTripResponse{
		Id: id,
	})
}

type AcceptTripInfo struct {
	// Note: normally this would be retrieved or verified from an authentication
	// service based on the client's auth token, but for the sake of simplicity
	// we'll trust this info directly from the client
	DriverId int64 `json:"driverId"`
}

type AcceptTripResponse struct {
	StartTime int64 `json:"startTime"`
}

// 1. Sets driver not available
// 2. Add start timestamp
func acceptTrip(w http.ResponseWriter, r *http.Request) {
	tripReqId := mux.Vars(r)["id"]

	var info AcceptTripInfo
	if ensureJson(w, r, &info) != nil {
		return
	}

	log.Println(info)

	// Delete driver from available
	stmt, err := db.Prepare(`
		DELETE FROM available_driver
		WHERE driverId = ?
	`)
	if err != nil {
		writeError(w, r, "DB err 1")
		return
	}

	res, err := stmt.Exec(info.DriverId)
	if err != nil {
		writeError(w, r, "DB err 2")
		log.Println("acceptTrip: Error in exec" + err.Error())
		return
	}

	// Update start time in ongoing_trip table
	stmt, err = db.Prepare(`
		UPDATE ongoing_trip
		SET startTime = ?
		WHERE id = ?
	`)
	if err != nil {
		writeError(w, r, "DB err 1")
		return
	}

	timestamp := time.Now().Unix()
	res, err = stmt.Exec(timestamp, tripReqId)
	if err != nil {
		writeError(w, r, "DB err 2")
		log.Println("createTrip: Error in exec" + err.Error())
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		writeError(w, r, "DB err 3")
		return
	}

	// Check if any rows were updated
	if count == 0 {
		writeError(w, r, "No records were updated")
		w.WriteHeader(http.StatusNotModified)
		return
	}

	json.NewEncoder(w).Encode(AcceptTripResponse{
		StartTime: timestamp,
	})
}

type EndTripInfo struct {
	// Note: normally this would be retrieved or verified from an authentication
	// service based on the client's auth token, but for the sake of simplicity
	// we'll trust this info directly from the client
	DriverId int64 `json:"driverId"`
}

// Describes the structure of a trip history record
type TripHistoryInfo struct {
	Id          int64
	PostalCode  string
	PassengerId int64
	DriverId    int64
	StartTime   int64
	EndTime     int64
}

// 1. Sets driver available
// 2. Rmv ongoing_trip record
// 3. Call tripHistory to archive trip
func endTrip(w http.ResponseWriter, r *http.Request) {
	tripReqId := mux.Vars(r)["id"]

	var info EndTripInfo
	if ensureJson(w, r, &info) != nil {
		return
	}

	// 3.1. Save info
	var tripHist TripHistoryInfo

	stmt, err := db.Prepare(`
		SELECT id, postalCode, passengerId, driverId
		FROM ongoing_trip
		WHERE id = ?
	`)
	if err != nil {
		writeError(w, r, "DB err 1")
		return
	}
	err = stmt.QueryRow(tripReqId).Scan(
		&tripHist.Id, &tripHist.PostalCode,
		&tripHist.PassengerId, &tripHist.DriverId,
	)
	if err != nil {
		writeError(w, r, "Trip not found: "+tripReqId)
		return
	}

	// 1. Sets driver available
	stmt, err = db.Prepare(`
		INSERT INTO
		available_driver (driverId)
		VALUES (?)
		ON DUPLICATE KEY UPDATE
	`)
	if err != nil {
		writeError(w, r, "DB err 2")
		return
	}
	_, err = stmt.Exec(info.DriverId)
	if err != nil {
		writeError(w, r, "DB err 3")
		log.Println("endTrip: Error in exec" + err.Error())
		return
	}

	// 2. Rmv ongoing_trip record
	stmt, err = db.Prepare(`
		DELETE FROM ongoing_trip
		WHERE id = ?
	`)
	if err != nil {
		writeError(w, r, "DB err 3")
		return
	}
	_, err = stmt.Exec(tripReqId)
	if err != nil {
		writeError(w, r, "DB err 4")
		log.Println("endTrip: Error in exec" + err.Error())
		return
	}

	// TODO:
	// 3. Call tripHistory to archive trip

}

// --------------

type GetDriverTripResponse struct {
	TripId      int64  `json:"tripId"`
	PostalCode  string `json:"postalCode"`
	PassengerId int64  `json:"passengerId"`
}

func getDriverTrip(w http.ResponseWriter, r *http.Request) {
	var resp GetDriverTripResponse
	reqId := mux.Vars(r)["id"]

	stmt, err := db.Prepare(`SELECT
		id, postalCode, passengerId
		FROM ongoing_trip
		WHERE driverId = ?`)
	if err != nil {
		writeError(w, r, "DB err 1")
		return
	}

	err = stmt.QueryRow(reqId).Scan(&resp.TripId, &resp.PostalCode, &resp.PassengerId)
	if err != nil {
		writeError(w, r, "Driver is not assigned to any trip: "+reqId)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

type SetAvailableDriverInfo struct {
	// Note: normally this would be retrieved or verified from an authentication
	// service based on the client's auth token, but for the sake of simplicity
	// we'll trust this info directly from the client
	DriverId int64 `json:"driverId"`
}

func setAvailableDriver(w http.ResponseWriter, r *http.Request) {
	var info SetAvailableDriverInfo
	if ensureJson(w, r, &info) != nil {
		return
	}

	log.Println(info)

	// Check if driver found

	stmt, err := db.Prepare(`
		SELECT driverId FROM available_driver
		WHERE driverId != ?
	`)
	if err != nil {
		writeError(w, r, "DB err 1")
		return
	}

	var driverId int64
	err = stmt.QueryRow(info.DriverId).Scan(&driverId)
	if err == nil {
		// driver found, do nothing
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Insert  table
	stmt, err = db.Prepare(`
		INSERT INTO
		available_driver (driverId)
		VALUES (?)
	`)
	if err != nil {
		writeError(w, r, "DB err 1")
		return
	}

	_, err = stmt.Exec(info.DriverId)
	if err != nil {
		writeError(w, r, "DB err 2")
		log.Println("setAvailableDriver: Error in exec" + err.Error())
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

	// Adds a trip request and assigns an available driver to it
	router.HandleFunc("/api/v1/trips", createTrip).Methods("POST")
	// Starts a trip
	router.HandleFunc("/api/v1/trips/{id}", acceptTrip).Methods("POST")
	// Ends a trip
	router.HandleFunc("/api/v1/trips/{id}", endTrip).Methods("DELETE")

	// Gets assigned trip for the driver
	router.HandleFunc("/api/v1/driver/{id}", getDriverTrip).Methods("GET")
	// Sets driver as available
	router.HandleFunc("/api/v1/driver", setAvailableDriver).Methods("POST")

	return router
}
