package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"

	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	renderer "github.com/naufalsuryasumirat/home-web-hobby/renderer"
	util "github.com/naufalsuryasumirat/home-web-hobby/utils"
)

var knownDocumentExtensions = map[string]bool{
	"md": true,
}
var knownImageExtensions = map[string]bool{
	"png":  true,
	"jpg":  true,
	"jpeg": true,
}

// TODO: frontend (vue.js / htmx)
// TODO: enable CORS, serve the images from the backend
// TODO: format image node correctly for HTML
// TODO: CRON updating pages (.md -> .html), query DB or generate view
// TODO: Edit markdowns, point to previous versions
// TODO: ping-pong, ding-dong, (bing-bong) verification (cookies?)

type FormMD struct {
	Document *multipart.FileHeader   `form:"document" binding:"required"`
	Images   []*multipart.FileHeader `form:"images" binding:"omitempty"`
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.Static("/images", string(os.Getenv("STORAGE_DIR")))

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.POST("/upload-markdown", func(c *gin.Context) {
		var form FormMD
		if err := c.ShouldBind(&form); err != nil {
			log.Println(err.Error())
			c.String(http.StatusBadRequest, "Request did not bind properly")
		}

		timestamp, timestampFormatted := util.GetTimestamp()

		document, _ := form.Document.Open()
		mdFile, _ := io.ReadAll(document)
		filenameFull := form.Document.Filename

		// sanitize/verify uploaded data (TODO: improve)
		filename, ext, found := strings.Cut(filenameFull, ".")
		if !found || !knownDocumentExtensions[ext] {
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

		hashOutput := util.HashUpload(timestampFormatted, filename)

		dirpath := filepath.Join(os.Getenv("STORAGE_DIR"), hashOutput)
		os.MkdirAll(dirpath, 0755)

		insertResult, err := util.GetDB().Exec(`
            INSERT INTO entries (dirpath, doc)
                VALUES (?, ?);
        `, hashOutput, timestamp)
		if err != nil {
			// TODO: handle (err? ok?)
			log.Println(err.Error())
		}

		lastEntryID, err := insertResult.LastInsertId()
		if err != nil {
			// TODO: handle (err? ok?)
			log.Println(err.Error())
		}
		var insertImagesQuery string = `
            INSERT INTO images (entry_id, fname) VALUES
        `
		values := []interface{}{}
		for _, image := range form.Images {
			imagename := util.SanitizeFilename(image.Filename)
			savepath := filepath.Join(dirpath, imagename)
			if err := c.SaveUploadedFile(image, savepath); err != nil {
				// TODO: handle err
				log.Println(err.Error())
			}
			insertImagesQuery += "(?, ?),"
			values = append(values, lastEntryID, imagename)
		}

		if len(form.Images) > 0 {
			statement, err := util.GetDB().Prepare(
				insertImagesQuery[0 : len(insertImagesQuery)-1],
			)
			if err != nil {
				// TODO: handle err
				log.Println(err.Error())
			}

			if _, err = statement.Exec(values...); err != nil {
				// TODO: handle err
				log.Println(err.Error())
			}
		}

		c.SaveUploadedFile(form.Document, filepath.Join(dirpath, filenameFull))
		generatedHTML := renderer.MdToHTML(
			mdFile, &hashOutput, util.SanitizeFilename)
		os.WriteFile(
			filepath.Join(dirpath, filename+".html"),
			generatedHTML,
			0755,
		)

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

func main() {
	if err := godotenv.Load("local.env"); err != nil {
		log.Fatal(err)
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal(err)
	}

	r := setupRouter()
	r.Run(fmt.Sprintf(":%d", port))
}
