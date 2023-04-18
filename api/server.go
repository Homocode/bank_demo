package api

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/homocode/bank_demo/db/sqlc"
)

type queries interface {
	AddAmountToAccountBalance(ctx context.Context, arg db.AddAmountToAccountBalanceParams) (db.Accounts, error)
	CreateAccount(ctx context.Context, arg db.CreateAccountParams) (db.Accounts, error)
	CreateEntry(ctx context.Context, arg db.CreateEntryParams) (db.Entries, error)
	CreateTransfer(ctx context.Context, arg db.CreateTransferParams) (db.Transfers, error)
	GetAccount(ctx context.Context, id int64) (db.Accounts, error)
	GetEntry(ctx context.Context, id int64) (db.Entries, error)
	GetTransfer(ctx context.Context, id int64) (db.Transfers, error)
	ListAccounts(ctx context.Context, arg db.ListAccountsParams) ([]db.Accounts, error)
	ListEntries(ctx context.Context, arg db.ListEntriesParams) ([]db.Entries, error)
	ListTransfers(ctx context.Context, arg db.ListTransfersParams) ([]db.Transfers, error)
}

var _ queries = (*db.Queries)(nil)

type Store interface {
	queries
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

	// Engine returns
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	const (
		pathAccounts = "/accounts"
		pathTransfer = "/transfer"
	)

	router.POST(fmt.Sprintf("%v", pathAccounts), server.createAccount)
	router.GET(fmt.Sprintf("%v/:id", pathAccounts), server.getAccount)
	router.GET(fmt.Sprintf("%v", pathAccounts), server.listAccounts)

	router.POST(fmt.Sprintf("%v", pathTransfer), server.transferBtwAccounts)

	server.router = router

	return server
}

// Runs the HTTP server on a specific address
func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
