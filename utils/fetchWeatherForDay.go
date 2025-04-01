package utils

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func FetchWeatherByDay(c *gin.Context) {
	// Fetch location and date parameters
	date := c.Param("day")
	
	// In a real implementation, you would fetch this from cache or API
	// For now, use placeholder data with the proper structure
	
	c.HTML(200, "day_details.html", gin.H{
		"title":      "Weather Details",
		"location":   "New York", // This should be retrieved from session or query
		"datetime":   date,
		"conditions": "Partly Cloudy",
		"tempmax":    75,
		"tempmin":    52,
		"currentConditions": map[string]interface{}{
			"temp":       70,
			"conditions": "Partly Cloudy",
			"humidity":   45,
		},
	})
	fmt.Println("Currently on the routed page for", date)
}
