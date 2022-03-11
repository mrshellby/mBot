package slack

import (
		"encoding/json"
	"fmt"
	"strconv"
	"time"
	
	"github.com/slack-go/slack"
	"github.com/un4gi/mBot/config"
)

func Hook(m string) {
	attachment := slack.Attachment{
		Color:         "good",
		Fallback:      "You successfully claimed a mission!",
		Text:          m,
		Ts:            json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
	}
	msg := slack.WebhookMessage{
		Attachments: []slack.Attachment{attachment},
	}

	err := slack.PostWebhook(config.Slack_Webhook, &msg)
	if err != nil {
		fmt.Println(err)
	}
}