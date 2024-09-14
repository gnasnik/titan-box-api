package api

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-box-api/core/dao"
	"io"
	"net/http"
	"strconv"
	"time"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	}
}

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := io.ReadAll(tee)
		c.Request.Body = io.NopCloser(&buf)
		log.Debug(string(body))
		//log.Debug(c.Request.Header)
		c.Next()
	}
}

func AuthorizationMiddlewareFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		ak := c.Request.Header.Get("ak")
		timestampStr := c.Request.Header.Get("timestamp")
		sign := c.Request.Header.Get("sign")

		if ak == "" || timestampStr == "" || sign == "" {
			c.JSON(http.StatusUnauthorized, nil)
			c.Abort()
			return
		}

		userKey, err := dao.GetUserKeyByAPIKey(c.Request.Context(), ak)
		if err != nil {
			c.JSON(http.StatusUnauthorized, nil)
			c.Abort()
			return
		}

		timestamp, _ := strconv.ParseInt(timestampStr, 10, 64)
		expectSign := generateMD5Hash(userKey.APISecret, userKey.Username, timestamp)
		if expectSign != sign {
			c.JSON(http.StatusUnauthorized, nil)
			c.Abort()
			return
		}

		if time.Unix(timestamp, 0).Add(5 * time.Minute).Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
