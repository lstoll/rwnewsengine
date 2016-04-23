package rwnewsengine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"
)

var (
	readabilityApiUrl = "https://www.readability.com/api/content/v1/parser"
)

type ReadableError struct {
	Code    int
	Message string
}

func (r *ReadableError) Error() string {
	return fmt.Sprintf("HTTP-%d: %q", r.Code, r.Message)
}

func GetReadable(c *Config, pageURL string) (*ReadabilityOutput, error) {
	params := url.Values{"token": {c.ReadabilityKey}, "url": {pageURL}}
	response, err := http.Get(fmt.Sprintf("%s?%s", readabilityApiUrl, params.Encode()))
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		log.WithField("fn", "GetReadable").Errorf("Non-200 reponse code: %d", response.StatusCode)
		var body string
		if response.Body != nil {
			if bytes, err := ioutil.ReadAll(response.Body); err != nil {
				body = "[unavailable]"
			} else {
				body = string(bytes)
			}
			log.WithField("fn", "GetReadable").Errorf("Body: %q", body)
		}
		return nil, &ReadableError{Code: response.StatusCode, Message: body}
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	data := &ReadabilityOutput{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return data, nil
}
