package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

type key string

var NonceKey key = "nonces"

type Nonces struct {
	Htmx            string
	ResponseTargets string
	Tw              string
	HtmxCSSHash     string
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func TextHTMLMiddleware(c *gin.Context) {
    c.Header("Content-Type", "text/html; charset=utf-8")
    c.Next()
}

func CSPMiddleware(c *gin.Context) {
    nonceSet := Nonces{
        Htmx:            generateRandomString(16),
        ResponseTargets: generateRandomString(16),
        Tw:              generateRandomString(16),
        HtmxCSSHash:     "sha256-pgn1TCGZX6O77zDvy0oTODMOxemn0oj0LeCnQTRj7Kg=",
    }

    // set nonces in context
    c.Set(string(NonceKey), nonceSet)
    ct := context.WithValue(c.Request.Context(), NonceKey, nonceSet)
    c.Request = c.Request.WithContext(ct)

    cspHeader := fmt.Sprintf(
        `default-src 'self';
            img-src 'self';
            script-src 'nonce-%s' 'nonce-%s';
            style-src 'nonce-%s' '%s'`,
    	nonceSet.Htmx,
    	nonceSet.ResponseTargets,
        nonceSet.Tw,
        nonceSet.HtmxCSSHash,
    )
    c.Header("Content-Security-Policy", cspHeader)

    c.Next()
}

func GetNonces(c context.Context) Nonces {
	nonceSet := c.Value(NonceKey)

	if nonceSet == nil {
		log.Fatal("error getting nonce set - is nil")
	}

	nonces, ok := nonceSet.(Nonces)

	if !ok {
		log.Fatal("error getting nonce set - not ok")
	}

	return nonces
}

func GetHtmxNonce(c context.Context) string {
	nonceSet := GetNonces(c)

	return nonceSet.Htmx
}

func GetResponseTargetsNonce(c context.Context) string {
	nonceSet := GetNonces(c)
	return nonceSet.ResponseTargets
}

func GetTwNonce(c context.Context) string {
	nonceSet := GetNonces(c)
	return nonceSet.Tw
}
