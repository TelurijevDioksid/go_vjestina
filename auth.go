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
    return encodeBase64(header) + "." + encodeBase64(payload) + "." + encodeBase64(signature)
}

func GenerateJwtHeader() []byte {
    headerJson, _ := json.Marshal(map[string]string{
        "alg": "HS256",
        "typ": "JWT",
    })
    return headerJson
}

func GenerateJwtPayload(email string) []byte {
    exp := time.Now().Add(time.Hour).Unix()
    payloadJson, _ := json.Marshal(map[string]string{
        "exp": strconv.FormatInt(exp, 10),
        "email": email,
    })
    return payloadJson
}

func GenerateJwtSignature(header, payload []byte) []byte {
    secret := os.Getenv("JWT_SECRET")
    message := encodeBase64(header) + "." + encodeBase64(payload)
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(message))
    return mac.Sum(nil)
}

func encodeBase64(data []byte) string {
    return base64.StdEncoding.EncodeToString(data)
}

func decodeBase64(data string) ([]byte, error) {
    res, err := base64.StdEncoding.DecodeString(data)
    if err != nil {
        return nil, nil
    }
    return res, nil
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

    exp, err := strconv.ParseInt(claims["exp"], 10, 64)
    if err != nil || time.Now().Unix() > exp {
        return false
    }

    decHeader, err := base64.StdEncoding.DecodeString(parts[0])
    if err != nil {
        return false
    }

    decPayload, err := base64.StdEncoding.DecodeString(parts[1])
    if err != nil {
        return false
    }
    
    testSignature := GenerateJwtSignature(decHeader, decPayload)
    return hmac.Equal([]byte(parts[2]), testSignature)
}
