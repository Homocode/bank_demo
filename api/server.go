package api

import (
	"context"

	"github.com/gin-gonic/gin"
	db "github.com/homocode/bank_demo/db/sqlc"
)

type Store interface {
	db.Querier
	TransferTx(ctx context.Context, arg db.TransferTxParams) (db.TransferTxResult, error)
}

type Server struct {
	store  Store
	router *gin.Engine
}

// Creates a new HTTP server and setup routing
func NewServer(store Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts/", server.listAccount)

	server.router = router
	return server
}

// Runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
