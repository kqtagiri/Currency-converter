package app

import (
	"Converter/database"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const PageDefault = 0

func (app App) CreateUser(w http.ResponseWriter, r *http.Request) {

	var user database.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	if err := database.ValidateUser(user.Name, user.Email); err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	data, err := json.MarshalIndent(user, "", "\t")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	if err := app.Db.InsertUser(user); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(201)
	w.Write(data)

}

func (app App) GetUsers(w http.ResponseWriter, r *http.Request) {

	page := PageDefault
	var err error

	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}
	}

	users, err := app.Db.GetUsers(page)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}

	data, err := json.MarshalIndent(users, "", "\t")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write(data)

}

func (app App) GetUserFullInfo(w http.ResponseWriter, r *http.Request) {

	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	user, err := app.Db.GetUserFullInfo(id)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}

	data, err := json.MarshalIndent(user, "", "\t")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write(data)

}

func (app App) CreateWallet(w http.ResponseWriter, r *http.Request) {

	var wallet database.Wallet
	if err := json.NewDecoder(r.Body).Decode(&wallet); err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	if err := database.ValidateWallet(wallet.Currency, wallet.Balance); err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	if err := app.Db.InsertWallet(wallet); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	data, err := json.MarshalIndent(wallet, "", "\t")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(201)
	w.Write(data)

}

func (app App) GetWallets(w http.ResponseWriter, r *http.Request) {

	page := PageDefault
	var err error

	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}
	}

	wallets, err := app.Db.GetWallets(page)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	data, err := json.MarshalIndent(wallets, "", "\t")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write(data)

}

func (app App) GetWallet(w http.ResponseWriter, r *http.Request) {

	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	wallet, err := app.Db.GetWallet(id)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	data, err := json.MarshalIndent(wallet, "", "\t")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write(data)

}

func (app App) Exchange(w http.ResponseWriter, r *http.Request) {

	var transDTO database.TransactionDTO
	if err := json.NewDecoder(r.Body).Decode(&transDTO); err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	if err := database.ValidateTransaction(transDTO.FromCurr, transDTO.ToCurr); err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	transaction := *database.ConvertFromTransactionDTO(transDTO)
	rate, err := app.Db.GetRate(transaction.FromCurr, transaction.ToCurr)
	if err != nil {
		w.WriteHeader(501)
		w.Write([]byte(err.Error()))
		return
	}
	transaction.Rate = rate.Rate
	transaction.ToAmount = transaction.FromAmount * transaction.Rate

	if err := app.Db.InsertTransaction(transaction); err != nil {
		w.WriteHeader(502)
		w.Write([]byte(err.Error()))
		return
	}

	data, err := json.MarshalIndent(transaction, "", "\t")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(201)
	w.Write(data)

}

func (app App) GetTransactionsHistory(w http.ResponseWriter, r *http.Request) {

	page := PageDefault
	var err error

	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}
	}

	transactions, err := app.Db.GetTransactionsHistory(page)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	data, err := json.MarshalIndent(transactions, "", "\t")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write(data)

}

func (app App) GetRate(w http.ResponseWriter, r *http.Request) {

	FromCurr := mux.Vars(r)["FromCurr"]
	ToCurr := mux.Vars(r)["ToCurr"]

	rate, err := app.Db.GetRate(FromCurr, ToCurr)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}

	data, err := json.MarshalIndent(rate, "", "\t")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write(data)

}

func (app App) GetRateHistory(w http.ResponseWriter, r *http.Request) {

	page := PageDefault
	var err error
	FromCurr := mux.Vars(r)["FromCurr"]
	ToCurr := mux.Vars(r)["ToCurr"]

	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}
	}

	rates, err := app.Db.GetRateHistory(FromCurr, ToCurr, page)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}

	data, err := json.MarshalIndent(rates, "", "\t")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write(data)

}
