package hashing

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string, cost ...int) ([]byte, error) {
	hc := bcrypt.DefaultCost
	if len(cost) > 0 {
		hc = cost[0]
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), hc)
	return bytes, err
}

func CheckPasswordHash(hash []byte, password []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, password)
	return err == nil
}
