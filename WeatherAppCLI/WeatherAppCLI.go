// API which stands for Application Programming Interface is a set of rules that allows different software applications to communicate with each other.
// In this code, the OpenWeatherMap API is used to fetch weather data for a given city and display it in the command line interface (CLI).
import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// WeatherData struct to hold the weather information
type WeatherData struct {
	Name string `json:"name"`
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run WeatherApp.go [city]")
		return
	}

	city := os.Args[1]
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENWEATHER_API_KEY environment variable is not set.")
		return
	}
	baseURL := "https://api.openweathermap.org/data/2.5/weather"
	urlPath := fmt.Sprintf("%s?q=%s&appid=%s&units=metric", baseURL, url.QueryEscape(city), apiKey)

	resp, err := http.Get(urlPath)
	if err != nil {
		fmt.Println("Error fetching weather data:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: received non-200 response code (%d). Check your API Key or City spelling.\n", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	var weatherData WeatherData
	err = json.Unmarshal(body, &weatherData)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	fmt.Printf("Weather in %s:\n", weatherData.Name)
	fmt.Printf("Temperature: %.2f°C\n", weatherData.Main.Temp)
	if len(weatherData.Weather) > 0 {
		fmt.Printf("Description: %s\n", weatherData.Weather[0].Description)
	}
}

