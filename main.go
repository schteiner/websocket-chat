//main.go
package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"src/chat/wschat/ws"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var database *sql.DB

func SignInPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fmt.Fprintln(w, "not POST")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	row, err := database.Query("SELECT * FROM TBwsuser WHERE username = ?", username)
	if err != nil {
		log.Println(err)
	}

	// проверка на сущестыование пользователя
	if row.Next() != false {
		fmt.Println("user exist")

		http.ServeFile(w, r, "SignInPageWithAlert.html")
		return
	}

	_, err = database.Exec("insert into DBwschat.TBwsuser(username, password) values (?, ?)", username, password)
	if err != nil {
		log.Println(err)
	}
	http.ServeFile(w, r, "LogInPage.html")
}

type User struct {
	Id       int
	Username string
	Password string
}

var u = User{}

type ViewData struct {
	Id string
}

func ChatPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fmt.Fprintln(w, "not POST chat")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	row, err := database.Query("SELECT * FROM TBwsuser WHERE username = ?", username)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer row.Close()

	for row.Next() {
		err := row.Scan(&u.Id, &u.Username, &u.Password)
		if err != nil {
			fmt.Println(err)
			continue
		}

	}
	if password != u.Password {
		fmt.Fprintln(w, "invalid password")
		return
	}
	//	fmt.Println("userid:", u.Id)

	//websocket
	data := ViewData{
		Id: strconv.Itoa(u.Id),
	}

	tmpl, _ := template.ParseFiles("chat.html")
	tmpl.Execute(w, data)

}

func main() {
	// conecting to database
	db, err := sql.Open("mysql", "root:password")
	if err != nil {
		log.Println(err)
	}
	database = db
	defer db.Close()

	//websocket
	hub := ws.NewHub()
	go hub.HubRun()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println("user id is", strconv.Itoa(u.Id))
		ws.ServeWS(w, r, hub, strconv.Itoa(u.Id), u.Username)
	})

	http.HandleFunc("/signIn", SignInPage)
	http.HandleFunc("/SignInPage.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "SignInPage.html")
	})
	http.HandleFunc("/chat", ChatPage)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "LogInPage.html")
	})

	fmt.Println("starting server at :8080")
	err = http.ListenAndServeTLS(":8080", "server.crt", "server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
