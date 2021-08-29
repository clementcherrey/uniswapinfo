package main

import (
	"flag"
	"log"

	"uniswapinfo/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	addr := flag.String("addr", ":8080", "address")
	mode := flag.String("mode", gin.TestMode, "gin mode")
	flag.Parse()

	gin.SetMode(*mode)
	router := gin.Default()

	asset := router.Group("/asset")
	asset.GET("/:id/pools", handlers.GetPools)
	asset.GET("/:id/volume", handlers.GetVolume)

	block := router.Group("/block")
	block.GET("/:number/swaps", handlers.GetSwaps)
	// block.GET("/block/:number/assets", nil)

	log.Printf("Listening and serving HTTP on %s", *addr)
	log.Fatal(router.Run(*addr))
}
