package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"

	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/joho/godotenv"

	// "github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

var db = make(map[string]string)

// TODO: how to parse image and show it (?)
// TODO: CRON job that updates the static page every day
// TODO: POST a .md file, and immediately store in database, but don't update yet
// TODO: make it possible to reference other posts
// TODO: make the first code recursive (references itself)

// TODO: use PING-PONG,DING-DONG for verification (http cookies?)

type FormMD struct {
    File *multipart.FileHeader `form:"file" binding:"required"`
}

func mdToHTML(md []byte) []byte {
    // create markdown parser with extensions
    extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
    p := parser.NewWithExtensions(extensions)
    doc := p.Parse(md)

    // create HTML renderer with extensions
    htmlFlags := html.CommonFlags | html.HrefTargetBlank
    opts := html.RendererOptions{Flags: htmlFlags}
    renderer := html.NewRenderer(opts)

    return markdown.Render(doc, renderer)
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

    r.POST("/upload-markdown", func(c *gin.Context) {
        // c.String(http.StatusOK, "adding markdown")
        // var upFile *multipart.FileHeader
        var form FormMD
        c.ShouldBind(&form)

        file, _ := form.File.Open()
        markdown, _ := io.ReadAll(file)
        // TODO: custom renderer, save to database, query/create new view everyday
        // TODO: handle images
        // TODO: *everyday* being the key part
        c.Data(http.StatusOK, "text/html; charset=utf-8", mdToHTML(markdown))

        // var md []byte
        // c.ShouldBind(md)
        // c.HTML(http.StatusOK, "text/html", mdToHTML(md))
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
