package messaging

import (
	"fmt"
	"github.com/lytics/slackhook"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"os"
	"time"
)

type slackConfig struct {
	webhook      string
	user         string
	RollupWindow time.Duration
	Client       *slackhook.Client
}

func NewSlackConfig(c *viper.Viper) (*slackConfig, error) {
	api := &slackConfig{}
	api.webhook = c.GetString("Slack.Url")
	api.user = c.GetString("Slack.Username")
	api.RollupWindow = c.GetDuration("Slack.MessageWindow")
	api.Client = slackhook.New(api.webhook)
	return api, nil
}

type slack interface {
	SendMessage(msg, msgType string) error
}

func (api *slackConfig) NewMessage(msg string) *slackhook.Message {
	m := new(slackhook.Message)
	m.Text = msg
	m.UserName = api.user
	m.IconEmoji = ":white_check_mark:"
	return m
}

// type webHookMessage slackhook.Message
// type Client *slackhook.Client

func newClient(webhook string) *slackhook.Client {
	c := slackhook.New(webhook)
	return c
}

func (api *slackConfig) SendMessage(msg, slType string, ts int) error {
	m := api.NewMessage(msg)

	m.Text = msg
	var col string
	switch slType {
	case "error":
		m.Text = "error"
		m.IconEmoji = ":sos:"
		col = "#f4241d"
	case "warn":
		m.Text = "warn"
		m.IconEmoji = ":question:"
		col = "#f49e1d"
	case "info":
		m.Text = "info"
		m.IconEmoji = ":information_source:"
		col = "#201df4"
	default:
		m.Text = "message"
		m.IconEmoji = ":white_check_mark:"
		col = "#41f41d"
	}

	var authur string
	authur, err := os.Hostname()
	if err != nil {
		authur = "Hashmonitor"
	}

	var a = slackhook.Attachment{
		Fallback:   msg,
		Color:      col,
		AuthorName: authur,
		AuthorLink: "",
		AuthorIcon: "",
		Title:      fmt.Sprintf("%v", slType),
		TitleLink:  "",
		Text:       msg,
		Fields:     nil,
		ImageURL:   "",
		ThumbURL:   "",
		FooterIcon: "",
		Footer:     "",
		Timestamp:  ts,
	}

	m.AddAttachment(&a)
	err = api.Client.Send(m)
	return errors.Wrap(err, "failed sending slack message")
}
