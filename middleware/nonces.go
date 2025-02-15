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

func CSPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new Nonces struct for every request when here.
		// move to outside the handler to use the same nonces in all responses
		nonceSet := Nonces{
			Htmx:            generateRandomString(16),
			ResponseTargets: generateRandomString(16),
			Tw:              generateRandomString(16),
			HtmxCSSHash:     "sha256-pgn1TCGZX6O77zDvy0oTODMOxemn0oj0LeCnQTRj7Kg=",
		}

		// set nonces in context
		c.Set(string(NonceKey), nonceSet)
		// c := context.WithValue(r.Context(), NonceKey, nonceSet)

		// insert the nonces into the content security policy header
		// cspHeader := fmt.Sprintf(
		//     `default-src 'self';
		//         img-src 'self' https://img.youtube.com;
		//         script-src 'nonce-%s' 'nonce-%s';
		//         style-src 'nonce-%s' '%s';`, // %s
		//     nonceSet.Htmx,
		//     nonceSet.ResponseTargets,
		//     nonceSet.Tw,
		//     nonceSet.HtmxCSSHash)

		// FIXME: actually properly set the required security
		cspHeader := fmt.Sprintf(`default-src *;`)

		// cspHeader := fmt.Sprintf(
		//           `default-src 'self';
		//               img-src 'self' https://img.youtube.com;
		//               script-src 'nonce-%s' 'nonce-%s';`,
		// 	nonceSet.Htmx,
		// 	nonceSet.ResponseTargets)
		// w.Header().Set("Content-Security-Policy", cspHeader)
		c.Header("Content-Security-Policy", cspHeader)

		c.Next()
	}
}

func TextHTMLMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Next()
	}
}

// get the Nonce from the context, it is a struct called Nonces,
// so we can get the nonce we need by the key, i.e. HtmxNonce
func GetNonces(c context.Context) Nonces {
	nonceSet := c.Value(NonceKey)
	return Nonces{}

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
