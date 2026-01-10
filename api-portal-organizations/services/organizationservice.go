package services

import (
	"context"
	"log"
	"net/http"

	"github.com/burkeclove/organizations-api/models/requests"
	"github.com/burkeclove/organizations-api/models/responses"
	"github.com/burkeclove/shared/db/helpers"
	"github.com/burkeclove/shared/db/sqlc"
	pb "github.com/burkeclove/shared/gen/go/protos"
	"github.com/gin-gonic/gin"
)

type OrganizationService struct {
	Q *sqlc.Queries
	AuthClient pb.AuthServiceClient
}

func NewOrganizationService(q *sqlc.Queries, auth_conn pb.AuthServiceClient) *OrganizationService {
	return &OrganizationService{Q: q, AuthClient: auth_conn}
}

func (o *OrganizationService) CreateOrganization(c *gin.Context) {
	userId, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	userUUID, err := helpers.UUIDFromString(userId.(string))

	var req requests.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// make sure organization doesn't exist with that nameo
	// later lol

	log.Println("creating org with name: ", req.Name)
	//uuid := pgtype.UUID{Bytes: id, Valid: true}
	org, err := o.Q.CreateOrganization(context.Background(), req.Name) 
	if err != nil {
		log.Println("an error occured while creating organization", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// create the organization user
	_, err = o.Q.CreateOrganizationUser(c.Request.Context(), sqlc.CreateOrganizationUserParams{
		OrganizationID: org.ID,
		UserID: userUUID,
	})
	if err != nil {
		log.Println("could not create org user. error : ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	} 

	// create api key
	log.Println("about to create api key... org id is: ", org.ID)
	createKeyReq := pb.CreateKeyRequest{
		OrganizationId: org.ID.String(),
	}

	log.Println("request has been formed")
	res, err := o.AuthClient.CreateKey(c.Request.Context(), &createKeyReq)
	if err != nil {
		log.Println("an error occured while creating api key during org creation: ", err.Error())
		createOrgResponse := responses.CreateOrganizationResponse{
			Name: org.Name,
			Id: org.ID.String(),
		}
		c.JSON(http.StatusCreated, createOrgResponse)
		return
	}

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
	userId, exists := c.Get("user_id")
	orgId := c.Param("orgId")
	if !exists {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	// 
	req := &pb.CheckUserOrganizationRequest{
		UserId: userId.(string),
		OrganizationId: orgId,
	}
	authRes, err := o.AuthClient.CheckUserOrganization(c.Request.Context(), req)
	if err != nil || authRes.Success {
		log.Println("an error occured while getting user organization: ", err.Error())
		c.JSON(http.StatusForbidden, gin.H{"error":err.Error()})
	}

	uuid, err := helpers.UUIDFromString(req.OrganizationId)
	org, err := o.Q.GetOrganizationByID(c.Request.Context(), uuid)
	if err != nil {
		log.Println("an error occured while getting uuid from string", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
	}

	c.JSON(http.StatusCreated, gin.H{"data": org})
}

func (o *OrganizationService) GetOrganizations(c *gin.Context) {
	userId, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	log.Println("user id: ", userId)
	uuid, err := helpers.UUIDFromString(userId.(string))
	if err != nil {
		log.Println("an error occured while getting uuid from string", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
	}

	orgs, err := o.Q.GetOrganizationsFromUserId(c.Request.Context(), uuid)
	if err != nil {
		log.Println("an error occured while getting organizations from user id", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"data": orgs})
}
