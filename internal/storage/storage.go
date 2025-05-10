package storage

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/romanSPB15/Calculator_Service_Final/pckg/types"
)

// Хранилище
type Storage struct {
	db *sql.DB // База данных
	mx *sync.Mutex
}

// Открытие хранилища из файла базы данных
func Open(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	st := &Storage{db, &sync.Mutex{}}
	return st, st.createTables()
}

// Создание таблиц, если они не существуют
func (st *Storage) createTables() error {
	st.mx.Lock()
	defer st.mx.Unlock()
	const (
		usersTable = `
	CREATE TABLE IF NOT EXISTS users(
		id TEXT PRIMARY KEY, 
		login TEXT,
		password TEXT
	);`

		expressionsTable = `
	CREATE TABLE IF NOT EXISTS expressions(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		data TEXT NOT NULL,
		status TEXT NOT NULL,
		result FLOAT,
		user_id TEXT NOT NULL
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
	st.mx.Lock()
	defer st.mx.Unlock()
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
func (st *Storage) InsertUser(user *types.User) error {
	st.mx.Lock()
	defer st.mx.Unlock()
	var q = `INSERT INTO users (id, login, password) values ($1, $2, $3)`
	_, err := st.db.Exec(q, user.ID, user.Login, user.Password)
	return err
}

// Добавить выражение
func (st *Storage) InsertExpression(exp *types.ExpressionWithID, forUser *types.User) error {
	st.mx.Lock()
	defer st.mx.Unlock()
	var q = `INSERT INTO expressions (id, data, status, result, user_id) values ($1, $2, $3, $4, $5)`
	_, err := st.db.Exec(q, exp.ID, exp.Data, exp.Status, exp.Result, forUser.ID)
	return err
}

// Получить всех пользователей
func (st *Storage) SelectAllUsers() ([]*types.User, error) {
	st.mx.Lock()
	defer st.mx.Unlock()
	var users []*types.User
	var q = `SELECT id, login, password FROM users`
	rows, err := st.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		u := &types.User{}
		err := rows.Scan(&u.ID, &u.Login, &u.Password)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

// Получить все выражения для пользователя user
func (st *Storage) SelectExpressionsForUser(user *types.User) ([]*types.ExpressionWithID, error) {
	st.mx.Lock()
	defer st.mx.Unlock()
	var expressions []*types.ExpressionWithID
	var q = `SELECT id, data, status, result, user_id FROM expressions`

	rows, err := st.db.Query(q)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		e := &types.ExpressionForUser{}
		err := rows.Scan(&e.ID, &e.Data, &e.Status, &e.Result, &e.UserID)
		if err != nil {
			return nil, err
		}
		if e.UserID == user.ID {
			expressions = append(expressions, &e.ExpressionWithID)
		}
	}

	return expressions, rows.Close()
}

// Получить все выражения в базе
func (st *Storage) SelectExpressions() ([]*types.ExpressionForUser, error) {
	st.mx.Lock()
	defer st.mx.Unlock()
	var expressions []*types.ExpressionForUser
	var q = `SELECT id, data, status, result, user_id FROM expressions`

	rows, err := st.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e := &types.ExpressionForUser{}
		err := rows.Scan(&e.ID, &e.Data, &e.Status, &e.Result, &e.UserID)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, e)
	}

	return expressions, nil
}

// Закрыть базу данных
func (st *Storage) Close() error {
	st.mx.Lock()
	err := st.db.Close()
	st.mx.Unlock()
	return err
}
