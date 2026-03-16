package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

const URL string = "https://api.openweathermap.org/data/2.5/weather?"

type WeatherData struct {
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Name string `json:"name"`
}

func getCurrentWeather(message string) *discordgo.MessageSend {
	// extract 5-digit US zip code from message
	r, _ := regexp.Compile(`\d{5}`)
	zip := r.FindString(message)

	// if ZIP not found, return an error
	if zip == "" {
		return &discordgo.MessageSend{
			Content: "Please provide a valid 5-digit US zip code",
		}
	}

	weatherURL := fmt.Sprintf("%szip=%s&units=imperial&appid=%s", URL, zip, OpenWeatherToken)

	// create new http client with timeout
	client := http.Client{Timeout: 5 * time.Second}

	// query OpenWeather API
	response, err := client.Get(weatherURL)
	if err != nil {
		return &discordgo.MessageSend{
			Content: "Sorry, there was an error trying to get the weather",
		}
	}

	// open http response body
	body, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	// convert JSON
	var data WeatherData
	json.Unmarshal([]byte(body), &data)

	// pull out desired weather info and convert to string if necessary
	city := data.Name
	conditions := data.Weather[0].Description
	temp := strconv.FormatFloat(data.Main.Temp, 'f', 2, 64)
	humidity := strconv.Itoa(data.Main.Humidity)
	windSpeed := strconv.FormatFloat(data.Wind.Speed, 'f', 2, 64)

	// build discord embed response
	embed := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{{
			Type: discordgo.EmbedTypeRich,
			Title: "Current Weather",
			Description: "for " + city,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name: "Conditions",
					Value: conditions,
					Inline: true,
				},
				{
					Name: "Temperature",
					Value: temp + " °F",
					Inline: true,
				},
				{
					Name: "Humidity",
					Value: humidity + " %",
					Inline: true,
				},
				{
					Name: "Wind Speed",
					Value: windSpeed + " mph",
					Inline: true,
				},
			},
		},
	},
	}

	return embed
}