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

func ApiKeyMiddleware(q *sqlc.Queries, auth_conn pb.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		got := c.GetHeader("X-Clove-Key")
		if got == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing X-Clove-Key header"})
			return
		}

		// grpc call
		req := pb.AuthenticateKeyRequest{
			Key: got,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second * 30) 
		defer cancel()

		check, err := auth_conn.AuthenticateKey(ctx, &req)		
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else if !check.Success {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing X-Clove-Key header"})
			return
		}

		c.Set("api_key", got)
		c.Next()
	}
}
