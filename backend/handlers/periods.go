// handlers/actual_periods.go
package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetPeriodRange(c *fiber.Ctx) error {
	center := c.Query("center") // örnek: 7-25

	var baseMonth, baseYear int

	if center != "" {
		parts := strings.Split(center, "-")
		if len(parts) == 2 {
			baseMonth, _ = strconv.Atoi(parts[0])
			yy, _ := strconv.Atoi(parts[1])
			baseYear = 2000 + yy
		}
	}

	// Eğer geçerli değilse bugünkü ay
	if baseMonth < 1 || baseMonth > 12 || baseYear < 2000 {
		now := time.Now()
		baseMonth = int(now.Month())
		baseYear = now.Year()
	}

	// Listeyi oluştur
	var periods []string
	for i := -6; i <= 6; i++ {
		t := time.Date(baseYear, time.Month(baseMonth)+time.Month(i), 1, 0, 0, 0, 0, time.UTC)
		period := fmt.Sprintf("%04d-%02d", t.Year(), int(t.Month()))

		periods = append(periods, period)
	}

	return c.JSON(periods)
}
