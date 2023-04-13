package api

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	// This is to make less verbose the console output from gin each time a the server runs for a test.
	// It`s to have a cleaner view of the test logs on the terminal.
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
