package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Rasulikus/chat/internal/api/http"
	"github.com/Rasulikus/chat/internal/api/ws"
	"github.com/Rasulikus/chat/internal/config"
	"github.com/Rasulikus/chat/internal/repository"
	messageRepo "github.com/Rasulikus/chat/internal/repository/message"
	roomRepo "github.com/Rasulikus/chat/internal/repository/room"
	"github.com/Rasulikus/chat/internal/service"
	"github.com/Rasulikus/chat/internal/service/message"
	"github.com/Rasulikus/chat/internal/service/room"
	wsruntime "github.com/Rasulikus/chat/internal/ws"
	"github.com/gin-gonic/gin"
)

// App initializes the application dependencies, configures routes, and returns the Gin engine.
func App(cfg *config.Config) *gin.Engine {

	db, err := repository.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	roomRepository := roomRepo.NewRepository(db.DB)
	roomService := room.NewService(roomRepository)
	roomHandler := http.NewRoomHandler(roomService)

	msgRepository := messageRepo.NewRepository(db.DB)
	msgService := message.NewService(msgRepository)

	hub := wsruntime.NewHub()
	go hub.Run()

	wsHandler := ws.NewWSHandler(hub, roomService, msgService)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	startRoomCleanup(ctx, roomService)

	router := gin.Default()

	roomApi := router.Group("/rooms")
	{
		roomApi.POST("", roomHandler.Create)
		roomApi.GET("", roomHandler.List)
		roomApi.GET("/:id", roomHandler.GetByID)
	}
	wsApi := router.Group("/ws")
	{
		wsApi.GET("", wsHandler.HandleWS)
	}

	return router
}

// startRoomCleanup launches a background job that periodically soft deletes inactive rooms.
func startRoomCleanup(ctx context.Context, roomService service.RoomService) {
	ticker := time.NewTicker(time.Hour)
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				affected, err := roomService.SoftDeleteInactiveOlderThan(context.Background(), 7*24*time.Hour)
				if err != nil {
					log.Println("room cleanup error:", err)
					continue
				}
				if affected > 0 {
					log.Printf("room cleanup: soft-deleted %d inactive rooms\n", affected)
				}
			}
		}
	}()
}
