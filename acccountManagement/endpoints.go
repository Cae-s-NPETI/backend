package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var db *sql.DB

// --------------
// Type mappings to the record in the db table
// --------------

type DBPassenger struct {
	Id        int
	FirstName string
	LastName  string
	MobileNo  string
	Email     string
}

type DBDriver struct {
	Id               int
	FirstName        string
	LastName         string
	MobileNo         string
	Email            string
	IdentificationNo string
	CarNo            string
}

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

type CreatePassengerInfo struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	MobileNo  string `json:"mobileNo"`
	Email     string `json:"email"`
}

type CreatePassengerResponse struct {
	Id int64 `json:"id"`
}

func createPassenger(w http.ResponseWriter, r *http.Request) {
	var info CreatePassengerInfo
	if ensureJson(w, r, &info) != nil {
		return
	}

	log.Println(info)

	// Insert user table
	stmt, err := db.Prepare("INSERT INTO `user` (`firstName`, `lastName`, `mobileNo`, `email`) VALUES (?, ?, ?, ?)")
	if err != nil {
		writeError(w, r, "DB err 1")
		log.Println("createPassenger: Error in prepare" + err.Error())
		return
	}

	res, err := stmt.Exec(info.FirstName, info.LastName, info.MobileNo, info.Email)
	if err != nil {
		writeError(w, r, "DB err 2")
		log.Println("createPassenger: Error in exec" + err.Error())
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		writeError(w, r, "DB err 3")
		log.Println("createPassenger: Error in last id" + err.Error())
		return
	}

	// Insert passenger table
	stmt, err = db.Prepare("INSERT INTO `passenger` (`userId`) VALUES (?)")
	if err != nil {
		writeError(w, r, "DB err 4")
		log.Println("createPassenger: Error in prepare" + err.Error())
		return
	}

	res, err = stmt.Exec(id)
	if err != nil {
		writeError(w, r, "DB err 5")
		log.Println("createPassenger: Error in exec" + err.Error())
		return
	}

	json.NewEncoder(w).Encode(CreatePassengerResponse{
		Id: id,
	})
}

type GetPassengerResponse struct {
	Id        int64  `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	MobileNo  string `json:"mobileNo"`
	Email     string `json:"email"`
}

func getPassenger(w http.ResponseWriter, r *http.Request) {
	var resp GetPassengerResponse
	reqId := mux.Vars(r)["id"]

	stmt, err := db.Prepare(`SELECT
		id, firstName, lastName, mobileNo, email
		FROM passenger p INNER JOIN user u on p.userId = u.id
		WHERE p.userId = ?`)
	if err != nil {
		writeError(w, r, "DB err 1")
		log.Println("getPassenger: Error in prepare" + err.Error())
		return
	}

	err = stmt.QueryRow(reqId).Scan(&resp.Id, &resp.FirstName, &resp.LastName, &resp.MobileNo, &resp.Email)
	if err != nil {
		writeError(w, r, "Id not found: "+reqId)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

type UpdatePassengerInfo struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	MobileNo  string `json:"mobileNo"`
	Email     string `json:"email"`
}

type UpdatePassengerResponse struct {
	Id int64 `json:"id"`
}

func updatePassenger(w http.ResponseWriter, r *http.Request) {
	var info CreatePassengerInfo
	if ensureJson(w, r, &info) != nil {
		return
	}

	reqId := mux.Vars(r)["id"]

	stmt, err := db.Prepare(`UPDATE user u INNER JOIN passenger p ON u.id = p.userId
		SET firstName = ?, lastName = ?, mobileNo = ?, email = ?
		WHERE u.id = ?`)
	if err != nil {
		writeError(w, r, "DB err 1")
		log.Println("getPassenger: Error in prepare" + err.Error())
		return
	}

	res, err := stmt.Exec(info.FirstName, info.LastName, info.MobileNo, info.Email, reqId)
	if err != nil {
		writeError(w, r, "DB err 2")
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
}

// --------------

type CreateDriverInfo struct {
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	MobileNo         string `json:"mobileNo"`
	Email            string `json:"email"`
	IdentificationNo string `json:"identificationNo"`
	CarNo            string `json:"carNo"`
}

type CreateDriverResponse struct {
	Id int64 `json:"id"`
}

func createDriver(w http.ResponseWriter, r *http.Request) {
	var info CreateDriverInfo
	if ensureJson(w, r, &info) != nil {
		return
	}

	log.Println(info)

	// Insert user table
	stmt, err := db.Prepare("INSERT INTO `user` (`firstName`, `lastName`, `mobileNo`, `email`) VALUES (?, ?, ?, ?)")
	if err != nil {
		writeError(w, r, "DB err 1")
		log.Println("createDriver: Error in prepare" + err.Error())
		return
	}

	res, err := stmt.Exec(info.FirstName, info.LastName, info.MobileNo, info.Email)
	if err != nil {
		writeError(w, r, "DB err 2")
		log.Println("createDriver: Error in exec" + err.Error())
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		writeError(w, r, "DB err 3")
		log.Println("createDriver: Error in last id" + err.Error())
		return
	}

	// Insert driver table
	stmt, err = db.Prepare("INSERT INTO `driver` (`userId`, `identificationNo`, `carNo`) VALUES (?, ?, ?)")
	if err != nil {
		writeError(w, r, "DB err 4")
		log.Println("createDriver: Error in prepare" + err.Error())
		return
	}

	res, err = stmt.Exec(id, info.IdentificationNo, info.CarNo)
	if err != nil {
		writeError(w, r, "DB err 5")
		log.Println("createDriver: Error in exec" + err.Error())
		return
	}

	json.NewEncoder(w).Encode(CreatePassengerResponse{
		Id: id,
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

	router.HandleFunc("/api/v1/passengers", createPassenger).Methods("POST")
	router.HandleFunc("/api/v1/passengers/{id}", getPassenger).Methods("GET")
	router.HandleFunc("/api/v1/passengers/{id}", updatePassenger).Methods("PUT")

	router.HandleFunc("/api/v1/drivers", createDriver).Methods("POST")
	router.HandleFunc("/api/v1/drivers/{id}", getDriver).Methods("GET")
	// router.HandleFunc("/api/v1/drivers/{id}", updateDriver).Methods("PUT")

	return router
}
