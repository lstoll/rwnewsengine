package rwnewsengine

// InboundEmail https://documentation.mailgun.com/user_manual.html#routes
type InboundEmail struct {
	Recipient      string
	Sender         string
	From           string
	Subject        string
	BodyPlain      string
	StrippedText   string
	BodyHTML       string
	StrippedHTML   string
	Token          string
	Signature      string
	MessageHeaders map[string]string
}

type OutboundEmail struct {
	Recipient string
	Sender    string
	Subject   string
	Body      string
	InReplyTo string
}

type ReadabilityOutput struct {
	Author  string
	Content string
	Domain  string
	Title   string
	URL     string
}
