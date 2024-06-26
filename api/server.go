package api

import (
	db "github.com/billalaashraf/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
)

type Server struct {
	store  *db.Store
	rotuer *gin.Engine
}

func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)
	router.PUT("/accounts", server.updateAccount)
	router.DELETE("/accounts/:id", server.deleteAccount)

	server.rotuer = router
	return server
}

func (server *Server) Start(address string) error {
	return server.rotuer.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
