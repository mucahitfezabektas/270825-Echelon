package open_trip

import (
	"context"
	"mini_CMS_Desktop_App/services"

	"github.com/gofiber/fiber/v2"
)

type OpenTripHandler struct {
	Service *services.OpenTripService
}

func NewOpenTripHandler(service *services.OpenTripService) *OpenTripHandler {
	return &OpenTripHandler{Service: service}
}

func (h *OpenTripHandler) GetOpenTrips(c *fiber.Ctx) error {
	period := c.Query("period")
	if period == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "period is required"})
	}

	results, err := h.Service.GetOpenTrips(context.Background(), period)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(results)
}
