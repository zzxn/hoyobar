package crypt

import (
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var passwordReg = regexp.MustCompile("^(?![0-9]+$)(?![a-z]+$)(?![A-Z]+$)(?![^0-9a-zA-Z]+$).{6,20}$")    

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
