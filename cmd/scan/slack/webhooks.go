package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Sharsie/reality-scanner/cmd/scan/config"
	"github.com/Sharsie/reality-scanner/cmd/scan/logger"
	"github.com/Sharsie/reality-scanner/cmd/scan/reality"
)

type slackTextDefintion struct {
	Text     string `json:"text"`
	Type     string `json:"type"`
	Verbatim bool   `json:"verbatim,omitempty"`
}

type slackElement struct {
	Action   string              `json:"action_id,omitempty"`
	Text     *slackTextDefintion `json:"text,omitempty"`
	AltText  string              `json:"alt_text,omitempty"`
	Type     string              `json:"type,omitempty"`
	Url      string              `json:"url,omitempty"`
	ImageUrl string              `json:"image_url,omitempty"`
	Style    string              `json:"style,omitempty"`
}

type slackAccessory struct {
	Type     string              `json:"type"`
	Text     *slackTextDefintion `json:"text,omitempty"`
	Value    string              `json:"value,omitempty"`
	AltText  string              `json:"alt_text,omitempty"`
	Url      string              `json:"url,omitempty"`
	ImageUrl string              `json:"image_url,omitempty"`
	ActionId string              `json:"action_id,omitempty"`
}
type slackBlock struct {
	Type      string              `json:"type"`
	Text      *slackTextDefintion `json:"text,omitempty"`
	AltText   string              `json:"alt_text,omitempty"`
	ImageUrl  string              `json:"image_url,omitempty"`
	Elements  []slackElement      `json:"elements,omitempty"`
	Accessory *slackAccessory     `json:"accessory,omitempty"`
}

type requestData struct {
	Text   string       `json:"text,omitempty"`
	Blocks []slackBlock `json:"blocks"`
}

func Send(l *logger.Log, message string) error {
	l.Debug("Sending slack message %s", message)

	payload := requestData{
		message,
		nil,
	}

	requestBody, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	client := http.Client{Timeout: 5 * time.Second}
	_, err = client.Post(config.SlackWebhookUrl, "application/json", bytes.NewBuffer(requestBody))

	return err
}

func SendNewReality(l *logger.Log, item reality.KnownReality) error {
	l.Debug("Sending slack message for reality id %s", item.Id)

	payload := requestData{
		"",
		[]slackBlock{},
	}

	appendText(item, &payload)
	appendRealityType(item, &payload)
	appendImages(item, &payload)

	payload.Blocks = append(payload.Blocks, slackBlock{
		Type: "divider",
	})

	requestBody, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	client := http.Client{Timeout: 5 * time.Second}
	_, err = client.Post(config.SlackWebhookUrl, "application/json", bytes.NewBuffer(requestBody))

	l.Debug("%s", requestBody)

	return err
}

func appendText(item reality.KnownReality, payload *requestData) {
	payload.Blocks = append(payload.Blocks, slackBlock{
		Type: "section",
		Text: &slackTextDefintion{
			Type:     "mrkdwn",
			Text:     fmt.Sprintf("Nová věc <@U01C5RDRAKZ> <@U01CH3DQ9EH>"),
			Verbatim: false,
		},
	})
}

func appendRealityType(item reality.KnownReality, payload *requestData) {
	payload.Blocks = append(payload.Blocks, slackBlock{
		Type: "section",
		Text: &slackTextDefintion{
			Type: "mrkdwn",
			Text: fmt.Sprintf("%s, %s Cena: %d,-", item.Title, item.Place, item.Price),
		},
		Accessory: &slackAccessory{
			Type: "button",
			Text: &slackTextDefintion{
				Type: "plain_text",
				Text: "Otevřít",
			},
			Url:      item.Link,
			ActionId: "reality=opener",
		},
	})
}

func appendImages(item reality.KnownReality, payload *requestData) {
	for _, image := range item.Images {
		payload.Blocks = append(payload.Blocks, slackBlock{
			Type:     "image",
			AltText:  "preview",
			ImageUrl: image,
		})
	}
}
