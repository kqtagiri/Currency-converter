package database

import "errors"

type Wallet struct {
	UserId   int     `json:"user_id"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}

type WalletModel struct {
	Id       int
	UserId   int
	Currency string
	Balance  float64
}

func NewWallet(user_id int, balance float64, currency string) *Wallet {
	return &Wallet{
		UserId:   user_id,
		Currency: currency,
		Balance:  balance,
	}
}

func NewWalletModel(id, user_id int, balance float64, currency string) *WalletModel {
	return &WalletModel{
		Id:       id,
		UserId:   user_id,
		Currency: currency,
		Balance:  balance,
	}
}

func ConvertFromWalletModel(model WalletModel) *Wallet {
	return &Wallet{
		UserId:   model.UserId,
		Currency: model.Currency,
		Balance:  model.Balance,
	}
}

func ValidateWallet(curr string, balance float64) error {
	if curr != "EUR" && curr != "USD" && curr != "BYN" {
		return errors.New("No have this currency")
	}
	if balance < 0 {
		return errors.New("Balance can`t be < 0")
	}
	return nil
}

func (db Database) CreateTableWallets() error {

	QueryRow := `CREATE TABLE IF NOT EXISTS wallets(
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL,
		currency VARCHAR(3) NOT NULL,
		balance DECIMAL(20, 2) NOT NULL
	);`
	_, err := db.Conn.Exec(db.Ctx, QueryRow)
	return err

}

func (db Database) InsertWallet(w Wallet) error {

	QueryRow := `SELECT EXISTS(SELECT * FROM users WHERE id = $1);`
	var exists bool
	if err := db.Conn.QueryRow(db.Ctx, QueryRow, w.UserId).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return errors.New("User not found")
	}

	QueryRow = `SELECT EXISTS(SELECT FROM wallets WHERE user_id = $1 AND currency = $2);`
	if err := db.Conn.QueryRow(db.Ctx, QueryRow, w.UserId, w.Currency).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return errors.New("Wallet already exist")
	}

	QueryRow = `INSERT INTO wallets (user_id, currency, balance) VALUES ($1,$2,$3);`
	_, err := db.Conn.Exec(db.Ctx, QueryRow, w.UserId, w.Currency, w.Balance)
	return err

}

func (db Database) GetWallets(offset int) ([]Wallet, error) {

	QueryRow := `SELECT * FROM wallets LIMIT 100 OFFSET 100*$1;`
	rows, err := db.Conn.Query(db.Ctx, QueryRow, offset)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, errors.New("No have wallets")
	}

	wallets := []Wallet{}
	model := WalletModel{}
	for rows.Next() {

		if err := rows.Scan(&model.Id, &model.UserId, &model.Currency, &model.Balance); err != nil {
			return nil, err
		}
		wallet := ConvertFromWalletModel(model)
		wallets = append(wallets, *wallet)

	}
	return wallets, nil

}

func (db Database) GetWallet(id int) (*Wallet, error) {

	QueryRow := `SELECT * FROM wallets WHERE id = $1;`
	row := db.Conn.QueryRow(db.Ctx, QueryRow, id)
	model := WalletModel{}
	if err := row.Scan(&model.Id, &model.UserId, &model.Currency, &model.Balance); err != nil {
		return nil, err
	}
	wallet := ConvertFromWalletModel(model)
	return wallet, nil

}
