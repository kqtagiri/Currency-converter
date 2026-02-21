package main

import (
	app "Converter/App"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	app := app.App{}
	if err := app.FillAPP(); err != nil {
		panic(err)
	}

	app.Db.CreateTables()

	if err := app.Db.CheckRatesUpdate(); err != nil {
		panic(err)
	}

	router := mux.NewRouter()
	router.Path("/users").Methods("POST").HandlerFunc(app.CreateUser)
	router.Path("/users/").Methods("GET").HandlerFunc(app.GetUsers)
	router.Path("/users/{id}").Methods("GET").HandlerFunc(app.GetUserFullInfo)
	router.Path("/wallets").Methods("POST").HandlerFunc(app.CreateWallet)
	router.Path("/wallets/").Methods("GET").HandlerFunc(app.GetWallets)
	router.Path("/wallets/{id}").Methods("GET").HandlerFunc(app.GetWallet)
	router.Path("/exchange").Methods("POST").HandlerFunc(app.Exchange)
	router.Path("/transactions/").Methods("GET").HandlerFunc(app.GetTransactionsHistory)
	router.Path("/rates/").Methods("GET").Queries("FromCurr", "{FromCurr}", "ToCurr", "{ToCurr}").HandlerFunc(app.GetRate)
	router.Path("/rates/history/").Methods("GET").Queries("FromCurr", "{FromCurr}", "ToCurr", "{ToCurr}").HandlerFunc(app.GetRateHistory)

	http.ListenAndServe(":9111", router)

}
