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
	rtr.HandleFunc("/edit/{id:[0-9]+}", editBank).Methods("GET")
	rtr.HandleFunc("/update", updateBank).Methods("POST")
	rtr.HandleFunc("/delete", deleteBank).Methods("POST")
	rtr.HandleFunc("/calc", calc).Methods("GET")
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

func editBank(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	t, err := template.ParseFiles("templates/edit.html", "templates/header.html", "templates/footer.html")
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

	edit := Bank{}
	for res.Next() {
		var b Bank
		err := res.Scan(&b.Id, &b.BankName, &b.InterestRate, &b.MaximumLoan, &b.MinimumDownPayment, &b.LoanTerm)
		if err != nil {
			panic(err)
		}
		edit = b
	}

	t.ExecuteTemplate(w, "edit", edit)
}

func updateBank(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	name := r.FormValue("name")
	interestRate := r.FormValue("interest_rate")
	maximumLoan := r.FormValue("maximum_loan")
	minimumDownPayment := r.FormValue("minimum_down_payment")
	loanTerm := r.FormValue("loan_term")

	if name == "" || interestRate == "" || maximumLoan == "" || minimumDownPayment == "" || loanTerm == "" {
		http.Redirect(w, r, "/edit/"+id, http.StatusSeeOther)
	}
	db, err := sql.Open("mysql", "root:test@tcp(127.0.0.1:3310)/rest_api")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	update, err := db.Query(fmt.Sprintf("UPDATE `banks` SET `bank_name`='%s', `interest_rate`='%s', `maximum_loan`='%s', `minimum_down_payment`='%s', `loan_term`='%s' WHERE `id`='%s'",
		name, interestRate, maximumLoan, minimumDownPayment, loanTerm, id))
	if err != nil {
		panic(err)
	}
	defer update.Close()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func calc(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/calc.html", "templates/header.html", "templates/footer.html")
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

	err = t.ExecuteTemplate(w, "calc", banks)
	if err != nil {
		log.Fatalln(err)
	}
}
