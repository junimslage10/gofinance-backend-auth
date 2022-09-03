package api

import (
	"bytes"
	"crypto/sha512"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "fmt"
	_ "log"
	"net/http"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	_ "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	db "github.com/junimslage10/gofinance-backend-auth/db/sqlc"
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

	// Apache Kafka ---
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": os.Getenv("KAFKA_BROKERCONNECT_HOST")})
	if err != nil {
		panic(err)
	}

	defer p.Close()

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	// Produce messages to topic (asynchronously)
	topic := "logins"
	messageText := &db.User{
		ID:        user.ID,
		Username:  user.Username,
		Password:  user.Password,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
	messageTextJson, _ := json.Marshal(messageText)
	p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(string(messageTextJson)),
	}, nil)

	// Wait for message deliveries before shutting down
	// Flush and close the producer and the events channel
	for p.Flush(1000) > 0 {
		fmt.Print("Still waiting to flush outstanding messages\n", p)
	}

	ctx.JSON(http.StatusOK, generatedTokenToString)
}
