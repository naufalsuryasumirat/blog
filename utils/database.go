package utils

import (
	"fmt"
	"log"
	"os"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func GetDB() *sql.DB {
	return db
}

func connectDB() {
	testDB, err := sql.Open("sqlite3", string(os.Getenv("DB_PATH")))
	if err != nil {
		log.Panic(err.Error())
	}
	db = testDB
	migrateDB()
}

func migrateDB() {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS entries (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            dirpath VARCHAR(64) NOT NULL,
            doc DATETIME NOT NULL);
        CREATE TABLE IF NOT EXISTS images (
            entry_id INTEGER NOT NULL,
            fname VARCHAR(64) NOT NULL,
            PRIMARY KEY (entry_id, fname),
            FOREIGN KEY (entry_id)
                REFERENCES entries (id));
    `)
	if err != nil {
		log.Println(err.Error())
	}
}

func init() {
	connectDB()
	fmt.Println("[DB]: Connected to Database")
}
