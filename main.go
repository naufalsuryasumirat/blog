package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"crypto/sha1"
	"database/sql"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/joho/godotenv"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var hasher = sha1.New()
var knownDocumentExtensions = map[string]bool{
    "md": true,
}
var knownImageExtensions = map[string]bool{
    "png": true,
    "jpg": true,
    "jpeg": true,
}

// TODO: enable CORS, serve the images from the backend
// TODO: format image node correctly for HTML
// TODO: CRON updating pages (.md -> .html), query DB or generate view
// TODO: ping-pong, ding-dong, (bing-bong) verification (cookies?)
// TODO: frontend (vue.js)
// TODO: research about custom renderer
// TODO: edit markdowns and preserve previous versions

func connectDB() {
    testDB, err := sql.Open("sqlite3", "./db/local.db") // TODO: env
    if (err != nil) { // TODO: handle errors
        println(err.Error())
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
        println(err.Error())
    }
}

type FormMD struct {
    Document *multipart.FileHeader `form:"file" binding:"required"`
    Images []*multipart.FileHeader `form:"images" binding:"omitempty"`
}

func hashUpload(datetime string, filename string) string {
    hasher.Write([]byte(datetime + "_" + filename))
    output := hasher.Sum(nil)
    hasher.Reset()
    
    return fmt.Sprintf("%x", output)
}

func getTimestamp()(time.Time, string) {
    const format string = "2006_01_02-15_04"
    timestamp := time.Now().UTC()
    timestampFormatted := timestamp.Format(format)
    
    return timestamp, timestampFormatted
}

func sanitizeFilename(filename string) string {
    replacer := strings.NewReplacer( // TODO: singleton (?)
        " ", "_",   ",", "",
        "/", "-",   "\\", "",
        "\"", "",   "'", "",
        "*", "",    "?", "",   
        "<", "",    ">", "",
        "(", "",    ")", "",
        "[", "",    "]", "",
        "{", "",    "}", "",
        ":", "",    ";", "",
        "|", "",    "~", "",
        "%", "",
    )
    println(filename, filename)
    return replacer.Replace(strings.ToLower(filename))
}

func renderImage(
    w io.Writer,
    node *ast.Image,
    entering bool,
    dirpath *string,
    fn func(string) string,
) {
    if entering {
        w.Write(
            []byte(fmt.Sprintf(
                "<test> title: %s, dirpath: %s",
                fn(string(node.Destination)),
                *dirpath,
            )),
        )
    } else {
        w.Write([]byte("</test>"))
    }
}

func makeRenderHook(dirpath *string, fn func(string) string) html.RenderNodeFunc {
    return func(
        w io.Writer,
        node ast.Node,
        entering bool,
    ) (ast.WalkStatus, bool) {
        if image, ok := node.(*ast.Image); ok {
            renderImage(w, image, entering, dirpath, fn)
            return ast.GoToNext, true
        }
        
        return ast.GoToNext, false
    }
}

func mdToHTML(md []byte, dirpath *string, fn func(string) string) []byte {
    // create markdown parser with extensions
    extensions := parser.CommonExtensions |
        parser.AutoHeadingIDs |
        parser.NoEmptyLineBeforeBlock
    p := parser.NewWithExtensions(extensions)
    doc := p.Parse(md)

    // create HTML renderer with extensions
    htmlFlags := html.CommonFlags | html.HrefTargetBlank
    opts := html.RendererOptions{
        Flags: htmlFlags,
        RenderNodeHook: makeRenderHook(dirpath, fn),
    }
    renderer := html.NewRenderer(opts)
    // TODO: singleton renderer (?)

    return markdown.Render(doc, renderer)
}

func setupRouter() *gin.Engine {
    connectDB()
    
	r := gin.Default()
    // TODO: better to use binding header (?), documentation!
    r.Static("/images", os.Getenv("STORAGE_DIR"))

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

    r.POST("/upload-markdown", func(c *gin.Context) {
        var form FormMD
        c.ShouldBind(&form) // TODO: handle errors
        
        document, _ := form.Document.Open()
        mdFile, _ := io.ReadAll(document)
        
        timestamp, timestampFormatted := getTimestamp()
        filenameFull := form.Document.Filename
        
        filename, ext, found := strings.Cut(filenameFull, ".")
        if !found || !knownDocumentExtensions[ext]  {
            c.String(http.StatusBadRequest, "Unknown document extension")
            return
        }
        
        for _, image := range form.Images {
            _, ext, found := strings.Cut(image.Filename, ".")
            if !found || !knownImageExtensions[ext] {
                c.String(http.StatusBadRequest, "Unknown image extension")
                return
            }
        }

        hashOutput := hashUpload(timestampFormatted, filename)

        dirpath := filepath.Join(os.Getenv("STORAGE_DIR"), hashOutput)
        os.MkdirAll(dirpath, 0755)
        
        insertResult, err := db.Exec(`
            INSERT INTO entries (dirpath, doc)
                VALUES (?, ?);
        `, hashOutput, timestamp)
        if err != nil { // TODO: handle (err? ok?)
            println(err.Error())
        }
        
        lastEntryID, err := insertResult.LastInsertId()
        if err != nil { // TODO: handle (err? ok?)
            println(err.Error())
        }
        var insertImagesQuery string = `
            INSERT INTO images (entry_id, fname) VALUES
        `
        values := []interface{}{}
        for _, image := range form.Images {
            imagename := sanitizeFilename(image.Filename)
            savepath := filepath.Join(dirpath, imagename)
            err := c.SaveUploadedFile(image, savepath)
            if err != nil {
                println(err.Error())
            }
            insertImagesQuery += "(?, ?),"
            values = append(values, lastEntryID, imagename)
        }

        if len(form.Images) > 0 {
            statement, err := db.Prepare(
                insertImagesQuery[0:len(insertImagesQuery) - 1],
            )
            if err != nil { // TODO: handle err
                println(err.Error())
            }
            statement.Exec(values...)
        }

        c.SaveUploadedFile(form.Document, filepath.Join(dirpath, filenameFull))
        generatedHTML := mdToHTML(mdFile, &hashOutput, sanitizeFilename)
        os.WriteFile(
            filepath.Join(dirpath, filename + ".html"),
            generatedHTML,
            0755,
        )

        // TODO: hash-output + image_name -> HTML
        c.Data(http.StatusOK, "text/html; charset=utf-8", generatedHTML)
    }) 

	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

    authorized.POST("add", func(c *gin.Context) {
        // user := c.MustGet(gin.AuthUserKey).(string)
        // var json struct {
        //     Value string `json:"value" binding:"required"`
        // }
    })

	return r
}

func loadPort() int { // TODO: expand if needed
    err := godotenv.Load("local.env")
    if err != nil {
        log.Fatal(err)
    }

    port := os.Getenv("PORT")
    portNum, err := strconv.Atoi(port)
    if err != nil {
        log.Fatal(err)
    }
    
    return portNum 
}

func main() { // TODO: gin.SetMode here
	r := setupRouter()
    r.Run(fmt.Sprintf(":%d", loadPort()))
}
