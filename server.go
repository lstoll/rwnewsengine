package rwnewsengine

import (
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func HTTPSetup(config *Config) {
	emails := make(chan *InboundEmail)
	http.HandleFunc("/submit", inboundEmailHandler(config, emails))
	go func() {
		for {
			email := <-emails
			go func() {
				err := ProcessEmail(config, email)
				if err != nil {
					log.Error("Error processing email", err)
				}
			}()
		}
	}()
}

func HTTPServe(config *Config) {
	// TODO
	http.ListenAndServe(":"+config.Port, LogHTTPRequest(http.DefaultServeMux))
}

func inboundEmailHandler(config *Config, emails chan *InboundEmail) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		email := &InboundEmail{
			Recipient:      r.PostFormValue("recipient"),
			Sender:         r.PostFormValue("sender"),
			From:           r.PostFormValue("from"),
			Subject:        r.PostFormValue("subject"),
			BodyPlain:      r.PostFormValue("body-plain"),
			StrippedText:   r.PostFormValue("stripped-text"),
			BodyHTML:       r.PostFormValue("body-html"),
			StrippedHTML:   r.PostFormValue("stripped-html"),
			Token:          r.PostFormValue("token"),
			Signature:      r.PostFormValue("signature"),
			MessageHeaders: map[string]string{},
		}

		if hdrs := r.PostFormValue("message-headers"); hdrs != "" {
			headers := [][]string{}
			if err := json.Unmarshal([]byte(hdrs), &headers); err != nil {
				log.Error("Error unmarshaling message headers", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			for _, hdr := range headers {
				email.MessageHeaders[strings.ToLower(hdr[0])] = hdr[1]
			}
		}

		// Sanity check
		if email.Recipient == "" || email.From == "" || email.BodyPlain == "" {
			log.Errorf("Inbound email failed sanity check, email: %#v", email)
		}

		select {
		case emails <- email:
			log.Info("Message sent for processing")
		default:
		}

		w.WriteHeader(http.StatusOK)
	}
}

func LogHTTPRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO - this can be better
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
