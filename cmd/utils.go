/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/posthog/posthog-go"
)

var AAITokenEnvName = "ASSMEBLYAI_TOKEN"
var AAIURL = "https://api.assemblyai.com/v2"
var PH_TOKEN string

func TelemetryCaptureEvent(event string, properties *PostHogProperties) {
	isTelemetryEnabled := getConfigFileValue("features.telemetry")
	if isTelemetryEnabled == "true" {

		if PH_TOKEN == "" {
			godotenv.Load()
			PH_TOKEN = os.Getenv("POSTHOG_API_TOKEN")
		}

		client := posthog.New(PH_TOKEN)
		defer client.Close()

		distinctId := getConfigFileValue("config.distinct_id")

		if distinctId == "" {
			distinctId = uuid.New().String()
			setConfigFileValue("config.distinct_id", distinctId)
		}
		if properties != nil {
			PhProperties := posthog.NewProperties().
				Set("poll", properties.Poll).
				Set("json", properties.Json).
				Set("speaker_labels", properties.SpeakerLabels).
				Set("punctuate", properties.Punctuate).
				Set("format_text", properties.FormatText).
				Set("dual_channel", properties.DualChannel).
				Set("redact_pii", properties.RedactPii).
				Set("auto_highlights", properties.AutoHighlights).
				Set("content_moderation", properties.ContentModeration).
				Set("topic_detection", properties.TopicDetection).
				Set("sentiment_analysis", properties.SentimentAnalysis).
				Set("auto_chapters", properties.AutoChapters).
				Set("entity_detection", properties.EntityDetection)

			client.Enqueue(posthog.Capture{
				DistinctId: distinctId,
				Event:      event,
				Properties: PhProperties,
			})
			return
		}

		client.Enqueue(posthog.Capture{
			DistinctId: distinctId,
			Event:      event,
		})
	}
}

func CallSpinner(message string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[7], 100*time.Millisecond, spinner.WithSuffix(message))
	s.Start()
	return s
}

func PrintError(err error) {
	if err != nil {
		fmt.Println(err)
		return
	}
}

func QueryApi(token string, path string, method string, body io.Reader) []byte {
	resp, err := http.NewRequest(method, AAIURL+path, body)
	PrintError(err)

	resp.Header.Add("Accept", "application/json")
	resp.Header.Add("Authorization", token)
	resp.Header.Add("Transfer-Encoding", "chunked")

	response, err := http.DefaultClient.Do(resp)
	PrintError(err)
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	PrintError(err)
	return responseData
}

func BeutifyJSON(data []byte) []byte {
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, data, "", "\t")
	if error != nil {
		return data
	}
	return prettyJSON.Bytes()
}

type PostHogProperties struct {
	Poll              bool  `json:"poll"`
	Json              bool  `json:"json"`
	SpeakerLabels     bool  `json:"speaker_labels"`
	Punctuate         bool  `json:"punctuate"`
	FormatText        bool  `json:"format_text"`
	DualChannel       *bool `json:"dual_channel"`
	RedactPii         bool  `json:"redact_pii"`
	AutoHighlights    bool  `json:"auto_highlights"`
	ContentModeration bool  `json:"content_safety"`
	TopicDetection    bool  `json:"iab_categories"`
	SentimentAnalysis bool  `json:"sentiment_analysis"`
	AutoChapters      bool  `json:"auto_chapters"`
	EntityDetection   bool  `json:"entity_detection"`
}
