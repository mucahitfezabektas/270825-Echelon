package handlers

import (
	"context"
	"log"

	"mini_CMS_Desktop_App/models"
	"mini_CMS_Desktop_App/repositories"

	"github.com/gofiber/fiber/v2"
)

// PublishQueryHandler, 'publishes' verilerini sorgulama mantığını içerir.
type PublishQueryHandler struct {
	publishRepo *repositories.PublishRepository
}

// NewPublishQueryHandler, handler'ın yeni bir örneğini oluşturur.
func NewPublishQueryHandler(publishRepo *repositories.PublishRepository) *PublishQueryHandler {
	return &PublishQueryHandler{publishRepo: publishRepo}
}

// GetPublishesByPersonID retrieves all 'publishes' data for a given 'person_id'.
// Required parameter: 'person_id' (as a query parameter).
// This method brings all publish data belonging to the staff ID.
func (h *PublishQueryHandler) GetPublishesByPersonID(c *fiber.Ctx) error {
	personID := c.Query("person_id")

	if personID == "" {
		log.Println("❌ Missing 'person_id' parameter.")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "'person_id' parameter is required."})
	}

	log.Printf("🔍 Querying all Publish data for Person ID: %s", personID)

	var publishes []models.Publish
	var err error

	// Call the repository method to get all publishes by person_id.
	publishes, err = h.publishRepo.GetPublishesByPersonID(context.Background(), personID)
	if err != nil {
		log.Printf("❌ Failed to retrieve Publish data for Person ID: %s: %v", personID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve Publish data", "details": err.Error()})
	}

	if len(publishes) == 0 {
		log.Printf("ℹ️  No Publish record found for Person ID: %s.", personID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "No Publish record found for the specified person."})
	}

	log.Printf("✅ Found %d Publish records for Person ID: %s.", len(publishes), personID)
	return c.JSON(publishes)
}
