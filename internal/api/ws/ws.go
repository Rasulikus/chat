package ws

import (
	"log"
	"net/http"

	"github.com/Rasulikus/chat/internal/service"
	wsruntime "github.com/Rasulikus/chat/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WSHandler struct {
	hub            *wsruntime.Hub
	roomService    service.RoomService
	messageService service.MessageService
}

func NewWSHandler(hub *wsruntime.Hub, roomService service.RoomService, messageService service.MessageService) *WSHandler {
	return &WSHandler{
		hub:            hub,
		roomService:    roomService,
		messageService: messageService,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWS upgrades the HTTP connection to a WebSocket and attaches the client to the hub.
//
// @Summary WebSocket endpoint
// @Description
// @Description Upgrades the HTTP connection to a WebSocket for real-time chat.
// @Description
// @Description WebSocket message protocol (JSON):
// @Description   Incoming events:
// @Description     - type: "join" | "message" | "load_history"
// @Description     - room_id: number (for "join")
// @Description     - nick: string (for "join")
// @Description     - password: string (for "join")
// @Description     - text: string (for "message")
// @Description     - before_id: number (for "load_history")
// @Description
// @Description   Outgoing events:
// @Description     - type: "message" | "history" | "join" | "error"
// @Description     - room_id: number
// @Description     - nick: string
// @Description     - message: Message (for "message")
// @Description     - messages: Message[] (for "load_history")
// @Description     - text: string (for "error")
// @Tags ws
// @Produce json
// @Success 101 "Switching Protocols"
// @Router /ws [get]
func (h *WSHandler) HandleWS(c *gin.Context) {

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("HandleWS: upgrader.Upgrade error:", err)
		return
	}

	client := wsruntime.NewClient(h.hub, conn, h.roomService, h.messageService)
	client.Start()
}
