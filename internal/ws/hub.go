package ws

import "sync"

type Broadcast struct {
	RoomID int64
	Event  OutgoingEvent
}

type Hub struct {
	mu    sync.RWMutex
	rooms map[int64]*RoomRuntime

	register   chan *Client
	unregister chan *Client
	broadcast  chan Broadcast
}

type RoomRuntime struct {
	ID      int64
	clients map[*Client]struct{}
}

// NewHub creates a new Hub instance with initialized room map and internal channels.
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[int64]*RoomRuntime),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Broadcast),
	}
}

// Run starts the hub event loop and processes client registration, unregistration, and broadcast requests.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.addClient(c)
		case c := <-h.unregister:
			h.removeClient(c)
		case b := <-h.broadcast:
			h.broadcastToRoom(b)
		}
	}
}

// addClient registers a client in the corresponding room runtime, creating the room if it does not exist.
func (h *Hub) addClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.rooms[c.RoomID]
	if !ok {
		room = &RoomRuntime{
			ID:      c.RoomID,
			clients: make(map[*Client]struct{}),
		}
		h.rooms[c.RoomID] = room
	}
	room.clients[c] = struct{}{}
}

// removeClient unregisters a client from its room and removes the room if it becomes empty.
func (h *Hub) removeClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.rooms[c.RoomID]
	if !ok {
		return
	}
	delete(room.clients, c)
	if len(room.clients) == 0 {
		delete(h.rooms, c.RoomID)
	}
}

// broadcastToRoom sends an event to all clients currently connected to the given room.
func (h *Hub) broadcastToRoom(b Broadcast) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, ok := h.rooms[b.RoomID]
	if !ok {
		return
	}

	for c := range room.clients {
		c.Send(b.Event)
	}
}

// Broadcast enqueues an outgoing event to be dispatched to all clients in the specified room.
func (h *Hub) Broadcast(event OutgoingEvent) {
	if event.RoomID == 0 {
		return
	}
	h.broadcast <- Broadcast{
		RoomID: event.RoomID,
		Event:  event,
	}
}
