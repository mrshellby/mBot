package slack

import (
	"fmt"

	"github.com/slack-go/slack"
	"github.com/un4gi/mBot/config"
)

func Hook(target string, title string) {
	attachment := slack.Attachment{
		Color:         "good",
		Fallback:      "You successfully claimed a mission!",
		Text:          "Target: "+target+"\nTitle: "+title+"\n",
		Ts:            json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
	}, target, amount
	msg := slack.WebhookMessage{
		Attachments: []slack.Attachment{attachment},
	}

	err := slack.PostWebhook(config.Slack_Webhook, &msg)
	if err != nil {
		fmt.Println(err)
	}
}