package rwnewsengine

import (
	"bytes"
	"fmt"
	"html/template"

	log "github.com/Sirupsen/logrus"
	"github.com/mvdan/xurls"
)

const htmlEmailTemplate = `
<html>
<head /><body>
<p>
{{.SourceEmail.StrippedText}}
</p>
<hr/>
<p>
{{.Parsed.Title}} by {{.Parsed.Author}}
</p>
<p>
{{.Parsed.Content}}
</p>
<p>
source: <a href="{{.Parsed.URL}}">{{.Parsed.URL}}</a>
</p>
</body>
</html>
`

// Just pass on the original. This sucks... but...
const plainEmailTemplate = `
{{.SourceEmail.BodyPlain}},
`

var (
	htmlBody  = template.Must(template.New("htmlemail").Parse(htmlEmailTemplate))
	plainBody = template.Must(template.New("plain").Parse(plainEmailTemplate))
)

func ProcessEmail(config *Config, email *InboundEmail) error {
	resp := &OutboundEmail{}
	// Find urls
	urls := xurls.Strict.FindAllString(email.StrippedText, -1)

	if len(urls) > 1 { //2manyurls
		resp.Recipient = email.Sender
		resp.Sender = email.Recipient
		resp.InReplyTo = email.MessageHeaders["message-id"]
		resp.Subject = fmt.Sprintf("RE: %s", email.Subject)
		resp.Body = "More than one URL was detected, this isn't supported"
		return SendEmail(config, resp)
	}

	if len(urls) == 0 { //no url, just proxy the email
		resp.Recipient = config.GroupAddress
		resp.Sender = email.From
		resp.Subject = email.Subject
		resp.Body = email.StrippedText
		return SendEmail(config, resp)
	}

	// We have a URL. Fetch it from readability
	parsed, err := GetReadable(config, urls[0])
	if err != nil {
		resp.Recipient = email.Sender
		resp.Sender = email.Recipient
		resp.InReplyTo = email.MessageHeaders["message-id"]
		resp.Subject = fmt.Sprintf("RE: %s", email.Subject)
		resp.Body = fmt.Sprintf("Error occured while talking to readable: %q", err)
		log.Info("Error talking to readability", err)
		return SendEmail(config, resp)
	}

	resp.Recipient = config.GroupAddress
	resp.Sender = email.From
	// Prefer the custom subject the sender used
	if email.Subject != "" {
		resp.Subject = email.Subject
	} else {
		resp.Subject = parsed.Title
	}

	var outBuf bytes.Buffer
	templVars := struct {
		SourceEmail *InboundEmail
		Parsed      *ReadabilityOutput
	}{
		SourceEmail: email,
		Parsed:      parsed,
	}
	plainBody.Execute(&outBuf, templVars)
	resp.Body = outBuf.String()
	htmlBody.Execute(&outBuf, templVars)
	resp.BodyHTML = outBuf.String()
	return SendEmail(config, resp)
}
