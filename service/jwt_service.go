package service

import (
	"errors"
	"fmt"
	"time"

	jwt "gopkg.in/dgrijalva/jwt-go.v3"
)

// 定义自定义声明结构体
type CustomClaims struct {
	jwt.StandardClaims
	UserID uint `json:"user_id"`
}

// JWTController 结构体
type JWTController struct {
	SecretKey []byte // 密钥，用于签发和验证 Token
	Salt      string // 随机字符串，用于加盐
}

// NewJWTController 函数，用于创建一个新的 JWTController 实例
func NewJWTController(secretKey, salt string) (*JWTController, error) {
	if secretKey == "" || salt == "" {
		return nil, errors.New("secretKey and salt are required")
	}
	return &JWTController{
		SecretKey: []byte(secretKey),
		Salt:      salt,
	}, nil
}

// GenerateToken 函数，用于生成 JWT Token
func (c *JWTController) GenerateToken(userID uint, expireDuration time.Duration) (string, error) {
	now := time.Now()
	expireTime := now.Add(expireDuration).Unix()

	claims := &CustomClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime,
			Issuer:    "jwt",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(c.SecretKey)
}

// VerifyToken 函数，用于验证 JWT Token
func (c *JWTController) VerifyToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return c.SecretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("invalid token")
	}
}

// RefreshToken 函数，用于刷新 JWT Token
func (c *JWTController) RefreshToken(tokenString string, expireDuration time.Duration) (string, error) {
	claims, err := c.VerifyToken(tokenString)
	if err != nil {
		return "", err
	}

	now := time.Now()
	expireTime := now.Add(expireDuration).Unix()

	claims.StandardClaims.ExpiresAt = expireTime
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(c.SecretKey)
}
