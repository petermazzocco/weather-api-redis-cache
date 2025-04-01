package main

import (
	"weather-redis-cache/initializers"
	"weather-redis-cache/utils"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.InitENV()
	initializers.InitRedis()
}

func main() {
	r := gin.Default()

	// Serve static files from static directory
	r.Static("/static", "./static")

	// Load HTML templates from the "templates" directory
	r.LoadHTMLGlob("templates/*")

	// Route for the homepage
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"title": "Weather App",
		})
	})

	r.POST("/weather", utils.FetchWeather)

	r.Run(":8080")
}
