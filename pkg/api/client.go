package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type HTTPMethod string

type SettingOption func(*Settings)

const (
	GET            HTTPMethod = "GET"
	POST           HTTPMethod = "POST"
	PUT            HTTPMethod = "PUT"
	DELETE         HTTPMethod = "DELETE"
	DefaultTimeout            = 30 * time.Second
	API_VERSION               = "/v3"
)

type Settings struct {
	BaseURL string
	Token   string
	Verify  bool
	Timeout time.Duration
}

type requestOptions struct {
	settings Settings
	URL      string
	header   http.Header
	method   HTTPMethod
	data     string
	stream   bool
	params   map[string]string
}

type GNS3ApiClient struct {
	settings Settings
	client   *http.Client
}

func NewSettings(opts ...SettingOption) Settings {
	s := Settings{
		BaseURL: "",
		Token:   "",
		Verify:  true,
		Timeout: DefaultTimeout,
	}
	for _, opt := range opts {
		opt(&s)
	}
	return s
}

func NewGNS3Client(settings Settings) *GNS3ApiClient {
	tr := &http.Transport{}
	if !settings.Verify {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &GNS3ApiClient{
		settings: settings,
		client: &http.Client{
			Transport: tr,
			Timeout:   settings.Timeout,
		},
	}
}

func NewRequestOptions(settings Settings) *requestOptions {
	hdr := make(http.Header)
	hdr.Set("Content-Type", "application/json")
	hdr.Set("Authorization", fmt.Sprintf("Bearer %s", settings.Token))
	return &requestOptions{
		settings: settings,
		header:   hdr,
		method:   GET,
		params:   make(map[string]string),
	}
}

func WithBaseURL(url string) SettingOption {
	return func(s *Settings) {
		if url != "" {
			s.BaseURL = url + API_VERSION
		}
	}
}

func WithToken(token string) SettingOption {
	return func(s *Settings) {
		s.Token = token
	}
}

func WithVerify(verify bool) SettingOption {
	return func(s *Settings) {
		s.Verify = verify
	}
}

func WithTimeout(d time.Duration) SettingOption {
	return func(s *Settings) {
		if d > 0 {
			s.Timeout = d
		}
	}
}
func (r *requestOptions) WithURL(path string) *requestOptions {
	r.URL = path
	return r
}

func (r *requestOptions) WithMethod(m HTTPMethod) *requestOptions {
	r.method = m
	return r
}

func (r *requestOptions) WithData(body string) *requestOptions {
	r.data = body
	return r
}

func (r *requestOptions) WithParam(key, val string) *requestOptions {
	r.params[key] = val
	return r
}

func (r *requestOptions) WithStream() *requestOptions {
	r.stream = true
	return r
}

func (c *GNS3ApiClient) Do(opts *requestOptions) ([]byte, *http.Response, error) {
	fullURL := c.settings.BaseURL + opts.URL

	if len(opts.params) > 0 {
		q := url.Values{}
		for k, v := range opts.params {
			q.Set(k, v)
		}
		fullURL += "?" + q.Encode()
	}

	if opts.stream {
		streamClient := *c.client
		streamClient.Timeout = 0

		req, err := http.NewRequest(string(opts.method), fullURL, bytes.NewBufferString(opts.data))
		if err != nil {
			return nil, nil, err
		}
		req.Header = opts.header

		resp, err := streamClient.Do(req)
		if err != nil {
			return nil, nil, err
		}

		if c.settings.Timeout > 0 {
			go func() {
				<-time.After(c.settings.Timeout)
				resp.Body.Close()
			}()
		}

		return nil, resp, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.settings.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx,
		string(opts.method),
		fullURL,
		bytes.NewBufferString(opts.data),
	)
	if err != nil {
		return nil, nil, err
	}
	req.Header = opts.header

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == 422 {
			var anyPayload any
			if err := json.Unmarshal(body, &anyPayload); err == nil {
				switch v := anyPayload.(type) {
				case map[string]any:
					if msg, ok := v["message"]; ok {
						if b, err := json.MarshalIndent(msg, "", "  "); err == nil {
							return body, resp, fmt.Errorf("validation error (422):\n%s", string(b))
						}
					}
				case []any:
					if b, err := json.MarshalIndent(v, "", "  "); err == nil {
						return body, resp, fmt.Errorf("validation error (422):\n%s", string(b))
					}
				}
			}
			return body, resp, fmt.Errorf("validation error (422): %s", string(body))
		}
		if resp.StatusCode == 403 {
			var errorMsg map[string]string
			if err := json.Unmarshal(body, &errorMsg); err == nil {
				return body, resp, fmt.Errorf("%s", errorMsg["message"])
			}
			return body, resp, fmt.Errorf("Unknown forbidden 403 error.")
		}
		return body, resp, fmt.Errorf("bad status %d: %s", resp.StatusCode, body)
	}

	return body, resp, nil
}
