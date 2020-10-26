package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/gorilla/mux"
)

type book struct {
	Isbn   string `json:"isbn"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

// database facade
var db map[string]book

// ---------------------------------------------------------
// GET /
// ---------------------------------------------------------
func indexPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome...")
}

type byTitle []book

func (m byTitle) Len() int           { return len(m) }
func (m byTitle) Less(i, j int) bool { return m[i].Title < m[j].Title }
func (m byTitle) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

// --------------------------------------------------------
// GET /books : return all books
// --------------------------------------------------------
func getBooks(w http.ResponseWriter, r *http.Request) {
	s := []book{}
	for _, value := range db {
		s = append(s, value)
	}
	sort.Sort(byTitle(s))

	var out []byte
	out, err := json.Marshal(s)
	if err != nil {
		fmt.Println("error")
	}

	returnJSONResponse(w, out)
}

func returnJSONResponse(w http.ResponseWriter, out []byte) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(out))
}

// -----------------------------------------------------------------
// GET /books/{id}
// -----------------------------------------------------------------
func getBook(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    mybook, ok := db[vars["id"]]
    if !ok {
	w.WriteHeader(http.StatusNotFound)
	return
     }
     // converrt to json
     var out []byte
     out, err := json.Marshal(mybook)
     if err != nil {
	fmt.Println("error")
     }

     returnJSONResponse(w, out)
}

// -----------------------------------------------------
// DELETE /books/{id}
// -----------------------------------------------------
func deleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mybook, ok := db[vars["id"]]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	delete(db, mybook.Isbn)
	w.WriteHeader(http.StatusOK)
}

// -------------------------------------------------------
// POST /books : add a book
// -------------------------------------------------------
func addBook(w http.ResponseWriter, r *http.Request) {

	var mybook book

	// convert posted json to struct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&mybook); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// check input
	err := verify(&mybook)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// add to database
	db[mybook.Isbn] = mybook

	// convert back to json
	var out []byte
	out, err2 := json.Marshal(mybook)
	if err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(out))

}

func verify(mybook *book) error {
	if mybook.Isbn == "" {
		return errors.New("Missing Isbn")
	}
	_, ok := db[mybook.Isbn]
	if ok {
		return errors.New("Duplicate Isbn")
	}
	return nil
}

// ----------------------------------------------------
// PUT /books : upate a book
// ---------------------------------------------------
func updateBook(w http.ResponseWriter, r *http.Request) {

	var mybook book

	// convert posted json to struct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&mybook); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	_, ok := db[mybook.Isbn]
	if !ok {
		// return response indicating invalid book
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db[mybook.Isbn] = mybook

	w.WriteHeader(http.StatusOK)

}

func runServer() {
	r := mux.NewRouter()
	r.HandleFunc("/", indexPage)
	r.HandleFunc("/books", getBooks).Methods("GET")
	r.HandleFunc("/books/{id}", getBook).Methods("GET")
	r.HandleFunc("/books/{id}", deleteBook).Methods("DELETE")
	r.HandleFunc("/books", addBook).Methods("POST")
	r.HandleFunc("/books", updateBook).Methods("PUT")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", r))
}

func main() {
	db = make(map[string]book)
	db["1"] = book{Isbn: "1", Title: "Star Wars", Author: "George Lucas"}
	db["2"] = book{Isbn: "2", Title: "The Empire Strikes Back", Author: "George Lucas"}
	db["3"] = book{Isbn: "3", Title: "Return Of The Jedi", Author: "George Lucas"}
	runServer()
}
