package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "github.com/naufalsuryasumirat/blog/auth"
     _ "github.com/naufalsuryasumirat/blog/templates"
)

// sets auth-cookie(s) for ping-pong/ding-dong verification

func resetCookie(c *gin.Context) {
	c.SetCookie("authkey", "lorem", -1, "/", "192.168.0.166", false, false)
}

func setAuthCookie(c *gin.Context, s string) {
    const limit = 7200 // 2-hours
	c.SetCookie("authkey", s, limit, "/", "192.168.0.166", false, false)
}

func Ping(c *gin.Context) {
    // cookie len actually defaults to zero
    cookie, _ := c.Cookie("authkey")
    defer c.String(http.StatusOK, "pong")

    key := auth.GetKey()
    if len(cookie) >= len(key) {
        resetCookie(c)
        c.Abort()
        return
    }

    setAuthCookie(c, cookie+key[:len(key)/2])
}

func Pong(c *gin.Context) {
    cookie, _ := c.Cookie("authkey")
    defer c.String(http.StatusOK, "bong")

    key := auth.GetKey()
    if len(cookie) >= len(key) {
        resetCookie(c)
        c.Abort()
        return
    }

    setAuthCookie(c, cookie+key[len(key)/2:])
}
