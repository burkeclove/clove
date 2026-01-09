package middleware

import (
	"time"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	//"crypto/subtle"
	"github.com/burkeclove/shared/db/sqlc"
	pb "github.com/burkeclove/shared/gen/go/protos"
)

func PortalMiddleware(q *sqlc.Queries, auth_conn pb.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		got := c.GetHeader("Authorization")
		if got == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		// grpc call
		req := pb.AuthenticateJwtRequest{
			AuthorizationHeader: got,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second * 30) 
		defer cancel()

		check, err := auth_conn.AuthenticateJwt(ctx, &req)		
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else if !check.Success {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		c.Set("user_id", check.UserId)
		c.Set("email", check.Email)
		c.Next()
	}
}
