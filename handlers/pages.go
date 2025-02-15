package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	t "github.com/naufalsuryasumirat/blog/templates"
	u "github.com/naufalsuryasumirat/blog/utils"
)

func NotFound(c *gin.Context) {
    cmp := t.NotFound()
    c.Status(http.StatusNotFound)
    t.Layout(cmp, "not-found").Render(c.Request.Context(), c.Writer)
}

// home is tech arts
func Home(c *gin.Context) {
    arts := u.GetArticlesList("tech")
    cmp := t.List(arts)

    c.Status(http.StatusOK)
    t.Layout(cmp, "blog").Render(c.Request.Context(), c.Writer)
}

// entertainment is ent arts
func Entertainment(c *gin.Context) {
    arts := u.GetArticlesList("ent")
    cmp := t.List(arts)

    c.Status(http.StatusOK)
    t.Layout(cmp, "blog-ent").Render(c.Request.Context(), c.Writer)
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

    t.MdLayout(content).Render(c.Request.Context(), c.Writer)
    c.Status(http.StatusOK)
}
