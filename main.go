package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	_ "github.com/heroku/x/hmetrics/onload"
	"log"
	"net/http"
	"text/template"
)

type Bank struct {
	Id                 uint64
	BankName           string
	InterestRate       float32
	MaximumLoan        uint64
	MinimumDownPayment uint64
	LoanTerm           uint64
}

func main() {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/", index).Methods("GET")
	http.Handle("/", rtr)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		fmt.Fprint(w, err.Error())
	}

	db, err := sql.Open("mysql", "root:test@tcp(127.0.0.1:3310)/rest_api")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	res, err := db.Query("SELECT * FROM `banks` ")
	if err != nil {
		panic(err)
	}

	var banks []Bank
	for res.Next() {
		var b Bank
		err := res.Scan(&b.Id, &b.BankName, &b.InterestRate, &b.MaximumLoan, &b.MinimumDownPayment, &b.LoanTerm)
		if err != nil {
			panic(err)
		}
		banks = append(banks, b)
	}

	err = t.ExecuteTemplate(w, "index.html", banks)
	if err != nil {
		log.Fatalln(err)
	}
}
