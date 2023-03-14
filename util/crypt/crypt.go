package crypt

import (
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// TODO: test it
var passwordReg = regexp.MustCompile(`^[0-9A-Za-z\W_]*([0-9][A-Za-z\W_]*[a-zA-Z]|[a-zA-Z][0-9\W_]*[0-9]|[0-9\W_]*[a-zA-Z][A-Za-z\W_]*|[a-zA-Z\W_]*[0-9][0-9A-Za-z\W_]*){6,20}$`)

func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h), err
}

func CompareHashAndPassword(hash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CheckPasswordStrength(password string) bool {
	return passwordReg.MatchString(password)
}
