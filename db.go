package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Name string `form:"name" binding:"required"`
	Pass string `form:"pass" binding:"required"`
}

func (u User) InsertDb(db *sql.DB) error {
	// Hash the password before storing it in the database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", u.Name, hashedPassword)
	return err
}

// Function to verify login credentials
func (u User) VerifyLogin(db *sql.DB) bool {
	// Retrieve the hashed password from the database
	var hashedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", u.Name).Scan(&hashedPassword)
	if err != nil {
		return false
	}

	// Compare the provided password with the hashed password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(u.Pass))
	return err == nil
}

// Function to query all users from the database
func QueryUsers(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT username FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "/tmp/db.sqlite3")
	if err != nil {
		return nil, err
	}

	// Create the users table if it doesn't exist
	_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS users (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					username TEXT NOT NULL UNIQUE,
					password TEXT NOT NULL
			)
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}
