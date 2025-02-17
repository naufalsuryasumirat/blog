package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	a "github.com/naufalsuryasumirat/blog/auth"
	t "github.com/naufalsuryasumirat/blog/templates"
	u "github.com/naufalsuryasumirat/blog/utils"
)

func NotFound(c *gin.Context) {
    cmp := t.NotFound()
    c.Status(http.StatusNotFound)
    t.Layout(cmp, "not-found").Render(c.Request.Context(), c.Writer)
}

// func Handler
func HandlerList(ctg string, title string) gin.HandlerFunc {
    return func(c *gin.Context) {
        htmxRequest := c.GetHeader("Hx-Request") == "true"

        if !htmxRequest {
            authorized := a.IsAuthorized(c)

            arts := u.GetArticlesList(ctg, 0)
            cmp := t.List(ctg, arts, authorized)

            c.Status(http.StatusOK)
            t.Layout(cmp, title).Render(c.Request.Context(), c.Writer)
        } else {
            cursorStr, exist := c.GetQuery("cursor")
            if !exist || cursorStr == "" {
                c.String(http.StatusInternalServerError, "something went terribly wrong...")
                c.Abort()
                return
            }

            cursor, err := strconv.Atoi(cursorStr)
            if err != nil {
                c.String(http.StatusInternalServerError, "something went terribly wrong...")
                c.Abort()
                return
            }

            arts := u.GetArticlesList(ctg, cursor)
            if len(arts) == 0 {
                c.Header("HX-Reswap", "outerHTML")
                c.String(http.StatusRequestedRangeNotSatisfiable, "you've reached the end...")
                c.Abort()
                return
            }

            c.Status(http.StatusOK)
            cmp := t.ArticleLoad(ctg, arts, cursor + 1)
            cmp.Render(c.Request.Context(), c.Writer)
        }
    }
}

// home is tech arts
func Home() gin.HandlerFunc {
    const ctg = "tech"
    const title = "blog"

    return HandlerList(ctg, title)
}

// entertainment is ent arts
func Entertainment() gin.HandlerFunc {
    const ctg = "ent"
    const title = "blog-ent"

    return HandlerList(ctg, title)
}

func Document(c *gin.Context) {
    art := c.Param("art")

    if art == "" || art == "/" {
        NotFound(c)
        c.Abort()
        return
    }

    content, found := GetLatestArticle(art)
    if !found {
        NotFound(c)
        c.Abort()
        return
    }

    audioName, found := GetArticleAudio(art)

    authorized := a.IsAuthorized(c)

    t.MdLayout(content, art, audioName, authorized).Render(c.Request.Context(), c.Writer)
    c.Status(http.StatusOK)
}
