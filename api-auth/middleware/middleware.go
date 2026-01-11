package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	//"crypto/subtle"
	"github.com/burkeclove/auth-api/internal"
	"github.com/burkeclove/shared/db/sqlc"
	//pb "github.com/burkeclove/shared/gen/go/protos"
)

func PortalMiddleware(q *sqlc.Queries, validate func(ctx context.Context, token string) (*internal.Claims, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		// look for api key first
		got := c.GetHeader("Authorization")
		if got == "" {
			log.Println("auth header was empty")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second * 30) 
		defer cancel()

		jwt := strings.Split(got, " ")
		if len(jwt) != 2 {
			log.Println("auth header was empty")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid Authorization header"})
			return
		}

		claims, err := validate(ctx, jwt[1])
		if err != nil {
			log.Println("could not get jwt from auth header: ", err.Error())

		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}
