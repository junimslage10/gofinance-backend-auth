package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/junimslage10/gofinance-backend-auth/db/sqlc"
)

type Server struct {
	store  *db.SQLStore
	router *gin.Engine
}

func NewServer(store *db.SQLStore) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// Rotas
	router.POST("/login", server.login)

	server.router = router
	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"Api as error": err.Error()}
}
