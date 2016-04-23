package rwnewsengine

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mailgun/mailgun-go"
)

var SendEmail = func(c *Config, e *OutboundEmail) error {
	mg := mailgun.NewMailgun(c.MailgunDomain, c.MailgunPrivateKey, c.MailgunPublicKey)
	m := mg.NewMessage(e.Sender, e.Subject, e.Body, e.Recipient)
	if e.InReplyTo != "" {
		m.AddHeader("In-Reply-To", e.InReplyTo)
	}
	if e.BodyHTML != "" {
		m.SetHtml(e.BodyHTML)
	}
	response, id, err := mg.Send(m)
	if err != nil {
		log.Error("Error sending mail", err)
		return err
	}
	log.WithFields(log.Fields{"at": "mail-sent", "reponse-id": id, "message": response}).Info("Message sent")
	return nil
}
