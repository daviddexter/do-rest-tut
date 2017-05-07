package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"encoding/json"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DBNAME = "app.db"
	LOGS   = "server.log"
)

var DB *sql.DB

type Person struct {
	Id        string `json:id`
	Firstname string `json:firstname`
	Lastname  string `json:lastname`
	Position  Job    `json:job,omitempty`
}

type Job struct {
	Title  string `json:title`
	Salary int32  `json:salary,omitempty`
}

type DataObj struct {
	Data Person
}

type ResponseBox struct {
	Hint string `json:hint`
}

func rowsCount() (count int) {
	rows, err := DB.Query("SELECT COUNT(*) as count FROM people")
	errorChecker(err)
	for rows.Next() {
		rows.Scan(&count)
	}
	return count
}
func fetchAllHandler(w http.ResponseWriter, r *http.Request) {
	if numRows := rowsCount(); numRows <= 0 {
		d, _ := json.MarshalIndent(ResponseBox{Hint: "Empty Database"}, "", " ")
		fmt.Fprint(w, string(d))
	} else {
		rows, err := DB.Query("SELECT * FROM people")
		errorChecker(err)
		var result []DataObj
		for rows.Next() {
			item := DataObj{}
			err := rows.Scan(&item.Data.Id, &item.Data.Firstname, &item.Data.Lastname, &item.Data.Position.Title, &item.Data.Position.Salary)
			errorChecker(err)
			result = append(result, item)
		}
		d, _ := json.MarshalIndent(result, "", " ")
		fmt.Fprint(w, string(d))
	}

}

func newHandler(w http.ResponseWriter, r *http.Request) {
	items := mux.Vars(r)
	stmt, err := DB.Prepare("INSERT INTO people (firstname,lastname,title,salary) VALUES(?,?,?,?)")
	errorChecker(err)
	stmt.Exec(items["firstname"], items["lastname"], items["title"], items["salary"])
	d, _ := json.MarshalIndent(ResponseBox{Hint: "New person added"}, "", " ")
	fmt.Fprint(w, string(d))
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	items := mux.Vars(r)
	stmt, err := DB.Prepare("UPDATE people SET firstname=?,lastname=?,title=?,salary=? WHERE id=?")
	errorChecker(err)
	stmt.Exec(items["firstname"], items["lastname"], items["title"], items["salary"], items["id"])
	d, _ := json.MarshalIndent(ResponseBox{Hint: "Person updated"}, "", " ")
	fmt.Fprint(w, d)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	items := mux.Vars(r)
	var result []DataObj
	item := DataObj{}
	err := DB.QueryRow("SELECT * FROM people WHERE id=?", items["id"]).Scan(&item.Data.Id, &item.Data.Firstname, &item.Data.Lastname, &item.Data.Position.Title, &item.Data.Position.Salary)
	switch {
	case err == sql.ErrNoRows:
		d, _ := json.MarshalIndent(ResponseBox{Hint: "No user with that ID."}, "", " ")
		fmt.Fprint(w, string(d))
	case err != nil:
		errorChecker(err)
	default:
		result = append(result, item)
		d, _ := json.MarshalIndent(result, "", " ")
		fmt.Fprint(w, string(d))
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	items := mux.Vars(r)
	stmt, err := DB.Prepare("DELETE FROM people WHERE id=?")
	errorChecker(err)
	stmt.Exec(items["id"])
	d, _ := json.MarshalIndent(ResponseBox{Hint: "Person deleted"}, "", " ")
	fmt.Fprint(w, string(d))
}

func errorChecker(e error) {
	if e != nil {
		panic(e)
	}
}

func init() {
	db, err := sql.Open("sqlite3", DBNAME)
	errorChecker(err)
	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS people (id INTEGER PRIMARY KEY AUTOINCREMENT,firstname TEXT,lastname TEXT,title TEXT,salary INTERGER)")
	errorChecker(err)
	stmt.Exec()
	DB = db

}

func main() {
	router := mux.NewRouter()
	indexH := http.HandlerFunc(fetchAllHandler)
	newH := http.HandlerFunc(newHandler)
	updateH := http.HandlerFunc(updateHandler)
	getH := http.HandlerFunc(getHandler)
	deleteH := http.HandlerFunc(deleteHandler)
	logFile, err := os.OpenFile(LOGS, os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0666)
	errorChecker(err)
	router.Handle("/all", handlers.LoggingHandler(logFile, indexH))
	router.Handle("/new/{firstname}/{lastname}/{title}/{salary}", handlers.LoggingHandler(logFile, newH))
	router.Handle("/update/{id}/{firstname}/{lastname}/{title}/{salary}", handlers.LoggingHandler(logFile, updateH))
	router.Handle("/get/{id}", handlers.LoggingHandler(logFile, getH))
	router.Handle("/delete/{id}", handlers.LoggingHandler(logFile, deleteH))
	server := http.Server{
		Addr:    ":8000",
		Handler: router,
	}
	fmt.Println("Listening at ...")
	server.ListenAndServe()
}
