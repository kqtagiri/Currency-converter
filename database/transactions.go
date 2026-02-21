package database

import (
	"errors"
	"time"
)

type Transaction struct {
	UserId     int     `json:"user_id"`
	FromCurr   string  `json:"from_currency"`
	ToCurr     string  `json:"to_currency"`
	FromAmount float64 `json:"from_amount"`
	ToAmount   float64 `json:"to_amount"`
	Rate       float64 `json:"rate"`
	CreatedAt  time.Time
}

type TransactionDTO struct {
	UserId     int     `json:"user_id"`
	FromCurr   string  `json:"from_currency"`
	ToCurr     string  `json:"to_currency"`
	FromAmount float64 `json:"from_amount"`
}

type TransactionModel struct {
	Id         int
	UserId     int
	FromCurr   string
	ToCurr     string
	FromAmount float64
	ToAmount   float64
	Rate       float64
	CreatedAt  time.Time
}

func NewTransaction(UserId int, FromAmount, ToAmount, Rate float64, FromCurr, ToCurr string) *Transaction {
	return &Transaction{
		UserId:     UserId,
		FromCurr:   FromCurr,
		ToCurr:     ToCurr,
		FromAmount: FromAmount,
		ToAmount:   ToAmount,
		Rate:       Rate,
		CreatedAt:  time.Now(),
	}
}

func NewTransactionModel(Id, UserId int, FromAmount, ToAmount, Rate float64, FromCurr, ToCurr string) *TransactionModel {
	return &TransactionModel{
		Id:         Id,
		UserId:     UserId,
		FromCurr:   FromCurr,
		ToCurr:     ToCurr,
		FromAmount: FromAmount,
		ToAmount:   ToAmount,
		Rate:       Rate,
		CreatedAt:  time.Now(),
	}
}

func ConvertFromTransactionModel(model TransactionModel) *Transaction {
	return &Transaction{
		UserId:     model.UserId,
		FromCurr:   model.FromCurr,
		ToCurr:     model.ToCurr,
		FromAmount: model.FromAmount,
		ToAmount:   model.ToAmount,
		Rate:       model.Rate,
		CreatedAt:  model.CreatedAt,
	}
}

func ConvertFromTransactionDTO(dto TransactionDTO) *Transaction {
	return &Transaction{
		UserId:     dto.UserId,
		FromCurr:   dto.FromCurr,
		ToCurr:     dto.ToCurr,
		FromAmount: dto.FromAmount,
		CreatedAt:  time.Now(),
	}
}

func ValidateTransaction(FromCurr, ToCurr string) error {
	if (FromCurr != "BYN" && FromCurr != "EUR" && FromCurr != "USD") ||
		(ToCurr != "BYN" && ToCurr != "EUR" && ToCurr != "USD") {
		return errors.New("No have this currency")
	}
	return nil
}

func (db Database) CreateTableTransactions() error {

	QueryRow := `CREATE TABLE IF NOT EXISTS transactions(
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL,
		from_currency VARCHAR(3) NOT NULL,
		to_currency VARCHAR(3) NOT NULL,
		from_amount DECIMAL(20, 2) NOT NULL,
		to_amount DECIMAL(20, 2) NOT NULL,
		rate DECIMAL(5,2) NOT NULL,
		created_at TIMESTAMP NOT NULL
	);`
	_, err := db.Conn.Exec(db.Ctx, QueryRow)
	return err

}

func (db Database) InsertTransaction(t Transaction) error {

	tx, err := db.Conn.Begin(db.Ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(db.Ctx)

	var balance, b float64
	QueryRow := `SELECT balance FROM wallets WHERE user_id = $1 AND currency = $2;`
	if err = tx.QueryRow(db.Ctx, QueryRow, t.UserId, t.FromCurr).Scan(&balance); err != nil {
		return err
	}

	if err = tx.QueryRow(db.Ctx, QueryRow, t.UserId, t.ToCurr).Scan(&b); err != nil {
		return err
	}

	if balance < t.FromAmount {
		return errors.New("Not enough balance to transaction")
	} else {
		QueryRow = `UPDATE wallets SET balance = balance-$1 WHERE user_id = $2 AND currency = $3;`
		if _, err := tx.Exec(db.Ctx, QueryRow, t.FromAmount, t.UserId, t.FromCurr); err != nil {
			return err
		}
		QueryRow = `UPDATE wallets SET balance = balance+$1 WHERE user_id = $2 AND currency = $3;`
		if _, err := tx.Exec(db.Ctx, QueryRow, t.ToAmount, t.UserId, t.ToCurr); err != nil {
			return err
		}
	}

	QueryRow = `INSERT INTO transactions (user_id, from_currency, to_currency, from_amount, to_amount, rate, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7);`
	if _, err := tx.Exec(db.Ctx, QueryRow, t.UserId, t.FromCurr, t.ToCurr, t.FromAmount, t.ToAmount, t.Rate, t.CreatedAt); err != nil {
		return err
	}

	tx.Commit(db.Ctx)
	return err

}

func (db Database) GetTransactionsHistory(offset int) ([]Transaction, error) {

	QueryRow := `SELECT * FROM transactions LIMIT 100 OFFSET 100*$1;`
	rows, err := db.Conn.Query(db.Ctx, QueryRow, offset)
	if err != nil {
		return nil, err
	}

	transactions := []Transaction{}
	model := TransactionModel{}
	count := 0
	for rows.Next() {

		count++
		if err := rows.Scan(&model.Id, &model.UserId, &model.FromCurr, &model.ToCurr, &model.FromAmount, &model.ToAmount, &model.Rate, &model.CreatedAt); err != nil {
			return nil, err
		}
		transaction := ConvertFromTransactionModel(model)
		transactions = append(transactions, *transaction)

	}
	if count == 0 {
		return nil, errors.New("No have transactions")
	}
	return transactions, nil

}
