package database

import (
	"errors"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserModel struct {
	Id    int
	Name  string
	Email string
}

type UserWithWallets struct {
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}

func NewUser(name, email string) *User {
	return &User{
		Name:  name,
		Email: email,
	}
}

func NewUserModel(id int, name, email string) *UserModel {
	return &UserModel{
		Id:    id,
		Name:  name,
		Email: email,
	}
}

func ConvertFromUserModel(model UserModel) *User {
	return &User{
		Name:  model.Name,
		Email: model.Email,
	}
}

func ValidateUser(name, email string) error {
	if name == "" {
		return errors.New("Name wasn`t be unfilled")
	}
	if email == "" {
		return errors.New("Email wasn`t be unfilled")
	}
	return nil
}

func (db Database) CreateTableUsers() error {

	QueryRow := `CREATE TABLE IF NOT EXISTS users(
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) NOT NULL,
		UNIQUE(email)
	);`
	_, err := db.Conn.Exec(db.Ctx, QueryRow)
	return err

}

func (db Database) InsertUser(user User) error {

	QueryRow := `INSERT INTO users (name, email) VALUES ($1, $2);`
	_, err := db.Conn.Exec(db.Ctx, QueryRow, user.Name, user.Email)
	return err

}

func (db Database) GetUsers(offset int) ([]User, error) {

	users := []User{}
	model := UserModel{}
	QueryRow := `SELECT * FROM users LIMIT 100 OFFSET 100*$1;`
	rows, err := db.Conn.Query(db.Ctx, QueryRow, offset)
	if err != nil {
		return nil, err
	}

	count := 0
	for rows.Next() {

		count++
		if err := rows.Scan(&model.Id, &model.Name, &model.Email); err != nil {
			return nil, err
		}
		user := ConvertFromUserModel(model)
		users = append(users, *user)

	}
	if count == 0 {
		return nil, errors.New("No have users")
	}
	return users, nil

}

func (db Database) GetUserFullInfo(id int) ([]UserWithWallets, error) {

	QueryRow := `SELECT u.name, u.email, w.currency, w.balance FROM users u JOIN wallets w ON u.id = w.user_id WHERE u.id = $1;`
	rows, err := db.Conn.Query(db.Ctx, QueryRow, id)
	if err != nil {
		return nil, err
	}

	users := []UserWithWallets{}
	count := 0
	for rows.Next() {

		count++
		user := UserWithWallets{}
		if err := rows.Scan(&user.Name, &user.Email, &user.Currency, &user.Balance); err != nil {
			return nil, err
		}
		users = append(users, user)

	}
	if count == 0 {
		return nil, errors.New("No have user with this name or user no have wallets")
	}
	return users, nil

}
