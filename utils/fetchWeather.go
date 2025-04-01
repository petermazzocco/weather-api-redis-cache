package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
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
	location := c.PostForm("location")

	// Validate that location is not empty
	if location == "" {
		if c.GetHeader("HX-Request") == "true" {
			c.HTML(http.StatusBadRequest, "weather_results.html", gin.H{
				"error": "Location is required",
			})
		} else {
			c.HTML(http.StatusBadRequest, "index.html", gin.H{
				"error": "Location is required",
			})
		}
		return
	}

	// Then use the location string directly
	cacheKey := fmt.Sprintf("weather:%s", location)

	// First, attempt to retrieve data from the Redis cache
	cachedData, err := initializers.RDB.Get(initializers.CTX, cacheKey).Result()

	// Get the current timestamp for response metadata
	currentTime := time.Now()
	formattedTime := currentTime.Format(time.RFC3339)

	// Cache hit case - we found data in the cache!
	if err == nil {
		fmt.Println("✅ Cache HIT for", location)

		// Parse the weather data
		var weatherData map[string]any
		err = json.Unmarshal([]byte(cachedData), &weatherData)
		if err != nil {
			if c.GetHeader("HX-Request") == "true" {
				c.HTML(http.StatusInternalServerError, "weather_results.html", gin.H{
					"error": "Failed to parse weather data",
				})
			} else {
				c.HTML(http.StatusInternalServerError, "index.html", gin.H{
					"error": "Failed to parse weather data",
				})
			}
			return
		}

		// Return appropriate template based on request type
		templateData := gin.H{
			"weather":   weatherData,
			"source":    "cache",
			"timestamp": formattedTime,
			"location":  location,
		}

		if c.GetHeader("HX-Request") == "true" {
			c.HTML(http.StatusOK, "weather_results.html", templateData)
		} else {
			c.HTML(http.StatusOK, "index.html", templateData)
		}
		return
	}

	// Handle potential Redis errors that aren't just "key not found"
	if err != redis.Nil {
		fmt.Println("⚠️ Redis error:", err.Error())
		errorMessage := "Redis error: " + err.Error()

		if c.GetHeader("HX-Request") == "true" {
			c.HTML(http.StatusInternalServerError, "weather_results.html", gin.H{
				"error": errorMessage,
			})
		} else {
			c.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"error": errorMessage,
			})
		}
		return
	}

	// Cache miss - need to fetch fresh data from the API
	fmt.Println("❌ Cache MISS for", location, "- fetching from API")

	// Clean and encode the location
	cleanedLocation := strings.ToLower(location)
	cleanedLocation = strings.ReplaceAll(cleanedLocation, ",", "")
	encodedLocation := url.QueryEscape(cleanedLocation)

	// Construct the API URL with the location and your API key
	apiKey := os.Getenv("KEY")
	url := "https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/" +
		encodedLocation + "?unitGroup=us&key=" + apiKey + "&contentType=json"

	fmt.Println("Making API request to:", url)

	// Make the HTTP request to the weather API
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("⚠️ API request failed:", err.Error())
		errorMessage := "Failed to fetch weather data: " + err.Error()

		if c.GetHeader("HX-Request") == "true" {
			c.HTML(http.StatusInternalServerError, "weather_results.html", gin.H{
				"error": errorMessage,
			})
		} else {
			c.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"error": errorMessage,
			})
		}
		return
	}

	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("⚠️ Failed to read API response:", err.Error())
		errorMessage := "Failed to read response body: " + err.Error()

		if c.GetHeader("HX-Request") == "true" {
			c.HTML(http.StatusInternalServerError, "weather_results.html", gin.H{
				"error": errorMessage,
			})
		} else {
			c.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"error": errorMessage,
			})
		}
		return
	}

	// Store the API response in Redis cache with a 1 minute TTL
	err = initializers.RDB.Set(initializers.CTX, cacheKey, string(body), 1*time.Minute).Err()
	if err != nil {
		fmt.Println("⚠️ Error caching data:", err.Error())
		// We'll continue even if caching fails - just log the error
	} else {
		fmt.Println("✅ Successfully cached data for", location)
	}

	// Parse the weather data - FIX: Use body instead of cachedData
	var weatherData map[string]any
	err = json.Unmarshal(body, &weatherData)
	if err != nil {
		errorMessage := "Failed to parse weather data"

		if c.GetHeader("HX-Request") == "true" {
			c.HTML(http.StatusInternalServerError, "weather_results.html", gin.H{
				"error": errorMessage,
			})
		} else {
			c.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"error": errorMessage,
			})
		}
		return
	}

	// Return based on response
	templateData := gin.H{
		"weather":   weatherData,
		"source":    "api",
		"timestamp": formattedTime,
		"location":  location,
	}

	if c.GetHeader("HX-Request") == "true" {
		c.HTML(http.StatusOK, "weather_results.html", templateData)
	} else {
		c.HTML(http.StatusOK, "index.html", templateData)
	}
}
