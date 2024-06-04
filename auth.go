package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func BcryptPassword(pwd string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(hash), nil
}

func ValidatePassword(encryPwd string, pwd string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(encryPwd), []byte(pwd))
    return err == nil
}

func GenerateJwtToken(email string) string {
    header := GenerateJwtHeader()
    payload := GenerateJwtPayload(email)
    signature := GenerateJwtSignature(header, payload)
    return header + "." + payload + "." + signature
}

func GenerateJwtHeader() string {
    headerJson, _ := json.Marshal(map[string]string{
        "alg": "HS256",
        "typ": "JWT",
    })
    return base64.StdEncoding.EncodeToString(headerJson)
}

func GenerateJwtPayload(email string) string {
    exp := time.Now().Add(time.Hour).Unix()
    payloadJson, _ := json.Marshal(map[string]string{
        "exp": strconv.FormatInt(exp, 10),
        "email": email,
    })
    return base64.StdEncoding.EncodeToString(payloadJson)
}

func GenerateJwtSignature(header, payload string) string {
    secret := os.Getenv("JWT_SECRET")
    message := header + "." + payload
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(message))
    signature := mac.Sum(nil)
    return base64.StdEncoding.EncodeToString(signature)
}

func ValidateJwt(token string) bool {
    parts := strings.Split(token, ".")
    if len(parts) != 3 {
        return false
    }

    claims := make(map[string]string)
    payload, _ := base64.StdEncoding.DecodeString(parts[1])
    if err := json.Unmarshal(payload, &claims); err != nil {
        return false
    }

    exp, _ := strconv.ParseInt(claims["exp"], 10, 64)
    if time.Now().Unix() > exp {
        return false
    }

    signature := parts[2]
    return signature == GenerateJwtSignature(parts[0], parts[1])
}
