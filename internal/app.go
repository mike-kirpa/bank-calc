package internal

import (
	"bank-calc/config"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/jackc/pgx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"text/template"
)

type Bank struct {
	Id                 uint64
	BankName           string
	InterestRate       float64
	MaximumLoan        uint64
	MinimumDownPayment uint64
	LoanTerm           uint64
}

// Application is application entry point.
type Application struct {
	cfg      *config.Config
	log      *zap.Logger
	database *sql.DB
}

func NewApplication() (app *Application) {
	app = &Application{}

	app.initConfig()
	app.initLogger()
	app.initDatabase()

	return
}

func (app *Application) initConfig() {
	var err error

	app.cfg, err = config.NewConfig()

	if err != nil {
		zap.S().Panic("Config init failed", zap.Error(err))
	}

	zap.S().Info("Configuration parsed successfully...")
}

func (app *Application) initLogger() {
	cfg := zap.NewProductionEncoderConfig()

	var writer io.Writer

	if app.cfg.LogToFileEnabled {
		writer = &lumberjack.Logger{
			Filename:   app.cfg.LogFilePath,
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   false,
		}
	} else {
		writer = os.Stderr
	}

	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder := zapcore.NewConsoleEncoder(cfg)

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(writer),
		config.GetZapLevel(app.cfg.LogLevel),
	)
	logger := zap.New(core, zap.AddCaller())

	app.log = logger.Named(config.ServiceName)
	zap.ReplaceGlobals(app.log)

	app.log.Info("Logger init...")
}

func (app *Application) initDatabase() {
	var err error

	app.database, err = sql.Open("sqlite3", "store.db")
	if err != nil {
		app.log.Fatal("Database connection failed", zap.Error(err))
	}

	app.log.Info("Database initialization successfully...")
}

// Run starts application
func (app *Application) Run() {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/", index).Methods("GET")
	rtr.HandleFunc("/create", create).Methods("GET")
	rtr.HandleFunc("/save", save).Methods("POST")
	rtr.HandleFunc("/bank/{id:[0-9]+}", showBank).Methods("GET")
	rtr.HandleFunc("/edit/{id:[0-9]+}", editBank).Methods("GET")
	rtr.HandleFunc("/update", updateBank).Methods("POST")
	rtr.HandleFunc("/delete", deleteBank).Methods("POST")
	rtr.HandleFunc("/result", result).Methods("POST")
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

	db, err := sql.Open("sqlite3", "store.db")
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
	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO banks (bank_name, interest_rate, maximum_loan, minimum_down_payment, loan_term) VALUES ($1, $2, $3, $4, $5)",
		name, interest_rate, maximum_loan, minimum_down_payment, loan_term)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func showBank(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	t, err := template.ParseFiles("templates/show.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprint(w, err.Error())
	}

	db, err := sql.Open("sqlite3", "store.db")
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

	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM banks WHERE id = $1", bankId)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editBank(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	t, err := template.ParseFiles("templates/edit.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprint(w, err.Error())
	}

	db, err := sql.Open("sqlite3", "store.db")
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
	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec("UPDATE banks SET bank_name=$1, interest_rate=$2, maximum_loan=$3, minimum_down_payment=$4, loan_term=$5 WHERE id=$6",
		name, interestRate, maximumLoan, minimumDownPayment, loanTerm, id)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func calc(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/calc.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprint(w, err.Error())
	}

	db, err := sql.Open("sqlite3", "store.db")
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

func result(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/result.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprint(w, err.Error())
	}

	id := r.FormValue("banklist")
	initialLoanS := r.FormValue("initial_loan")
	downPaymentS := r.FormValue("down_payment")
	initialLoan, err := strconv.ParseUint(initialLoanS, 10, 64)
	if err != nil {
		panic(err)
	}
	downPayment, err := strconv.ParseUint(downPaymentS, 10, 64)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	res, err := db.Query(fmt.Sprintf("SELECT * FROM `banks` WHERE `id` = '%s'", id))
	if err != nil {
		panic(err)
	}
	bank := Bank{}
	for res.Next() {
		var b Bank
		err := res.Scan(&b.Id, &b.BankName, &b.InterestRate, &b.MaximumLoan, &b.MinimumDownPayment, &b.LoanTerm)
		if err != nil {
			panic(err)
		}
		bank = b
	}
	var result string
	if initialLoan > bank.MaximumLoan {
		result = fmt.Sprintf("you want to receive: $%d, but the bank can issue the maximum amount: $%d", initialLoan, bank.MaximumLoan)
	} else if downPayment < bank.MinimumDownPayment {
		result = fmt.Sprintf("you want to deposit the initial amount: $%d, but the bank needs more: $%d", downPayment, bank.MinimumDownPayment)
	} else {
		r := (float64(initialLoan-downPayment) * (bank.InterestRate / (12 * 100)) * math.Pow(1+bank.InterestRate/(12*100), float64(bank.LoanTerm))) / (math.Pow(1+bank.InterestRate/(12*100), float64(bank.LoanTerm)) - 1)
		result = fmt.Sprintf("your monthly payment is: $%.2f", r)
	}

	t.ExecuteTemplate(w, "result", result)
}
