package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

type HTTPMethod string

type SettingOption func(*Settings)

const (
	GET            HTTPMethod = "GET"
	POST           HTTPMethod = "POST"
	PUT            HTTPMethod = "PUT"
	DELETE         HTTPMethod = "DELETE"
	DefaultTimeout            = 30 * time.Second
	APIVersion                = "/v3"
)

type Settings struct {
	BaseURL  string
	Token    string
	Verify   bool
	Timeout  time.Duration
	UseHTTP3 bool
	CACert   []byte
}

type requestOptions struct {
	URL    string
	header http.Header
	method HTTPMethod
	data   string
	stream bool
	params map[string]string
}

type BaseClient struct {
	settings Settings
	client   *http.Client
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

func NewRequestOptions(settings *Settings) *requestOptions {
	hdr := make(http.Header)
	hdr.Set("Content-Type", "application/json")
	hdr.Set("Authorization", fmt.Sprintf("Bearer %s", settings.Token))
	return &requestOptions{
		header: hdr,
		method: GET,
		params: make(map[string]string),
	}
}

func WithBaseURL(baseURL string) SettingOption {
	return func(s *Settings) {
		if baseURL != "" {
			s.BaseURL = baseURL + APIVersion
		}
	}
}

func WithBaseURLV2(baseURL string) SettingOption {
	return func(s *Settings) {
		if baseURL != "" {
			s.BaseURL = baseURL
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

func WithCA(cert []byte) SettingOption {
	return func(s *Settings) {
		s.CACert = cert
	}
}

func WithHTTP3(enabled bool) SettingOption {
	return func(s *Settings) {
		s.UseHTTP3 = enabled
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

// createTLSConfig builds a TLS config with three behaviours:
//
//  1. Pinned CA cert present – skip hostname verification but manually verify
//     the chain against the pinned CA.  This handles the common case where the
//     server cert was generated with different IPs/hostnames than the one the
//     client is connecting through (e.g. WiFi IP vs ethernet IP), yet trust was
//     already established via interactive fingerprint verification at bootstrap.
//
//  2. No CA cert, Verify=true – standard TLS using the system trust store
//     (hostname + chain both checked by Go's TLS stack).
//
//  3. No CA cert, Verify=false – skip everything (used for the initial
//     bootstrap connection before any cert has been pinned).
//
// TODO(ingress): case 1 (the InsecureSkipVerify workaround) exists purely
// because cert SANs are snapshotted from live IPs at startup and may not
// include every address a client connects through.  Once nodes get a stable
// ingress hostname/domain baked into their SANs (see TODO in
// pkg/web/certs/csr.go), clients can connect via that name and standard TLS
// verification will pass without needing this workaround.
func createTLSConfig(settings *Settings) *tls.Config {
	if len(settings.CACert) > 0 {
		caCertPool, err := x509.SystemCertPool()
		if err != nil || caCertPool == nil {
			caCertPool = x509.NewCertPool()
		}
		caCertPool.AppendCertsFromPEM(settings.CACert)

		return &tls.Config{
			InsecureSkipVerify: true, // #nosec G402 – chain verified below via VerifyPeerCertificate
			RootCAs:            caCertPool,
			VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
				if len(rawCerts) == 0 {
					return fmt.Errorf("server presented no certificates")
				}
				certs := make([]*x509.Certificate, 0, len(rawCerts))
				for _, raw := range rawCerts {
					cert, parseErr := x509.ParseCertificate(raw)
					if parseErr != nil {
						return fmt.Errorf("failed to parse server certificate: %w", parseErr)
					}
					certs = append(certs, cert)
				}
				intermediates := x509.NewCertPool()
				for _, cert := range certs[1:] {
					intermediates.AddCert(cert)
				}
				_, verifyErr := certs[0].Verify(x509.VerifyOptions{
					Roots:         caCertPool,
					Intermediates: intermediates,
				})
				return verifyErr
			},
		}
	}

	// No pinned CA – fall back to the simple verify flag.
	return &tls.Config{
		InsecureSkipVerify: !settings.Verify, // #nosec G402
	}
}

func NewBaseClient(settings *Settings) *BaseClient {
	var transport http.RoundTripper

	if settings.UseHTTP3 {
		transport = &http3.Transport{
			TLSClientConfig: createTLSConfig(settings),
			QUICConfig: &quic.Config{
				KeepAlivePeriod: 15 * time.Second,
				MaxIdleTimeout:  30 * time.Second,
			},
		}
	} else {
		transport = &http.Transport{
			TLSClientConfig:   createTLSConfig(settings),
			ForceAttemptHTTP2: true,
			MaxIdleConns:      100,
			IdleConnTimeout:   90 * time.Second,
		}
	}

	return &BaseClient{
		settings: *settings,
		client: &http.Client{
			Transport: transport,
			Timeout:   settings.Timeout,
		},
	}
}

func NewGNS3Client(settings *Settings) *GNS3ApiClient {
	tr := &http.Transport{
		TLSClientConfig: createTLSConfig(settings),
	}

	return &GNS3ApiClient{
		settings: *settings,
		client: &http.Client{
			Transport: tr,
			Timeout:   settings.Timeout,
		},
	}
}

func buildURL(baseURL string, opts *requestOptions) string {
	fullURL := baseURL + opts.URL
	if len(opts.params) > 0 {
		q := url.Values{}
		for k, v := range opts.params {
			q.Set(k, v)
		}
		fullURL += "?" + q.Encode()
	}
	return fullURL
}

func doRequest(ctx context.Context, client *http.Client, baseURL string, opts *requestOptions) ([]byte, *http.Response, error) {
	fullURL := buildURL(baseURL, opts)

	req, err := http.NewRequestWithContext(ctx, string(opts.method), fullURL, bytes.NewBufferString(opts.data))
	if err != nil {
		return nil, nil, err
	}
	req.Header = opts.header

	if opts.stream {
		streamClient := &http.Client{
			Transport: client.Transport,
		}
		resp, streamErr := streamClient.Do(req)
		if streamErr != nil {
			return nil, nil, streamErr
		}
		return nil, resp, nil
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}

	return body, resp, nil
}

func (c *BaseClient) Do(ctx context.Context, opts *requestOptions) ([]byte, *http.Response, error) {
	return doRequest(ctx, c.client, c.settings.BaseURL, opts)
}

func (c *GNS3ApiClient) Do(opts *requestOptions) ([]byte, *http.Response, error) {
	ctx := context.Background()
	if c.settings.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.settings.Timeout)
		defer cancel()
	}

	body, resp, err := doRequest(ctx, c.client, c.settings.BaseURL, opts)
	if err != nil {
		return nil, nil, err
	}

	if resp != nil && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		return body, resp, c.parseGNS3Error(body, resp.StatusCode)
	}

	return body, resp, nil
}

func (c *GNS3ApiClient) parseGNS3Error(body []byte, statusCode int) error {
	if statusCode == 422 {
		var payload any
		if err := json.Unmarshal(body, &payload); err == nil {
			switch v := payload.(type) {
			case map[string]any:
				if msg, ok := v["message"]; ok {
					if b, err := json.MarshalIndent(msg, "", "  "); err == nil {
						return fmt.Errorf("validation error (422):\n%s", string(b))
					}
				}
			case []any:
				if b, err := json.MarshalIndent(v, "", "  "); err == nil {
					return fmt.Errorf("validation error (422):\n%s", string(b))
				}
			}
		}
		return fmt.Errorf("validation error (422): %s", string(body))
	}

	if statusCode == 403 {
		var errorMsg map[string]string
		if err := json.Unmarshal(body, &errorMsg); err == nil {
			return fmt.Errorf("%s", errorMsg["message"])
		}
		return fmt.Errorf("unknown forbidden 403 error")
	}

	return fmt.Errorf("bad status %d: %s", statusCode, string(body))
}
