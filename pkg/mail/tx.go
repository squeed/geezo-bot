package mail

import (
	"io"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/pkg/errors"

	"github.com/squeed/geezo-bot/pkg/config"
)

func SendMessage(c *config.Config, msg io.Reader) error {
	// Set up authentication information.
	auth := sasl.NewPlainClient("", c.Smtp.Username, c.Smtp.Password)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	to := []string{c.Main.To}
	err := smtp.SendMail(c.Smtp.Server, auth, c.Main.From, to, msg)
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}
	return nil

}
