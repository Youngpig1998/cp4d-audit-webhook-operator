// ------------------------------------------------------ {COPYRIGHT-TOP} ---
// IBM Confidential
// Automated Tests
// Copyright IBM Corp. 2021
// ------------------------------------------------------ {COPYRIGHT-END} ---
package clients

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
)

// HTTPClient defines a HTTP Client that provides a number of useful testing functions
type HTTPClient struct {
	*http.Client
	username     string
	password     string
	hasBasicAuth bool
	uri          string
}

// HTTPClientBuilder is a utilty type to help build a HTTPClient
type HTTPClientBuilder struct {
	transport    *http.Transport
	username     string
	password     string
	hasBasicAuth bool
	uri          string
}

// NewHTTPClientBuilder returns a new HTTPClientBuilder for the given uri
func NewHTTPClientBuilder(uri string) *HTTPClientBuilder {
	return &HTTPClientBuilder{
		uri:          uri,
		hasBasicAuth: false,
		transport: &http.Transport{
			TLSClientConfig: &tls.Config{},
		},
	}
}

// WithBasicAuth adds basic authentiaction to the client with the given username and password
func (cb *HTTPClientBuilder) WithBasicAuth(username, password string) *HTTPClientBuilder {
	cb.username = username
	cb.password = password
	cb.hasBasicAuth = true
	return cb
}

// WithTLS add TLS to the client with the given CA certificate and whether the communication should
// be insecure
func (cb *HTTPClientBuilder) WithTLS(cert []byte, insecure bool) *HTTPClientBuilder {
	transportConfig := configureTransportSecurity(cert)
	transportConfig.InsecureSkipVerify = insecure
	cb.transport = &http.Transport{
		TLSClientConfig: transportConfig,
	}
	return cb
}

// Build constructs a HTTPClient from the builder
func (cb *HTTPClientBuilder) Build() *HTTPClient {
	client := &HTTPClient{
		Client: &http.Client{
			Transport: cb.transport,
		},
		username:     cb.username,
		password:     cb.password,
		hasBasicAuth: cb.hasBasicAuth,
		uri:          cb.uri,
	}
	return client
}

func configureTransportSecurity(certificate []byte) *tls.Config {
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certificate)
	return &tls.Config{
		RootCAs: certPool,
	}
}

type RequestParams struct {
	Method  string
	Path    string
	Body    io.Reader
	Headers map[string]string
}

// CustomsSendRequest sends a HTTP request to the provided path with for the given method
// it handles adding authentication, a request body can be optionally provided
// a set of headers can be optionally provided
func (c HTTPClient) CustomSendRequest(requestParams RequestParams) (*http.Response, error) {
	uri := fmt.Sprintf("%s/%s", c.uri, requestParams.Path)
	req, err := http.NewRequest(requestParams.Method, uri, requestParams.Body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	if requestParams.Body != nil {
		req.Header.Add("Content-type", "application/json")
	}

	// Set will override any previously set headers
	for key, value := range requestParams.Headers {
		req.Header.Set(key, value)
	}

	if c.hasBasicAuth {
		req.SetBasicAuth(c.username, c.password)
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, err
}

// SendRequest sends a HTTP request to the provided path with for the given method
// it handles adding authentication, a request body can be optionally provided
func (c HTTPClient) SendRequest(method, path string, body io.Reader) (*http.Response, error) {
	uri := fmt.Sprintf("%s/%s", c.uri, path)
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	if body != nil {
		req.Header.Add("Content-type", "application/json")
	}

	if c.hasBasicAuth {
		req.SetBasicAuth(c.username, c.password)
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, err
}
