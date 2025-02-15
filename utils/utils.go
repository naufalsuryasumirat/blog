package utils

import (
	"log"
    "os"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var StorageDir string
var StaticDir string

func chk(err error) {
    if err != nil {
        log.Panic(err)
    }
}

func init() {
	if err := godotenv.Load("local.env"); err != nil {
		log.Fatal(err)
	}

	connectDB()
	log.Println("[DB]: Connected to Database")

	StorageDir = os.Getenv("STORAGE_DIR")
    StaticDir = os.Getenv("STATIC_DIR")
}
