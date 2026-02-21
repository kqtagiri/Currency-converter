package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Rates struct {
	FromCurr  string    `json:"from_currency"`
	ToCurr    string    `json:"to_currency"`
	Rate      float64   `json:"rate"`
	CreatedAt time.Time `json:"created_at"`
}

type RateDTO struct {
	FromCurr string  `json:"Cur_Abbreviation"`
	Rate     float64 `json:"Cur_OfficialRate"`
	Date     string  `json:"Date"`
}

type RateModel struct {
	Id        int
	FromCurr  string
	ToCurr    string
	Rate      float64
	CreatedAt time.Time
}

func NewRate(FromCurr, ToCurr string, Rate float64, Date time.Time) *Rates {
	return &Rates{
		FromCurr:  FromCurr,
		ToCurr:    ToCurr,
		Rate:      Rate,
		CreatedAt: Date,
	}
}

func NewRateModel(FromCurr, ToCurr string, Id int, Rate float64, Date time.Time) *RateModel {
	return &RateModel{
		Id:        Id,
		FromCurr:  FromCurr,
		ToCurr:    ToCurr,
		Rate:      Rate,
		CreatedAt: Date,
	}
}

func ConvertFromRateModel(model RateModel) *Rates {
	return &Rates{
		FromCurr:  model.FromCurr,
		ToCurr:    model.ToCurr,
		Rate:      model.Rate,
		CreatedAt: model.CreatedAt,
	}
}

func ValidateRate(FromCurr, ToCurr string) error {
	if (FromCurr != "BYN" && FromCurr != "EUR" && FromCurr != "USD") ||
		(ToCurr != "BYN" && ToCurr != "EUR" && ToCurr != "USD") {
		return errors.New("No have this currency")
	}
	return nil
}

func (db Database) CreateTableRates() error {

	QueryRow := `CREATE TABLE IF NOT EXISTS rates(
		id SERIAL PRIMARY KEY,
		from_currency VARCHAR(3) NOT NULL,
		to_currency VARCHAR(3) NOT NULL,
		rate DECIMAL(5,2) NOT NULL,
		created_at TIMESTAMP NOT NULL
	);`
	_, err := db.Conn.Exec(db.Ctx, QueryRow)
	return err

}

func (db Database) CheckRatesUpdate() error {

	resp, err := http.Get("https://api.nbrb.by/exrates/rates?periodicity=0")
	if err != nil {
		return err
	}
	var rateEUR, rateUSD float64
	defer resp.Body.Close()

	rates := []RateDTO{}
	if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
		return err
	}
	rateDate, err := time.Parse("2006-01-02T15:04:05", rates[0].Date)
	if err != nil {
		return err
	}

	for _, rate := range rates {

		if err := ValidateRate(rate.FromCurr, "BYN"); err != nil {
			continue
		}
		if rate.FromCurr == "USD" {
			rateUSD = rate.Rate
		} else {
			rateEUR = rate.Rate
		}

		QueryRow := `SELECT created_at FROM rates WHERE from_currency = $1 AND to_currency = $2 ORDER BY created_at DESC;`
		var date time.Time
		if err := db.Conn.QueryRow(db.Ctx, QueryRow, rate.FromCurr, "BYN").Scan(&date); err != nil || date != rateDate {
			fmt.Println(date, '\n', rateDate)
			QueryRow = `INSERT INTO rates (from_currency, to_currency, rate, created_at) VALUES ($1,$2,$3,$4);`
			if _, err := db.Conn.Exec(db.Ctx, QueryRow, rate.FromCurr, "BYN", rate.Rate, rate.Date); err != nil {
				return err
			}
		}

		QueryRow = `SELECT created_at FROM rates WHERE from_currency = $1 AND to_currency = $2 ORDER BY created_at DESC;`
		if err := db.Conn.QueryRow(db.Ctx, QueryRow, "BYN", rate.FromCurr).Scan(&date); err != nil || date != rateDate {
			QueryRow = `INSERT INTO rates (from_currency, to_currency, rate, created_at) VALUES ($1,$2,$3,$4);`
			if _, err := db.Conn.Exec(db.Ctx, QueryRow, "BYN", rate.FromCurr, 1.0/rate.Rate, rate.Date); err != nil {
				return err
			}
		}

	}

	var date time.Time
	QueryRow := `SELECT created_at FROM rates WHERE from_currency = $1 AND to_currency = $2 ORDER BY created_at DESC;`
	if err := db.Conn.QueryRow(db.Ctx, QueryRow, "EUR", "USD").Scan(&date); err != nil || date != rateDate {
		QueryRow = `INSERT INTO rates (from_currency, to_currency, rate, created_at) VALUES ($1,$2,$3,$4);`
		if _, err := db.Conn.Exec(db.Ctx, QueryRow, "EUR", "USD", rateEUR/rateUSD, rates[0].Date); err != nil {
			return err
		}
	}

	QueryRow = `SELECT created_at FROM rates WHERE from_currency = $1 AND to_currency = $2 ORDER BY created_at DESC;`
	if err := db.Conn.QueryRow(db.Ctx, QueryRow, "USD", "EUR").Scan(&date); err != nil || date != rateDate {
		QueryRow = `INSERT INTO rates (from_currency, to_currency, rate, created_at) VALUES ($1,$2,$3,$4);`
		if _, err := db.Conn.Exec(db.Ctx, QueryRow, "USD", "EUR", rateUSD/rateEUR, rates[0].Date); err != nil {
			return err
		}
	}

	return nil

}

func (db Database) GetRate(FromCurr, ToCurr string) (*Rates, error) {

	db.CheckRatesUpdate()

	QueryRow := `SELECT * FROM rates WHERE from_currency = $1 AND to_currency = $2 ORDER BY created_at DESC LIMIT 1;`
	row := db.Conn.QueryRow(db.Ctx, QueryRow, FromCurr, ToCurr)
	model := RateModel{}
	if err := row.Scan(&model.Id, &model.FromCurr, &model.ToCurr, &model.Rate, &model.CreatedAt); err != nil {
		return nil, err
	}
	rate := ConvertFromRateModel(model)
	return rate, nil

}

func (db Database) GetRateHistory(FromCurr, ToCurr string, offset int) ([]Rates, error) {

	db.CheckRatesUpdate()

	QueryRow := `SELECT * FROM rates WHERE from_currency = $1 AND to_currency = $2 ORDER BY created_at DESC LIMIT 100 OFFSET 100*$3;`
	rows, err := db.Conn.Query(db.Ctx, QueryRow, FromCurr, ToCurr, offset)
	if err != nil {
		return nil, err
	}

	rates := []Rates{}
	model := RateModel{}
	count := 0
	for rows.Next() {

		count++
		if err := rows.Scan(&model.Id, &model.FromCurr, &model.ToCurr, &model.Rate, &model.CreatedAt); err != nil {
			return nil, err
		}
		rate := ConvertFromRateModel(model)
		rates = append(rates, *rate)

	}
	if count == 0 {
		return nil, errors.New("No have rate history")
	}
	return rates, nil

}
