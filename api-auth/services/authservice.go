package services

import (
	"fmt"
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	//"github.com/google/uuid"
	"net/http"
	//"github.com/jackc/pgx/v5/pgtype"
	"github.com/burkeclove/auth-api/internal"
	"github.com/burkeclove/auth-api/models/requests"
	"github.com/burkeclove/shared/db/sqlc"

	pb "github.com/burkeclove/shared/gen/go/protos"
)

type AuthService struct {
	Q *sqlc.Queries
	pb.UnimplementedAuthServiceServer
	JwtService *JwtService
}

func NewAuthService(q *sqlc.Queries) *AuthService {
	privatePEM, err := readPEMFile("../keys/private.pem")
	if err != nil {
		log.Fatal(err)
	}

	publicPEM, err := readPEMFile("../keys/public.pem")
	if err != nil {
		log.Fatal(err)
	}

	jwtSvc, err := NewJwtService(
		internal.Config{
			Issuer:    "clove-auth",
			Audience:  "clove-api",
			AccessTTL: 15 * time.Minute,
			Leeway:    30 * time.Second,
		},
		privatePEM,
		publicPEM,
	)
	if err != nil {
		log.Fatalf("an error occured while creating jwt svc %s", err.Error())
	}
	return &AuthService{Q: q, JwtService: jwtSvc}
}

func (a *AuthService) CreateApiKey(c *gin.Context) {
	log.Println("got a request to create api key")
	var req requests.CreateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println("creating api key with name: ", req.Name)
	//uuid := pgtype.UUID{Bytes: id, Valid: true}
	a.Q.CreateApiKey(context.Background(), sqlc.CreateApiKeyParams{
		Name: req.Name,
		//OrganizationID: pgtype.UUID,
	})
}

func (a *AuthService) GenerateSig4Keys(c *gin.Context) {

}

func (a *AuthService) GetApiKeys(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{})
}

// GRPC
func (a *AuthService) AuthenticateKey(ctx context.Context, req *pb.AuthenticateKeyRequest) (*pb.AuthenticateKeyResponse, error) {
	key := req.Key			
	
	getOrgRet, err := a.Q.GetOrgFromApiKey(ctx, key)
	if err != nil {
		return &pb.AuthenticateKeyResponse{
			Success:  false,
			ErrorMessage: err.Error(),
		}, err
	} else {
		return &pb.AuthenticateKeyResponse{
			Success: getOrgRet.ID.Valid,
			ErrorMessage: "",
			OrgId: getOrgRet.ID.String(),
		}, nil
	}
}

func (a *AuthService) AuthenticateJwt(ctx context.Context, req *pb.AuthenticateJwtRequest) (*pb.AuthenticateJwtResponse, error) {
	header := req.AuthorizationHeader	
	parts := strings.Split(header, " ")
	if len(parts) != 2 {
		return &pb.AuthenticateJwtResponse{
			Success: false,
			ErrorMessage: "Length of authorization header is not 2",
		}, errors.New("Length of authorization header is not 2")
	}
	jwt := parts[1]
	claims, err := a.JwtService.Validate(ctx, jwt)	
	if err != nil {
		return &pb.AuthenticateJwtResponse{
			Success:  false,
			ErrorMessage: err.Error(),
		}, err
	} else {
		return &pb.AuthenticateJwtResponse{
			Success: true,
			ErrorMessage: "",
			UserId: claims.UserID,
			Email: claims.Email,
		}, nil
	}
}

func (a *AuthService) CreateJwt(ctx context.Context, req *pb.CreateJwtRequest) (*pb.CreateJwtResponse, error) {
	if req.UserId == "" || req.Email == "" {
		return &pb.CreateJwtResponse{
			Success: false,
			ErrorMessage: "The jwt requires an email and user id",
		}, errors.New("The jwt requires an email and user id")
	}
	key, exp, err := a.JwtService.Mint(ctx, req.UserId, req.Email)
	if err != nil {
		errMsg := fmt.Sprintf("An error occured while creating a jwt: %s", err.Error())
		return &pb.CreateJwtResponse{
			Success: false,
			ErrorMessage: errMsg,
		}, err
	}
	return &pb.CreateJwtResponse{
		Success: true,
		Token: key,
		ExpiresAt: exp.String(),
	}, nil
}
