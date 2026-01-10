package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	//"github.com/google/uuid"
	"net/http"
	//"github.com/jackc/pgx/v5/pgtype"
	"github.com/burkeclove/auth-api/internal"
	"github.com/burkeclove/auth-api/models/requests"
	"github.com/burkeclove/shared/db/helpers"
	"github.com/burkeclove/shared/db/sqlc"
	"github.com/burkeclove/auth-api/functions/passwords"

	pb "github.com/burkeclove/shared/gen/go/protos"
)

type AuthService struct {
	Q *sqlc.Queries
	pb.UnimplementedAuthServiceServer
	JwtService *JwtService
}

func NewAuthService(q *sqlc.Queries) *AuthService {
	privatePEM, err := readPEMFile("./keys/private.pem")
	if err != nil {
		log.Fatal(err)
	}

	publicPEM, err := readPEMFile("./keys/public.pem")
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

	log.Printf("creating api key with name: %s and for org: %s", req.Name, req.OrgId)
	uuid, err := helpers.UUIDFromString(req.OrgId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	a.Q.CreateApiKey(context.Background(), sqlc.CreateApiKeyParams{
		Name: req.Name,
		OrganizationID: uuid,
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


func (a *AuthService) CheckUserOrganization(ctx context.Context, req *pb.CheckUserOrganizationRequest) (*pb.CheckUserOrganizationResponse, error) {
	log.Printf("checking user organization for user id: %s and org id: %s: ", req.UserId, req.OrganizationId)
	userId, err := helpers.UUIDFromString(req.UserId)
	if err != nil {
		errMsg := fmt.Sprintf("an error occured while getting uuid from user id %s", err.Error())
		return &pb.CheckUserOrganizationResponse{
			Success: false,
			ErrorMessage: errMsg,
		}, err
	}
	
	orgId, err := helpers.UUIDFromString(req.OrganizationId)
	if err != nil {
		errMsg := fmt.Sprintf("an error occured while getting uuid from org id %s", err.Error())
		return &pb.CheckUserOrganizationResponse{
			Success: false,
			ErrorMessage: errMsg,
		}, err
	}


	ret, err := a.Q.CheckOrganizationUserExists(ctx, sqlc.CheckOrganizationUserExistsParams{
		UserID: userId,
		OrganizationID: orgId,
	})

	if err != nil {
		errMsg := fmt.Sprintf("an error occured checking if user %s belongs to org id %s, err: %s", req.UserId, req.OrganizationId, err.Error())
		return &pb.CheckUserOrganizationResponse{
			Success: false,
			ErrorMessage: errMsg,
		}, err
	}
	return &pb.CheckUserOrganizationResponse{
		Success: true,
		Check: ret,
	}, nil
}

func (a *AuthService) CreateKey(ctx context.Context, req *pb.CreateKeyRequest) (*pb.CreateKeyResponse, error) {
	log.Println("creating api key for org: ", req.OrganizationId)
	uuid, err := helpers.UUIDFromString(req.OrganizationId)
	if err != nil {
		return &pb.CreateKeyResponse{
			Success: false,
			ErrorMessage: err.Error(),
		}, err
	}

	_, err = a.Q.CreateApiKey(context.Background(), sqlc.CreateApiKeyParams{
		Name: "First Key",
		OrganizationID: uuid,
	})
	if err != nil {
		return &pb.CreateKeyResponse{
			Success: false,
			ErrorMessage: err.Error(),
		}, err
	}
	return &pb.CreateKeyResponse{
		Success: true,
		KeyId: "",
		Key: "",
	}, nil
}

func (a *AuthService) HashPassword(ctx context.Context, req *pb.HashPasswordRequest) (*pb.HashPasswordResponse, error) {
	log.Println("hashing password")
	passwordHash, err := passwords.HashPassword(req.Password, passwords.DefaultParams)
	if err != nil {
		return &pb.HashPasswordResponse{
			Success: false,
			ErrorMessage: err.Error(),
		}, err
	}

	return &pb.HashPasswordResponse{
		Success: true,
		PasswordHash: passwordHash,
	}, nil
}

func (a *AuthService) Login(c *gin.Context) {
	var req requests.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// hash password
	password_hash, err := passwords.HashPassword(req.Password, passwords.DefaultParams)
	if err != nil {
		log.Println("an error occured while hashing password: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// get user
	user, err := a.Q.GetUserByEmailPassword(c.Request.Context(), sqlc.GetUserByEmailPasswordParams{
		Email: req.Email,
		PasswordHash: password_hash,
	})		
	if err != nil {
		log.Println("an error occured while getting user by email password: ", err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	key, _, err := a.JwtService.Mint(c.Request.Context(), user.ID.String(), user.Email)
	if err != nil {
		errMsg := fmt.Sprintf("An error occured while creating a jwt: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})	
		return
	}

	// generate jwt
	c.JSON(http.StatusUnauthorized, gin.H{"user": user, "jwt": key})
}
