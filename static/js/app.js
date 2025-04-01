document.addEventListener("DOMContentLoaded", function () {
  // Get form elements
  const weatherForm = document.getElementById("weatherForm");
  const locationInput = document.getElementById("location");

  // Get results display elements
  const weatherResults = document.getElementById("weatherResults");
  const locationName = document.getElementById("locationName"); // This is the span to display the location
  const dataSource = document.getElementById("dataSource");
  const timeInfo = document.getElementById("timeInfo");
  const weatherData = document.getElementById("weatherData");
  const loadingIndicator = document.getElementById("loadingIndicator");
  const errorMessage = document.getElementById("errorMessage");

  weatherForm.addEventListener("submit", function (e) {
    e.preventDefault();

    // Get location from the single input field
    const location = locationInput.value.trim();

    // Validate input
    if (!location) {
      alert("Please enter a location");
      return;
    }

    // Show loading, hide results and errors
    loadingIndicator.classList.remove("hidden");
    weatherResults.classList.add("hidden");
    errorMessage.classList.add("hidden");

    // Make the API request
    fetchWeather(location);
  });

  function fetchWeather(location) {
    // Log what we're sending to help with debugging
    console.log("Sending location:", location);
    console.log("Request body:", JSON.stringify({ location: location }));

    fetch("/weather", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ location: location }),
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error("Network response was not ok");
        }
        return response.json();
      })
      .then((data) => {
        console.log("Received data:", data);

        // Hide loading indicator
        loadingIndicator.classList.add("hidden");

        // Display the location in the results span
        locationName.textContent = location;

        // Show where the data came from (cache or API)
        dataSource.textContent = data.source || "API"; // Default to API if not specified
        dataSource.className = data.source || "api"; // Add class for styling

        // Show timing information
        if (data.source === "cache") {
          timeInfo.textContent = `Retrieved from cache at: ${data.cached_at}`;
        } else if (data.fetched_at) {
          timeInfo.textContent = `Fetched from API at: ${data.fetched_at}. Cache expires: ${data.cache_expires}`;
        } else {
          timeInfo.textContent = ""; // Clear if no timing info
        }

        // Parse the weather data
        let weatherJson;

        try {
          // The weather data might be a string that needs parsing
          if (typeof data.weather === "string") {
            weatherJson = JSON.parse(data.weather);
          } else {
            weatherJson = data.weather;
          }

          // Display the weather information
          displayWeatherData(weatherJson);

          // Show the results
          weatherResults.classList.remove("hidden");
        } catch (error) {
          console.error("Error parsing weather data:", error);
          console.error("Raw weather data:", data.weather);
          errorMessage.classList.remove("hidden");
        }
      })
      .catch((error) => {
        console.error("Error fetching weather:", error);
        loadingIndicator.classList.add("hidden");
        errorMessage.classList.remove("hidden");
      });
  }

  function displayWeatherData(data) {
    // Clear the previous weather data
    weatherData.innerHTML = "";

    // Create a weather summary
    let html = `
            <div class="weather-summary">
                <h3>Current Conditions</h3>
                <p><strong>Temperature:</strong> ${
                  data.currentConditions?.temp || "N/A"
                } °F</p>
                <p><strong>Conditions:</strong> ${
                  data.currentConditions?.conditions || "N/A"
                }</p>
                <p><strong>Humidity:</strong> ${
                  data.currentConditions?.humidity || "N/A"
                }%</p>
                <p><strong>Wind:</strong> ${
                  data.currentConditions?.windspeed || "N/A"
                } mph</p>
            </div>
            
            <div class="forecast">
                <h3>Next Few Days</h3>
                <div class="forecast-days">
        `;

    // Add forecast for the next few days if available
    if (data.days && data.days.length > 0) {
      // Take the first 5 days only
      const forecastDays = data.days.slice(0, 5);

      forecastDays.forEach((day) => {
        const date = new Date(day.datetime);
        const formattedDate = date.toLocaleDateString("en-US", {
          weekday: "short",
          month: "short",
          day: "numeric",
        });

        html += `
                    <div class="forecast-day">
                        <h4>${formattedDate}</h4>
                        <p>High: ${day.tempmax} °F</p>
                        <p>Low: ${day.tempmin} °F</p>
                        <p>${day.conditions}</p>
                    </div>
                `;
      });
    } else {
      html += "<p>No forecast data available</p>";
    }

    html += `
                </div>
            </div>
            
            <div class="location-info">
                <h3>Location Information</h3>
                <p><strong>Address:</strong> ${
                  data.resolvedAddress || "N/A"
                }</p>
                <p><strong>Timezone:</strong> ${data.timezone || "N/A"}</p>
                <p><strong>Latitude:</strong> ${data.latitude || "N/A"}</p>
                <p><strong>Longitude:</strong> ${data.longitude || "N/A"}</p>
            </div>
        `;

    // Add the HTML to the weather data div
    weatherData.innerHTML = html;
  }
});
