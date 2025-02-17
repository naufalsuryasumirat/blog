package utils

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

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
            doc DATETIME NOT NULL
        );
        CREATE TABLE IF NOT EXISTS images (
            entry_id INTEGER NOT NULL,
            fname VARCHAR(64) NOT NULL,
            PRIMARY KEY (entry_id, fname),
            FOREIGN KEY (entry_id)
                REFERENCES entries (id)
        );
        CREATE TABLE IF NOT EXISTS articles (
            entry_id INTEGER PRIMARY KEY,
            title VARCHAR(128) NOT NULL,
            blurb VARCHAR(256) NOT NULL,
            category VARCHAR(64) DEFAULT 'tech' NOT NULL,
            hidden INTEGER NOT NULL CHECK (hidden IN (0, 1)),
            FOREIGN KEY (entry_id)
                REFERENCES entries(id)
        );
    `)
	if err != nil {
		log.Println(err.Error())
	}
}

type Entry struct {
	Dirpath string    `db:"dirpath"`
	Doc     time.Time `db:"doc"`
}

type Article struct {
    Title   string    `db:"title"`
    Blurb   string    `db:"blurb"`
	Dirpath string    `db:"dirpath"`
	Doc     time.Time `db:"doc"`
    Image   string    `db:"fname"`
}

// returns entries (latest first), found boolean
func GetEntries(path string) ([]Entry, bool) {
	var entries []Entry
	row, err := db.Query(
		"SELECT dirpath, doc FROM entries WHERE dirpath = ? ORDER BY doc DESC;",
		path,
	)
	if err != nil && err != sql.ErrNoRows {
		return entries, false
	}

	defer row.Close()

	for row.Next() {
		var e Entry
		row.Scan(&e.Dirpath, &e.Doc)
		entries = append(entries, e)
	}

	return entries, true
}

// returns entry, found boolean
func GetLatestEntry(path string) (Entry, bool) {
	entries, found := GetEntries(path)
	if !found || len(entries) == 0 {
		return Entry{}, false
	}

	return entries[0], true
}

func GetArticlesList(ctg string, cursor int) []Article {
    const rowLimit = 5
    row, err := db.Query(
        `SELECT t2.dirpath, t1.title, t1.blurb, t2.doc, t3.fname
            FROM articles t1
                LEFT OUTER JOIN entries t2 ON (t1.entry_id=t2.id)
                LEFT OUTER JOIN images t3 ON (t1.entry_id=t3.entry_id)
            WHERE t1.category=?
                AND t1.hidden=FALSE
                GROUP BY t1.entry_id
                ORDER BY t2.doc DESC
            LIMIT ?
            OFFSET ?;`,
        ctg,
        rowLimit,
        cursor * rowLimit,
    )
    if err != sql.ErrNoRows {
        chk(err)
    }

    var arts []Article
    for row.Next() {
        var art Article
        row.Scan(&art.Dirpath, &art.Title, &art.Blurb, &art.Doc, &art.Image)

        if len(art.Image) > 0 {
            // FIXME: check if thumbnail exists first
            art.Image = fmt.Sprintf("images/%s/%s.thumbnail", art.Dirpath, art.Image)
        }

        art.Doc = art.Doc.In(time.Now().Location())

        arts = append(arts, art)
    }

    return arts
}

// TODO: implement, get all articles revision and page them
func GetArticles(path string) ([]string, bool) {
	var res []string
	return res, false
}

