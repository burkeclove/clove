package services

import (
	"errors"
	"log"
	"strings"

	"github.com/burkeclove/shared/db/helpers"
	"github.com/jackc/pgx/v5/pgtype"
)

// like "Bearer ldjslajdlfj"
func GetJwtFromAuthorizationHeader(header string) (string, error) {
	parts := strings.Split(header, " ")
	parts_len := len(parts)
	if parts_len != 2 {
		log.Printf("while in gt jwt from authorization header, parts size was not 2 (was %d)", parts_len)
		return "", errors.New("while in gt jwt from authorization header, parts size was not")
	}
	return parts[1], nil
}

func GetUserIdOrgId(orgId string, userId string) (pgtype.UUID, pgtype.UUID, error) {
	orgUUID, err := helpers.UUIDFromString(orgId)
	if err != nil {
		log.Println("could not get uuid from org id. err: ", err.Error())
		return pgtype.UUID{}, pgtype.UUID{}, err
	}
	userUUID, err := helpers.UUIDFromString(userId)
	if err != nil {
		log.Println("could not get uuid from user id. err: ", err.Error())
		return pgtype.UUID{}, pgtype.UUID{}, err
	}
	return orgUUID, userUUID, nil
	
}
