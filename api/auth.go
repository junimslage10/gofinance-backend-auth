package api

import (
	"bytes"
	"crypto/sha512"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	db "github.com/junimslage10/gofinance-backend-auth/db/sqlc"
	"github.com/junimslage10/gofinance-backend-auth/infrastucture/kafka"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type UseCaseAuthentication struct {
	KafkaProducer kafka.KafkaProducer
}

func (server *Server) login(ctx *gin.Context) {
	var req loginRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	hashedInput := sha512.Sum512_256([]byte(req.Password))
	trimmedHash := bytes.Trim(hashedInput[:], "\x00")
	preparedPassword := string(trimmedHash)
	plainTextInBytes := []byte(preparedPassword)
	hashTextInBytes := []byte(user.Password)
	err = bcrypt.CompareHashAndPassword(hashTextInBytes, plainTextInBytes)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	expirationTime := time.Now().Add(3 * time.Minute)

	claims := &Claims{
		Username: req.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	generatedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	var jwtSignedKey = []byte("secret_key")
	generatedTokenToString, err := generatedToken.SignedString(jwtSignedKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authorizationTokenDto := db.User{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: time.Now(),
	}
	authorizationTokenJson, err := json.Marshal(authorizationTokenDto)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	var u UseCaseAuthentication
	err = u.KafkaProducer.Publish(string(authorizationTokenJson), "generate_tokens")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	fmt.Println(err)
	ctx.JSON(http.StatusOK, generatedTokenToString)

}
