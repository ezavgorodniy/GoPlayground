package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

func allArticles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("AllArticles Endpoint Hit")

	db, err := sql.Open("mysql", "<dbUser>:<dbPassword>@tcp(<ip>)/<startSchema")

	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	results, err := db.Query("SELECT id, username FROM accounts")

	if err != nil {
		panic(err.Error())
	}

	var users []User
	for results.Next() {
		var user User

		err = results.Scan(&user.Id, &user.Username)
		if err != nil {
			panic(err.Error())
		}

		users = append(users, user)
	}

	json.NewEncoder(w).Encode(users)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage Endpoint Hit")
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/all", allArticles)

	log.Fatal(http.ListenAndServe(":8081", myRouter))
}

func main() {
	handleRequests()
}
