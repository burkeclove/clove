package services

import (
	"errors"
	"log"
	"strings"
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
