package ws

import (
	"context"
	"log"
	"sync"

	"github.com/Rasulikus/chat/internal/model"
	"github.com/Rasulikus/chat/internal/service"
	"github.com/gorilla/websocket"
)

type Client struct {
	Nick   string
	RoomID int64

	hub            *Hub
	conn           *websocket.Conn
	roomService    service.RoomService
	messageService service.MessageService

	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once

	send chan OutgoingEvent
}

// NewClient constructs a new WebSocket client bound to a hub and room/message services.
func NewClient(h *Hub, conn *websocket.Conn, roomService service.RoomService, messageService service.MessageService) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		hub:            h,
		conn:           conn,
		roomService:    roomService,
		messageService: messageService,
		ctx:            ctx,
		cancel:         cancel,
		send:           make(chan OutgoingEvent, 32),
	}
}

// Start launches the client read and write loops in separate goroutines.
func (c *Client) Start() {
	go c.readLoop()
	go c.writeLoop()
}

// handleTypeMessage processes an incoming message event, persists it, and broadcasts it to the room.
func (c *Client) handleTypeMessage(in IncomingEvent) {
	if c.RoomID == 0 || c.Nick == "" {
		c.Send(OutgoingEvent{
			Type: EventTypeError,
			Text: model.ErrUnauthorized.Error(),
		})
		return
	}

	msg, err := c.messageService.Create(c.ctx, service.CreateMessageInput{
		RoomID: c.RoomID,
		Nick:   c.Nick,
		Text:   in.Text,
	})
	if err != nil {
		log.Println("ws: message service Create err:", err)
		c.Send(OutgoingEvent{
			Type: EventTypeError,
			Text: model.ErrBadRequest.Error(),
		})
		return
	}

	c.hub.Broadcast(OutgoingEvent{
		Type:    EventTypeMessage,
		RoomID:  c.RoomID,
		Nick:    c.Nick,
		Message: msg,
	})
}

// handleTypeHistory processes a history request event and sends recent messages back to the client.
func (c *Client) handleTypeHistory(in IncomingEvent) {
	if c.RoomID == 0 || c.Nick == "" {
		c.Send(OutgoingEvent{
			Type: EventTypeError,
			Text: model.ErrUnauthorized.Error(),
		})
		return
	}

	msgs, err := c.messageService.ListByRoom(c.ctx, c.RoomID, in.BeforeID, 50)
	if err != nil {
		log.Println("ws: message service ListByRoom err:", err)
		c.Send(OutgoingEvent{
			Type: EventTypeError,
			Text: model.ErrBadRequest.Error(),
		})
		return
	}
	c.Send(OutgoingEvent{
		Type:     EventTypeHistory,
		RoomID:   c.RoomID,
		Nick:     c.Nick,
		Messages: msgs,
	})
}

// handleTypeJoin processes a join event, validates the password, registers the client in the hub, and broadcasts the join.
func (c *Client) handleTypeJoin(in IncomingEvent) {
	ok, err := c.roomService.CheckPassword(c.ctx, in.RoomID, in.Password)
	if err != nil {
		log.Println("ws: room service CheckPassword err:", err)
	}

	if !ok {
		c.Send(OutgoingEvent{
			Type:   EventTypeError,
			RoomID: in.RoomID,
			Text:   model.ErrWrongPassword.Error(),
		})
		return
	}

	if err = c.roomService.TouchActivity(c.ctx, in.RoomID); err != nil {
		log.Println("ws: room service TouchActivity err:", err)
	}

	c.RoomID = in.RoomID
	c.Nick = in.Nick

	c.hub.register <- c

	c.hub.Broadcast(OutgoingEvent{
		Type:   EventTypeJoin,
		RoomID: in.RoomID,
		Nick:   in.Nick,
	})
}

// readLoop continuously reads incoming events from the WebSocket connection, validates, and dispatches them.
func (c *Client) readLoop() {
	defer func() {
		c.hub.unregister <- c
		c.Close()
	}()

	for {
		var in IncomingEvent

		if err := c.conn.ReadJSON(&in); err != nil {
			log.Println("ws: readjson err:", err)
			return
		}
		if err := in.Validate(); err != nil {
			c.Send(OutgoingEvent{
				Type:   "error",
				RoomID: in.RoomID,
				Text:   err.Error(),
			})
			continue
		}
		switch in.Type {
		case EventTypeMessage:
			c.handleTypeMessage(in)
		case EventTypeHistory:
			c.handleTypeHistory(in)
		case EventTypeJoin:
			c.handleTypeJoin(in)
		default:
			log.Println("ws: unknown event type:", in.Type)
		}
	}
}

// writeLoop continuously sends outgoing events from the send buffer to the WebSocket connection.
func (c *Client) writeLoop() {
	defer func() {
		c.Close()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case event, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.conn.WriteJSON(event); err != nil {
				log.Println("ws: writejson err:", err)
				return
			}
		}
	}
}

// Send enqueues an outgoing event into the client send buffer or closes the client if the buffer is full.
func (c *Client) Send(event OutgoingEvent) {
	select {
	case c.send <- event:
	default:
		log.Printf("ws: send buffer full for nick=%s room=%d, closing client", c.Nick, c.RoomID)
		c.Close()
	}
}

// Close shuts down the client once, cancelling its context, closing the send channel, and closing the WebSocket connection.
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.cancel()
		close(c.send)
		if err := c.conn.Close(); err != nil {
			log.Println("ws: close err:", err)
		}
	})
}
