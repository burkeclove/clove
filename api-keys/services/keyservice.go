package services

import (
	"net/http"
	"github.com/burkeclove/shared/db/sqlc"
	pb "github.com/burkeclove/shared/gen/go/protos"
	"github.com/gin-gonic/gin"
)

type KeyService struct {
	Q *sqlc.Queries
	AuthClient pb.AuthServiceClient
}

func NewKeyService(q *sqlc.Queries, auth_conn pb.AuthServiceClient) *KeyService {
	return &KeyService{Q: q, AuthClient: auth_conn}
}

func (o *KeyService) WhoAmI(c *gin.Context) {
	org, exists := c.Get("org_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}
	c.JSON(http.StatusOK, gin.H{"org": org})
}
