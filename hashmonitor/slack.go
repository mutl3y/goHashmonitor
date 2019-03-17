package hashmonitor

import (
	"fmt"
	"github.com/lytics/slackhook"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"time"
)

type Client struct {
	slackhook.Client
}
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

func (api *slackConfig) NewMessage(msg string) error {
	m := new(slackhook.Message)
	m.Text = msg
	m.UserName = api.user
	m.IconEmoji = ":sos:"
	// m.Attachments = append([]sl.Attachment, new(sl.Attachment))
	m.Channel = "hashmonitor_dev"
	return nil
}

// type webHookMessage slackhook.Message
// type Client *slackhook.Client

func NewClient() *slackhook.Client {
	c := slackhook.New(cfg.GetString("Slack.Url"))
	return c
}

func (s *slackConfig) SendMessage(msg, slType string) error {
	m := new(slackhook.Message)
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
	fmt.Println(m)
	// var text = "var slackUsername"

	var a = slackhook.Attachment{
		Fallback:   msg,
		Color:      col,
		AuthorName: "",
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
		Timestamp:  0,
	}

	m.AddAttachment(&a)
	fmt.Println(m)
	err := s.Client.Send(m)
	return errors.Wrap(err, "failed sending slack message")
}
