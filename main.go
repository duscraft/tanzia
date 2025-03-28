package main

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"html/template"
	"log"
	"net/http"
	"os"
)

func getData() Person {
	person := Person{
		Name:     "World",
		Tantieme: 1,
	}

	return person
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	data := getData()
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, data)
}

func personHandler(w http.ResponseWriter, r *http.Request) {

}

func billsHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, _ := sql.Open("sqlite3", "./tantieme.db")
	defer db.Close()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/persons", personHandler)
	http.HandleFunc("/bills", billsHandler)

	fmt.Fprint(os.Stdout, "Listening on port ", os.Getenv("PORT"), "...")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil))
}
