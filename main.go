package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"

	a "github.com/naufalsuryasumirat/blog/auth"
	h "github.com/naufalsuryasumirat/blog/handlers"
	m "github.com/naufalsuryasumirat/blog/middleware"
	u "github.com/naufalsuryasumirat/blog/utils"
)

func route() *gin.Engine {
	r := gin.Default()
    r.Use(m.CORSMiddleware())

	r.Static("/audios", u.StorageDir)
	r.Static("/images", u.StorageDir)
	r.Static("/static", u.StaticDir) 
    r.StaticFile("/favicon.ico", filepath.Join(u.StaticDir, "favicon.ico"))

    // setting middlewares
    r.Use(
        m.TextHTMLMiddleware(),
        m.CSPMiddleware(),
    )

    r.NoRoute(h.NotFound)
    r.GET("/", func(c *gin.Context) {
        c.Redirect(http.StatusFound, "/tech")
    })
	r.GET("/tech", h.Home())
	r.GET("/ent", h.Entertainment())
    r.GET("/doc/:art", h.Document)
	r.GET("/about", func(c *gin.Context) {
        c.Redirect(
            http.StatusFound,
            fmt.Sprintf("/doc/%s", os.Getenv("ABOUT_HASH")),
        )
    })
	r.GET("/ping", h.Ping)
	r.GET("/bing", h.Pong)

	authorized := r.Group("/", a.CookieAuth)
	{
        authorized.GET("/add", h.GetUploadMarkdown)
        authorized.POST("/add", h.PostUploadMarkdown)
        authorized.GET("/edit/:art", h.GetEditArticle)
        authorized.PATCH("/edit/:art", h.PatchArticle)
	}

	return r
}

// schedule every job that needs scheduling
func schedule() gocron.Scheduler {
	s, err := gocron.NewScheduler()
	chk(err)

	j, err := s.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(1, 0, 0))),
		gocron.NewTask(a.GenRandomSha),
	)
	chk(err)
	log.Printf("JobGen[ID]: %s\n", j.ID().String())

	return s
}

func chk(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal(err)
	}

	s := schedule()
	s.Start()
	defer s.Shutdown()

	r := route()
	r.Run(fmt.Sprintf(":%d", port))
}
