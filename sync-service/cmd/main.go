package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/Hoglandets-IT/go-syncService/internal/utils"
)

func main() {

	os.Setenv("SYNC_SECRET", "testsecret123")
	os.Setenv("SYNC_ADDR", "0.0.0.0:8080")

	secret := os.Getenv("SYNC_SECRET")
	if secret == "" {
		panic("unable to start service, missing env \"SYNC_SECRET\".")
	}

	addr := os.Getenv("SYNC_ADDR")
	if addr == "" {
		panic("unable to start service, missing env \"SYNC_ADDR\".")
	}

	r := gin.Default()

	r.POST("/sync", utils.HandleSyncRequest(secret))

	r.Run(addr)
}