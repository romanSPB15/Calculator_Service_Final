package application

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// Хранилище
type Storage struct {
	db *sql.DB // База данных
}

const (
	AppStoragePath  = "./data/data.db"     // Директория файла базы данных когда рабочая директория /cmd
	TestStoragePath = "../../data/data.db" // Директория файла базы данных когда рабочая директория /internal/application(для тестов)
)

// Открытие хранилища из файла базы данных
func OpenStorage(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	st := &Storage{db}
	return st, st.createTables()
}

// Создание таблиц, если они не существуют
func (st *Storage) createTables() error {
	const (
		usersTable = `
	CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		login TEXT,
		password TEXT
	);`

		expressionsTable = `
	CREATE TABLE IF NOT EXISTS expressions(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		data TEXT NOT NULL,
		user_id INTEGER NOT NULL,
		status TEXT NOT NULL,
		result FLOAT
	);`
	)

	if _, err := st.db.Exec(usersTable); err != nil {
		return err
	}

	if _, err := st.db.Exec(expressionsTable); err != nil {
		return err
	}

	return nil
}

// Очистить базу
func (st *Storage) Clear() error {
	var (
		q1 = `DELETE FROM users;`
		q2 = `DELETE FROM expressions;`
	)
	if _, err := st.db.Exec(q1); err != nil {
		return err
	}
	if _, err := st.db.Exec(q2); err != nil {
		return err
	}
	return nil
}

// Добавить пользователя
func (st *Storage) InsertUser(user *User) error {
	var q = `INSERT INTO users (id, login, password) values ($1, $2, $3)`
	_, err := st.db.Exec(q, user.ID, user.Login, user.Password)
	return err
}

// Добавить выражение
func (st *Storage) InsertExpression(exp *ExpressionWithID, forUser *User) error {
	var q = `INSERT INTO expressions (id, data, status, result, user_id) values ($1, $2, $3, $4, $5)`
	_, err := st.db.Exec(q, exp.ID, exp.Data, exp.Status, exp.Result, forUser.ID)
	return err
}

// Получить всех пользователей
func (st *Storage) SelectAllUsers() ([]*User, error) {
	var users []*User
	var q = `SELECT id, login, password FROM users`
	rows, err := st.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		u := &User{}
		err := rows.Scan(&u.ID, &u.Login, &u.Password)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

// Выражение с UserID
type ExpressionForUser struct {
	ExpressionWithID
	UserID uint32
}

// Получить все выражения для пользователя user
func (st *Storage) SelectExpressionsForUser(user *User) ([]ExpressionWithID, error) {
	var expressions []ExpressionWithID
	var q = `SELECT id, data, status, result, user_id FROM expressions`

	rows, err := st.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e := ExpressionForUser{}
		err := rows.Scan(&e.ID, &e.Data, &e.Status, &e.Result, &e.UserID)
		if err != nil {
			return nil, err
		}
		if e.UserID == user.ID {
			expressions = append(expressions, e.ExpressionWithID)
		}
	}

	return expressions, nil
}

// Получить все выражения в базе
func (st *Storage) SelectExpressions() ([]ExpressionForUser, error) {
	var expressions []ExpressionForUser
	var q = `SELECT id, data, status, result, user_id FROM expressions`

	rows, err := st.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e := ExpressionForUser{}
		err := rows.Scan(&e.ID, &e.Data, &e.Status, &e.Result, &e.UserID)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, e)
	}

	return expressions, nil
}

// Закрыть базу данных
func (s *Storage) Close() error {
	return s.db.Close()
}
