package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

const PORT = 21802

func main() {
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/etia1tripmanagement")

	// handle error
	if err != nil {
		panic(err.Error())
	} else {
		log.Println("Database opened")
	}

	log.Printf("Listening at http://localhost:%v", PORT)
	err = http.ListenAndServe(fmt.Sprintf("localhost:%v", PORT), registerEndpoints(db))

	// Shouldn't get here
	db.Close()
	log.Fatal(err)
}
