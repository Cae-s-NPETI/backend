package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
)

const PORT = 21803

func main() {
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/etia1tripmanagement")

	// handle error
	if err != nil {
		panic(err.Error())
	} else {
		log.Println("Database opened")
	}

	header := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})

	log.Printf("Listening at http://localhost:%v", PORT)
	err = http.ListenAndServe(fmt.Sprintf(":%v", PORT), handlers.CORS(header, methods, origins)(registerEndpoints(db)))

	// Shouldn't get here
	db.Close()
	log.Fatal(err)
}
