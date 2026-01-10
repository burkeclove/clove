package services

import (
	"context"
	"log"
	"net/http"
	"github.com/burkeclove/shared/db/helpers"
	"github.com/burkeclove/shared/db/sqlc"
	pb "github.com/burkeclove/shared/gen/go/protos"
	"github.com/burkeclove/users-api/models/requests"
	"github.com/burkeclove/users-api/models/responses"
	"github.com/gin-gonic/gin"
)

type UserService struct {
	Q *sqlc.Queries
	AuthClient pb.AuthServiceClient
}

func NewUserService(q *sqlc.Queries, auth_conn pb.AuthServiceClient) *UserService {
	return &UserService{Q: q, AuthClient: auth_conn}
}

func (o *UserService) CreateUser(c *gin.Context) {
	var req requests.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println("creating user with email: ", req.Email)

	hashPassReq := pb.HashPasswordRequest{
		Password: req.Password,	
	}
	hashPassRes, err := o.AuthClient.HashPassword(c.Request.Context(), &hashPassReq)
	if err != nil || !hashPassRes.Success {
		log.Fatalf("an error occured while hashing the password: %s", err.Error())
	}

	user, err := o.Q.CreateUser(context.Background(), sqlc.CreateUserParams{
		Email: req.Email,
		PasswordHash: hashPassRes.PasswordHash,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// create jwt
	jwtReq := pb.CreateJwtRequest{
		UserId: user.ID.String(),
		Email: req.Email,
	}
	jwtRes, err := o.AuthClient.CreateJwt(c.Request.Context(), &jwtReq)

	res := responses.CreateUserResponse{
		Id: user.ID.String(),
		Jwt: jwtRes.Token,
	}

	c.JSON(http.StatusCreated, res)
}

func (o *UserService) GetUser(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	jwtReq := pb.AuthenticateJwtRequest{
		AuthorizationHeader: authHeader,
	}
	jwtRes, err := o.AuthClient.AuthenticateJwt(c.Request.Context(), &jwtReq)

	if !jwtRes.Success {
		c.JSON(http.StatusUnauthorized, gin.H{"error": jwtRes.ErrorMessage})
		return
	}

	uuid, err := helpers.UUIDFromString(jwtRes.UserId)
	if err != nil {
		log.Println("could not get uuid from user id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	org, err := o.Q.GetUserById(c.Request.Context(), uuid)
	if err != nil {
		log.Println("an error occured while getting organization from api key")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}	

	c.JSON(http.StatusCreated, gin.H{"data": org})
}
