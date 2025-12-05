package http

import (
	"net/http"
	"strconv"

	"github.com/Rasulikus/chat/internal/model"
	"github.com/Rasulikus/chat/internal/service"
	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	s service.RoomService
}

func NewRoomHandler(s service.RoomService) *RoomHandler {
	return &RoomHandler{
		s: s,
	}
}

// CreateRoomReq represents a request payload for creating a new room.
type CreateRoomReq struct {
	Name     string `json:"name" binding:"required,min=3,max=30"`
	Password string `json:"password" binding:"omitempty,max=30"`
}

// Create handles room creation.
//
// @Summary Create a new room
// @Description Creates a new chat room with an optional password.
// @Tags rooms
// @Accept json
// @Produce json
// @Param request body CreateRoomReq true "Room creation payload"
// @Success 201 {object} model.Room
// @Failure 400 {object} model.PublicError "invalid request or validation error"
// @Failure 500 {object} model.PublicError "internal server error"
// @Router /rooms [post]
func (h *RoomHandler) Create(c *gin.Context) {
	var req CreateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		if vErr, as := model.AsValidationError(req, err); as {
			status, pub := model.ToHTTP(vErr)
			c.AbortWithStatusJSON(status, pub)
			return
		} else {
			status, pub := model.ToHTTP(model.ErrBadRequest)
			c.AbortWithStatusJSON(status, pub)
			return
		}
	}

	ctx := c.Request.Context()

	room, err := h.s.Create(ctx, service.CreateRoomInput{
		Name:     req.Name,
		Password: req.Password,
	})
	if err != nil {
		status, pub := model.ToHTTP(err)
		c.AbortWithStatusJSON(status, pub)
	}
	c.JSON(http.StatusCreated, room)
}

type RoomListQuery struct {
	Limit    int    `form:"limit" binding:"omitempty,gte=1,lte=100"`
	BeforeID *int64 `form:"before_id"`
	Order    string `form:"order,default=created_at desc" binding:"omitempty,oneof='created_at desc' 'created_at asc' 'last_active_at desc' 'last_active_at asc' 'id desc' 'id asc''"`
}

// List returns a paginated list of rooms.
//
// @Summary List rooms
// @Description Returns a paginated list of rooms with optional ordering and cursor-based pagination.
// @Tags rooms
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of rooms to return (1-100)"
// @Param before_id query int false "Return rooms with IDs less than this value (cursor pagination)"
// @Param order query string false "Ordering key" Enums(created_at desc,created_at asc,last_activity_at desc,last_activity_at asc,id desc,id asc)
// @Success 200 {array} model.Room
// @Failure 400 {object} model.PublicError "invalid query parameters"
// @Failure 500 {object} model.PublicError "internal server error"
// @Router /rooms [get]
func (h *RoomHandler) List(c *gin.Context) {
	var q RoomListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		if vErr, as := model.AsValidationError(q, err); as {
			status, pub := model.ToHTTP(vErr)
			c.AbortWithStatusJSON(status, pub)
			return
		} else {
			status, pub := model.ToHTTP(model.ErrBadRequest)
			c.AbortWithStatusJSON(status, pub)
			return
		}
	}

	ctx := c.Request.Context()
	rooms, err := h.s.List(ctx, q.Limit, q.Order, q.BeforeID)
	if err != nil {
		status, pub := model.ToHTTP(err)
		c.AbortWithStatusJSON(status, pub)
		return
	}
	c.JSON(http.StatusOK, rooms)
}

// GetByID returns a room by its ID.
//
// @Summary Get room by ID
// @Description Returns a single room by its numeric identifier.
// @Tags rooms
// @Accept json
// @Produce json
// @Param id path int true "Room ID"
// @Success 200 {object} model.Room
// @Failure 400 {object} model.PublicError "invalid room ID"
// @Failure 404 {object} model.PublicError "room not found"
// @Failure 500 {object} model.PublicError "internal server error"
// @Router /rooms/{id} [get]
func (h *RoomHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		status, pub := model.ToHTTP(model.ErrBadRequest)
		c.AbortWithStatusJSON(status, pub)
		return
	}

	ctx := c.Request.Context()
	room, err := h.s.GetByID(ctx, id)
	if err != nil {
		status, pub := model.ToHTTP(err)
		c.AbortWithStatusJSON(status, pub)
		return
	}
	c.JSON(http.StatusOK, room)
}
