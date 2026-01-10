
package middleware

import (
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/burkeclove/shared/db/sqlc"
	pb "github.com/burkeclove/shared/gen/go/protos"
)

func SigV4Middleware(q *sqlc.Queries, auth_conn pb.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		
		// authenticate
		authReq :=  &pb.AuthenticateSigV4Request{
			AuthorizationHeader: authHeader,		
		}
		authRes, err := auth_conn.AuthenticateSigV4(c.Request.Context(), authReq)
		if err != nil {
			log.Println("an error occured while authenticating sigv4. err: ", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set("org_id", authRes.OrgId)
		c.Next()
	}
}
