package auth

import (
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	util "github.com/naufalsuryasumirat/blog/utils"
    t "github.com/naufalsuryasumirat/blog/templates"
)

var db *sql.DB
var key string
var muKey sync.RWMutex

func GenRandomSha() {
	t := uint64(time.Now().Unix())
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, t)

	muKey.Lock()
	defer muKey.Unlock()
	key = fmt.Sprintf("%x", sha256.Sum256(b))
}

func GetKey() string {
	muKey.RLock()
	defer muKey.RUnlock()
	return key
}

func CookieAuth(c *gin.Context) {
	key := GetKey()
	cookie, _ := c.Cookie("authkey")

	// intentionally ambiguous
	if cookie != key {
        cmp := t.NotFound()
        c.Status(http.StatusNotFound)
        t.Layout(cmp, "not-found").Render(c.Request.Context(), c.Writer)
		c.Abort()
		return
	}

	c.Next()
}

func IsAuthorized(c *gin.Context) bool {
    key := GetKey()
    cookie, _ := c.Cookie("authkey")

    return cookie == key
}

func init() {
	db = util.GetDB()
	GenRandomSha()
}
