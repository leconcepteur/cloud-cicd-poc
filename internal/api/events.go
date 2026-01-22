package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/leconcepteur/cloud-cicd-poc/internal/matchmaking"
	"github.com/leconcepteur/cloud-cicd-poc/internal/middleware"
)

type EventsHandler struct {
	eventHub *matchmaking.EventHub
}

func NewEventsHandler(eventHub *matchmaking.EventHub) *EventsHandler {
	return &EventsHandler{
		eventHub: eventHub,
	}
}

func (h *EventsHandler) Stream(c echo.Context) error {
	userID := middleware.GetUserID(c)

	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no")

	// Subscribe to events
	eventChan := h.eventHub.Subscribe(userID)
	defer h.eventHub.Unsubscribe(userID)

	// Send initial connection event
	fmt.Fprintf(c.Response(), "event: connected\ndata: {\"status\":\"connected\"}\n\n")
	c.Response().Flush()

	ctx := c.Request().Context()

	for {
		select {
		case <-ctx.Done():
			return nil
		case data, ok := <-eventChan:
			if !ok {
				return nil
			}
			fmt.Fprintf(c.Response(), "data: %s\n\n", data)
			c.Response().Flush()
		}
	}
}

func (h *EventsHandler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
