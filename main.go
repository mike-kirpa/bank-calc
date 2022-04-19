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
	rtr.HandleFunc("/create", create).Methods("GET")
	rtr.HandleFunc("/save", save).Methods("POST")
	rtr.HandleFunc("/bank/{id:[0-9]+}", showBank).Methods("GET")
	rtr.HandleFunc("/delete", deleteBank).Methods("POST")
	http.Handle("/", rtr)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")
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

	err = t.ExecuteTemplate(w, "index", banks)
	if err != nil {
		log.Fatalln(err)
	}
}

func create(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/create.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprint(w, err.Error())
	}
	err = t.ExecuteTemplate(w, "create", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func save(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	interest_rate := r.FormValue("interest_rate")
	maximum_loan := r.FormValue("maximum_loan")
	minimum_down_payment := r.FormValue("minimum_down_payment")
	loan_term := r.FormValue("loan_term")

	if name == "" || interest_rate == "" || maximum_loan == "" || minimum_down_payment == "" || loan_term == "" {
		http.Redirect(w, r, "/create", http.StatusSeeOther)
	}
	db, err := sql.Open("mysql", "root:test@tcp(127.0.0.1:3310)/rest_api")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	insert, err := db.Query(fmt.Sprintf("INSERT INTO `banks` (`bank_name`, `interest_rate`, `maximum_loan`, `minimum_down_payment`, `loan_term`) VALUES ('%s', '%s', '%s', '%s', '%s')",
		name, interest_rate, maximum_loan, minimum_down_payment, loan_term))
	if err != nil {
		panic(err)
	}
	defer insert.Close()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func showBank(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	t, err := template.ParseFiles("templates/show.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprint(w, err.Error())
	}

	db, err := sql.Open("mysql", "root:test@tcp(127.0.0.1:3310)/rest_api")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	res, err := db.Query(fmt.Sprintf("SELECT * FROM `banks` WHERE `id` = '%s'", vars["id"]))
	if err != nil {
		panic(err)
	}

	show := Bank{}
	for res.Next() {
		var b Bank
		err := res.Scan(&b.Id, &b.BankName, &b.InterestRate, &b.MaximumLoan, &b.MinimumDownPayment, &b.LoanTerm)
		if err != nil {
			panic(err)
		}
		show = b
	}

	t.ExecuteTemplate(w, "show", show)
}

func deleteBank(w http.ResponseWriter, r *http.Request) {
	// once again, we will need to parse the path parameters
	//vars := mux.Vars(r)
	bankId := r.FormValue("bankId")

	db, err := sql.Open("mysql", "root:test@tcp(127.0.0.1:3310)/rest_api")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	res, err := db.Query(fmt.Sprintf("DELETE FROM `banks` WHERE `id` = '%s'", bankId))
	if err != nil {
		panic(err)
	}
	defer res.Close()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
