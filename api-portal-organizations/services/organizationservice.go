package services

import (
	"context"
	"log"
	"net/http"
	pb "github.com/burkeclove/shared/gen/go/protos"
	"github.com/burkeclove/organizations-api/models/requests"
	"github.com/burkeclove/organizations-api/models/responses"
	"github.com/gin-gonic/gin"
	"github.com/burkeclove/shared/db/sqlc"
)

type OrganizationService struct {
	Q *sqlc.Queries
	AuthClient pb.AuthServiceClient
}

func NewOrganizationService(q *sqlc.Queries, auth_conn pb.AuthServiceClient) *OrganizationService {
	return &OrganizationService{Q: q, AuthClient: auth_conn}
}

func (o *OrganizationService) CreateOrganization(c *gin.Context) {
	var req requests.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println("creating org with name: ", req.Name)
	//uuid := pgtype.UUID{Bytes: id, Valid: true}
	org, err := o.Q.CreateOrganization(context.Background(), req.Name) 
	if err != nil {
		log.Println("an error occured while creating organization", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// create api key
	log.Println("about to create api key... org id is: ", org.ID)
	createKeyReq := pb.CreateKeyRequest{
		OrganizationId: org.ID.String(),
	}

	log.Println("request has been formed")
	res, err := o.AuthClient.CreateKey(c.Request.Context(), &createKeyReq)

	id := res.KeyId
	key := res.Key
	
	log.Println("key created... key id:", id)
	createOrgResponse := responses.CreateOrganizationResponse{
		Name: org.Name,
		Id: org.ID.String(),
		ApiKeyId: id,
		ApiKey: key,
	}

	c.JSON(http.StatusCreated, createOrgResponse)
}

func (o *OrganizationService) GetOrganizationById(c *gin.Context) {
	apiKey, exists := c.Get("apiKey")
	if !exists {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	key := apiKey.(string)

	org, err := o.Q.GetOrganizationFromApiKey(c.Request.Context(), key)
	if err != nil {
		log.Println("an error occured while getting organization from api key")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusCreated, gin.H{"data": org})
}
