package middleware

import (
	"mime"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const msgUnssuportedContentType = "(Unsupported Media Type) response. Please provide application/json."

func ContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		if hasContentType(c.ContentType(), "application/json") {
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{"message": msgUnssuportedContentType})
		}
	}
}

func hasContentType(contentType, mimetype string) bool {
	if contentType == "" {
		return false
	}
	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		if t == mimetype {
			return true
		}
	}
	return false
}
