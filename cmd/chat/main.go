package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Rasulikus/chat/internal/app"
	"github.com/Rasulikus/chat/internal/config"

	_ "github.com/Rasulikus/chat/docs" // swagger docs
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Chat API
// @version 1.0
// @description Simple chat service with rooms and WebSocket messaging.
// @BasePath /
func main() {
	cfg := config.LoadConfig()

	// Инициализируем Gin-роутер через наше приложение.
	router := app.App(cfg)

	// Регистрируем Swagger UI по пути /swagger/*any.
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	server := http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler: router,
	}

	log.Print("Server start at address: " + server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
