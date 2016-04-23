package rwnewsengine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestErrything(t *testing.T) {
	config := &Config{
		GroupAddress: "group@google.com",
	}
	HTTPSetup(config)
	cases := [...]struct {
		inboundBody         *string
		readabilityURL      string
		readabilityResponse *ReadabilityOutput
		matchSubject        *regexp.Regexp
		matchBody           *regexp.Regexp
		matchHTMLBody       *regexp.Regexp
		matchSender         *regexp.Regexp
		matchRecipient      *regexp.Regexp
	}{
		0: {
			// Basic email, with no link
			matchSubject:   regexp.MustCompile("Subject"),
			matchBody:      regexp.MustCompile("Body"),
			matchSender:    regexp.MustCompile(`lincoln.stoll@gmail.com`),
			matchRecipient: regexp.MustCompile("group@google.com"),
		},
		1: {
			// Email with subject, body and single link
			readabilityURL: "https://blog.golang.org/go1.6",
			readabilityResponse: &ReadabilityOutput{
				Content: "Go 1.6 is released",
			},
			matchSubject:   regexp.MustCompile("Go 1.6 has been released, or so it seems"),
			matchBody:      regexp.MustCompile("This is an email from the past"),
			matchHTMLBody:  regexp.MustCompile("Go 1.6 is released"),
			matchSender:    regexp.MustCompile(`lincoln.stoll@gmail.com`),
			matchRecipient: regexp.MustCompile("group@google.com"),
		},
		2: {
			// email with subject, body and two links. Should get rejected
			matchSubject:   regexp.MustCompile("RE: Go 1.6 and Go 1.5 released"),
			matchBody:      regexp.MustCompile("More than one URL"),
			matchSender:    regexp.MustCompile(`test@rwnews.lds.li`),
			matchRecipient: regexp.MustCompile("lincoln.stoll@gmail.com"),
		},
		3: {
			// Email with one link, and no subject. Should get article subject
			readabilityURL: "https://blog.golang.org/go1.4",
			readabilityResponse: &ReadabilityOutput{
				Title:   "Go 1.4 Release",
				Content: "Go 1.4 is released",
			},
			matchSubject:   regexp.MustCompile("Go 1.4 Release"),
			matchBody:      regexp.MustCompile("https://blog.golang.org/go1.4"),
			matchHTMLBody:  regexp.MustCompile("Go 1.4 is released"),
			matchSender:    regexp.MustCompile(`lincoln.stoll@gmail.com`),
			matchRecipient: regexp.MustCompile("group@google.com"),
		},
		4: {
			// Email with no link. Should just get passed on
			matchSubject:   regexp.MustCompile("Hey this is cool"),
			matchBody:      regexp.MustCompile("or is it?"),
			matchSender:    regexp.MustCompile(`lincoln.stoll@gmail.com`),
			matchRecipient: regexp.MustCompile("group@google.com"),
		},
	}
	for idx, tc := range cases {
		fatalf := func(format string, args ...interface{}) {
			t.Fatalf("Test case %d: %s", idx, fmt.Sprintf(format, args...))
		}
		fatal := func(args ...interface{}) {
			t.Fatal(fmt.Sprintf("Test case %d", idx), args)
		}

		// Load the fixture if not explicit
		if tc.inboundBody == nil {
			data, err := ioutil.ReadFile(fmt.Sprintf("fixtures/email/%d.formdata", idx))
			if err != nil {
				fatal("Error loading fixture from disk", err)
			}
			str := string(data)
			tc.inboundBody = &str
		}

		// Set up the email recorder
		var sentEmail *OutboundEmail
		emailSent := make(chan struct{})
		SendEmail = func(c *Config, e *OutboundEmail) error {
			sentEmail = e
			emailSent <- struct{}{}
			return nil
		}

		// Set up the readability API
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			pageURL := r.URL.Query().Get("url")

			if pageURL != tc.readabilityURL {
				fatalf("Expected readability url to be %q, got %q", tc.readabilityURL, pageURL)
			}

			data, err := json.Marshal(tc.readabilityResponse)
			if err != nil {
				fatal(err)
			}
			w.Write(data)
		}))
		readabilityApiUrl = ts.URL

		// Sent the request to the server
		resp := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/submit", strings.NewReader(*tc.inboundBody))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if err != nil {
			fatal(err)
		}
		http.DefaultServeMux.ServeHTTP(resp, req)
		if _, err := ioutil.ReadAll(resp.Body); err != nil {
			fatal(err)
		}
		if resp.Code != http.StatusOK {
			fatal("Non-200 returned")
		}

		timeout := make(chan struct{}, 1)
		go func() {
			time.Sleep(5 * time.Second)
			timeout <- struct{}{}
		}()
		select {
		case <-emailSent:
		case <-timeout:
		}

		if sentEmail == nil {
			fatal("No email sent")
		}

		// Check the email
		if tc.matchSubject != nil && !tc.matchSubject.MatchString(sentEmail.Subject) {
			fatalf("Outgoing subject %q does not match %q", sentEmail.Subject, tc.matchSubject)
		}
		if tc.matchBody != nil && !tc.matchBody.MatchString(sentEmail.Body) {
			fatalf("Outgoing plain body %q does not match %q", sentEmail.Body, tc.matchBody)
		}
		if tc.matchHTMLBody != nil && !tc.matchHTMLBody.MatchString(sentEmail.BodyHTML) {
			fatalf("Outgoing html body %q does not match %q", sentEmail.BodyHTML, tc.matchHTMLBody)
		}

		if tc.matchSender != nil && !tc.matchSender.MatchString(sentEmail.Sender) {
			fatalf("Outgoing sender %q does not match %q", sentEmail.Sender, tc.matchSender)
		}
		if tc.matchRecipient != nil && !tc.matchRecipient.MatchString(sentEmail.Recipient) {
			fatalf("Outgoing recipient %q does not match %q", sentEmail.Recipient, tc.matchRecipient)
		}

		ts.Close()
	}
}
