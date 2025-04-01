package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"weather-redis-cache/initializers"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Location struct
type Location struct {
	Location string `json:"location" binding:"required"`
}

func FetchWeather(c *gin.Context) {
	// Validate body request
	var location Location
	if err := c.BindJSON(&location); err != nil {
		fmt.Println("JSON binding error:", err.Error())
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Log the location to help with debugging
	fmt.Println("Received location:", location.Location)

	// Validate that location is not empty
	if location.Location == "" {
		c.JSON(400, gin.H{"error": "Location is required"})
		return
	}

	// Create a unique cache key for this specific location
	cacheKey := fmt.Sprintf("weather:%s", location.Location)

	// First, attempt to retrieve data from the Redis cache
	cachedData, err := initializers.RDB.Get(initializers.CTX, cacheKey).Result()

	// Get the current timestamp for response metadata
	currentTime := time.Now()
	formattedTime := currentTime.Format(time.RFC3339)

	// Cache hit case - we found data in the cache!
	if err == nil {
		fmt.Println("✅ Cache HIT for", location.Location)

		// Return the cached data along with metadata indicating it came from cache
		c.JSON(http.StatusOK, gin.H{
			"weather":   cachedData,
			"source":    "cache",
			"cached_at": formattedTime,
		})
		return
	}

	// Handle potential Redis errors that aren't just "key not found"
	if err != redis.Nil {
		fmt.Println("⚠️ Redis error:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis error: " + err.Error()})
		return
	}

	// Cache miss - need to fetch fresh data from the API
	fmt.Println("❌ Cache MISS for", location.Location, "- fetching from API")

	// Construct the API URL with the location and your API key
	apiKey := os.Getenv("KEY")
	url := "https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/" +
		location.Location + "?unitGroup=us&key=" + apiKey + "&contentType=json"

	fmt.Println("Making API request to:", url)

	// Make the HTTP request to the weather API
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("⚠️ API request failed:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch weather data: " + err.Error()})
		return
	}

	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("⚠️ Failed to read API response:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body: " + err.Error()})
		return
	}

	// Store the API response in Redis cache with a 1 minute TTL
	err = initializers.RDB.Set(initializers.CTX, cacheKey, string(body), 1*time.Minute).Err()
	if err != nil {
		fmt.Println("⚠️ Error caching data:", err.Error())
		// We'll continue even if caching fails - just log the error
	} else {
		fmt.Println("✅ Successfully cached data for", location.Location)
	}

	// Return the fresh API data to the client with metadata
	c.JSON(http.StatusOK, gin.H{
		"weather":       string(body),
		"source":        "api",
		"fetched_at":    formattedTime,
		"cache_expires": currentTime.Add(1 * time.Minute).Format(time.RFC3339),
	})
}
