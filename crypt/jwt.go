package crypt

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
)

//face info
type Jwt struct {
	secret string `secret key string`
	token *jwt.Token `jwt token instance`
	claims jwt.MapClaims `jwt claims object`
}

//construct
func NewJwt(securityKeys ...string) *Jwt {
	securityKey := SecretKeyOfJwt
	if securityKeys != nil && len(securityKeys) > 0 {
		securityKey = securityKeys[0]
	}
	this := &Jwt{
		secret: securityKey,
		token:jwt.New(jwt.SigningMethodHS256),
		claims:make(jwt.MapClaims),
	}
	return this
}

//encode
func (j *Jwt) Encode(input map[string]interface{}) (string, error) {
	j.claims = input
	j.token.Claims = j.claims
	result, err := j.token.SignedString([]byte(j.secret))
	if err != nil {
		return "", err
	}
	return result, nil
}

//decode
func (j *Jwt) Decode(input string) (map[string]interface{}, error) {
	//parse input string
	token, err := jwt.Parse(input, j.getValidationKey)
	if err != nil {
		return nil, err
	}
	//check header
	if jwt.SigningMethodHS256.Alg() != token.Header["alg"] {
		return nil, errors.New("header error")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

//set key
func (j *Jwt) SetKey(key string) error {
	if key == "" {
		return errors.New("invalid parameter")
	}
	j.secret = key
	return nil
}

//get validate key
func (j *Jwt) getValidationKey(*jwt.Token) (interface{}, error) {
	return []byte(j.secret), nil
}