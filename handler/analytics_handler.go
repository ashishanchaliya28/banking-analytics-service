package handler

import (
	"github.com/banking-superapp/analytics-service/model"
	"github.com/banking-superapp/analytics-service/service"
	"github.com/gofiber/fiber/v2"
)

type AnalyticsHandler struct{ svc service.AnalyticsService }

func NewAnalyticsHandler(svc service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

func (h *AnalyticsHandler) RecordEvent(c *fiber.Ctx) error {
	userID := c.Get("X-User-ID")
	var req model.RecordEventRequest
	if err := c.BodyParser(&req); err != nil {
		return respond(c, fiber.StatusBadRequest, nil, "invalid request body")
	}
	if req.EventName == "" {
		return respond(c, fiber.StatusBadRequest, nil, "event_name is required")
	}
	event, err := h.svc.RecordEvent(c.Context(), userID, &req)
	if err != nil {
		return respond(c, fiber.StatusInternalServerError, nil, err.Error())
	}
	return respond(c, fiber.StatusCreated, event, "")
}

func (h *AnalyticsHandler) GetSegment(c *fiber.Ctx) error {
	userID := c.Get("X-User-ID")
	segment, err := h.svc.GetSegment(c.Context(), userID)
	if err != nil {
		return respond(c, fiber.StatusInternalServerError, nil, err.Error())
	}
	return respond(c, fiber.StatusOK, segment, "")
}

func (h *AnalyticsHandler) GetCrossSellOffers(c *fiber.Ctx) error {
	userID := c.Get("X-User-ID")
	offers, err := h.svc.GetCrossSellOffers(c.Context(), userID)
	if err != nil {
		return respond(c, fiber.StatusInternalServerError, nil, err.Error())
	}
	return respond(c, fiber.StatusOK, fiber.Map{"offers": offers, "count": len(offers)}, "")
}

func respond(c *fiber.Ctx, status int, data interface{}, errMsg string) error {
	if errMsg != "" {
		return c.Status(status).JSON(fiber.Map{"success": false, "error": errMsg})
	}
	return c.Status(status).JSON(fiber.Map{"success": true, "data": data})
}
