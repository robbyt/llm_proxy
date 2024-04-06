package schema

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sync"

	px "github.com/kardianos/mitmproxy/proxy"
	"github.com/robbyt/llm_proxy/schema/utils"
	log "github.com/sirupsen/logrus"
)

type TrafficObject struct {
	StatusCode       int         `json:"statusCode,omitempty"`
	Method           string      `json:"method,omitempty"`
	URL              *url.URL    `json:"url,omitempty"`
	Header           http.Header `json:"header"`
	Body             string      `json:"body"`
	Proto            string      `json:"proto,omitempty"`
	headersToFilter  []string    `json:"-"`
	headerFilterDone sync.Once   `json:"-"`
}

// UnmarshalJSON performs a non-threadsafe load of json data into THIS TrafficObject
func (t *TrafficObject) UnmarshalJSON(data []byte) error {
	r := make(map[string]any)
	err := json.Unmarshal(data, &r)
	if err != nil {
		return err
	}

	// handle status code
	if statusCode, ok := r["statusCode"]; ok {
		statusFloat, ok := statusCode.(float64)
		if !ok {
			return errors.New("statusCode parse error")
		}
		t.StatusCode = int(statusFloat)
	}

	// handle method
	method, ok := r["method"]
	if ok {
		t.Method = method.(string)
	}

	// handle URL
	rawURL, ok := r["url"]
	if ok {
		strURL, ok := rawURL.(string)
		if !ok {
			return errors.New("url parse error")
		}
		u, err := url.Parse(strURL)
		if err != nil {
			return err
		}
		t.URL = u
	}

	// handle headers
	rawheader, ok := r["header"].(map[string]any)
	if ok {
		header := make(map[string][]string)
		for k, v := range rawheader {
			vals, ok := v.([]any)
			if !ok {
				return errors.New("header parse error")
			}

			svals := make([]string, 0)
			for _, val := range vals {
				sval, ok := val.(string)
				if !ok {
					return errors.New("header parse error")
				}
				svals = append(svals, sval)
			}
			header[k] = svals
		}
		t.Header = header
	}

	// handle body
	body, ok := r["body"]
	if ok {
		t.Body, ok = body.(string)
		if !ok {
			return errors.New("body parse error")
		}
	}

	// handle proto
	proto, ok := r["proto"]
	if ok {
		t.Proto, ok = proto.(string)
		if !ok {
			return errors.New("proto parse error")
		}
	}

	t.filterHeaders()
	return nil
}

// MarshalJSON dumps this TrafficObject into a byte array containing JSON
func (t *TrafficObject) MarshalJSON() ([]byte, error) {
	var urlString string
	if t.URL != nil {
		urlString = t.URL.String()
	}

	type Alias TrafficObject
	return json.Marshal(&struct {
		URL string `json:"url,omitempty"`
		*Alias
	}{
		URL:   urlString,
		Alias: (*Alias)(t),
	})
}

func (t *TrafficObject) ToProxyResponse() *px.Response {
	return &px.Response{
		StatusCode: t.StatusCode,
		Header:     t.Header,
		Body:       []byte(t.Body),
	}
}

// HeaderString returns the headers as a flat string
func (t *TrafficObject) HeaderString() string {
	buf := new(bytes.Buffer)
	if err := t.Header.WriteSubset(buf, nil); err != nil {
		return ""
	}
	return buf.String()
}

// filterHeaders removes sensitive headers from the traffic object
func (t *TrafficObject) filterHeaders() {
	t.headerFilterDone.Do(func() {
		log.Debugf("Filtering headers: %v", t.headersToFilter)
		for _, header := range t.headersToFilter {
			t.Header.Del(header)
		}
	})
}

// NewTrafficObject creates a new TrafficObject with some fields populated
func NewTrafficObject(headers http.Header, body string, headersToFilter []string) *TrafficObject {
	to := &TrafficObject{
		Header:          headers,
		Body:            body,
		headersToFilter: headersToFilter,
	}
	to.filterHeaders()
	return to
}

// NewFromProxyRequest creates a new TrafficObject from a proxy request object
func NewFromProxyRequest(req *px.Request, headersToFilter []string) *TrafficObject {
	if req == nil {
		return nil
	}

	body, bodyIsPrintable := utils.CanPrintFast(req.Body)
	if !bodyIsPrintable {
		log.Warnf("Request body is not printable, skipping body: %s", req.URL)
	}

	tObj := &TrafficObject{
		Method:          req.Method,
		URL:             req.URL,
		Header:          req.Header,
		Proto:           req.Proto,
		Body:            body,
		headersToFilter: headersToFilter,
	}
	tObj.filterHeaders()
	return tObj
}

// NewFromProxyResponse creates a new TrafficObject from a proxy response object
func NewFromProxyResponse(resp *px.Response, headersToFilter []string) *TrafficObject {
	if resp == nil {
		return nil
	}

	body, bodyIsPrintable := utils.CanPrintFast(resp.Body)
	if !bodyIsPrintable {
		log.Warn("Response body is not printable, skipping")
	}

	tObj := &TrafficObject{
		StatusCode:      resp.StatusCode,
		Header:          resp.Header,
		Body:            body,
		headersToFilter: headersToFilter,
	}
	tObj.filterHeaders()
	return tObj
}

// NewFromJSONBytes unmarshals a JSON object into a TrafficObject
func NewFromJSONBytes(data []byte, headersToFilter []string) (*TrafficObject, error) {
	t := &TrafficObject{
		headersToFilter: headersToFilter,
	}
	err := json.Unmarshal(data, t)
	if err != nil {
		return nil, err
	}

	return t, nil
}
