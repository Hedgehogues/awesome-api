package client

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	json "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"h12.io/socks"
)

const (
	ProtocolHttp   = "http"
	ProtocolHttps  = "https"
	ProtocolSocks4 = "socks4"
	ProtocolSocks5 = "socks5"
)

type Options struct {
	ClientTimeout *time.Duration

	DialerTimeout   *time.Duration
	DialerKeepAlive *time.Duration
	DialerDualStack *bool

	TransportMaxIdleConns          *int
	TransportIdleConnTimeout       *time.Duration
	TransportTLSHandshakeTimeout   *time.Duration
	TransportExpectContinueTimeout *time.Duration

	CountRetry uint8
}

func create(t *http.Transport, timeout *time.Duration) *http.Client {
	client := http.Client{
		Transport: t,
	}
	if timeout != nil {
		client.Timeout = *timeout
	}
	return &client
}

type Client struct {
	url     string
	options Options
}

func NewClientUrl(url string, opts *Options) *Client {
	if len(url) < 1 {
		return nil
	}
	if opts == nil {
		return &Client{
			url:     url,
			options: Options{},
		}
	}
	return &Client{
		url:     url,
		options: *opts,
	}
}

type StatusCode int16
type Response struct {
	api  map[string]interface{}
	text string

	Bytes      []byte
	Headers    []KVPair
	StatusCode StatusCode
	CountRetry uint
}

func (r *Response) Map() (map[string]interface{}, error) {
	var v map[string]interface{}
	err := json.Unmarshal(r.Bytes, &v)
	if err != nil {
		logrus.WithError(err).Error(ReadBodyError)
		return nil, UnmarshalResponseError
	}
	return v, nil
}

func (r *Response) Text() string {
	return string(r.Bytes)
}

func (r *Response) Struct(v interface{}) error {
	err := json.Unmarshal(r.Bytes, &v)
	if err != nil {
		logrus.WithError(err).Error(ReadBodyError)
		return UnmarshalResponseError
	}
	return nil
}

func NewClient(schema, host string, port *string, opts *Options) *Client {
	URL := fmt.Sprintf("%s://%s", schema, host)
	if port != nil {
		URL = fmt.Sprintf("%s://%s:%s", schema, host, *port)
	}
	return NewClientUrl(URL, opts)
}

func (c *Client) checkStatusCode(code StatusCode, m Method) bool {
	for _, s := range m.GetAcceptStatusCodes() {
		if code == s {
			return true
		}
	}
	return false
}

func (c *Client) retry(req *http.Request, client *http.Client, m Method) (*Response, error) {
	var resp *http.Response
	var statusCode StatusCode
	var countRetry uint
	var err error
	for i := 0; i < int(c.options.CountRetry)+1; i++ {
		resp, err = client.Do(req)
		if err != nil {
			countRetry = uint(i)
			continue
		}

		statusCode = StatusCode(resp.StatusCode)
		if c.checkStatusCode(statusCode, m) {
			break
		}
		countRetry = uint(i)
	}
	if err != nil {
		return nil, err
	}

	r, err := m.ResponseProcess(resp.Body, resp.Header, statusCode, countRetry)
	if err != nil {
		return nil, err
	}

	r.CountRetry = countRetry
	return r, nil
}

func (c *Client) Request(m Method, proxy *string) (*Response, error) {
	path, err := m.GetPath()
	if err != nil {
		return nil, err
	}

	method, err := m.GetMethod()
	if err != nil {
		return nil, err
	}

	t, err := c.buildTransport(proxy)
	if err != nil {
		return nil, err
	}
	// It needs to create clients for each request as we need different proxies
	client := create(t, c.options.ClientTimeout)

	b, err := m.GetBody()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, c.url+path, b)
	if err != nil {
		return nil, err
	}

	err = setQueryParams(req, m.GetQueryParams())
	if err != nil {
		return nil, err
	}
	setRequestHeaders(req, m.GetHeader())
	setRequestCookies(req, m.GetCookies())

	return c.retry(req, client, m)
}

func setQueryParams(req *http.Request, queryParams map[string]string) (err error) {
	q, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return err
	}
	for k, v := range queryParams {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	return nil
}

func setRequestHeaders(req *http.Request, headers []KVPair) {
	for _, p := range headers {
		req.Header.Add(p.Key, p.Value)
	}
}

func setRequestCookies(req *http.Request, cookies []KVPair) {
	for _, v := range cookies {
		req.AddCookie(&http.Cookie{
			Name:  v.Key,
			Value: v.Value,
		})
	}
}

func (c *Client) buildDialer() *net.Dialer {
	d := &net.Dialer{}
	if c.options.DialerTimeout != nil {
		d.Timeout = *c.options.DialerTimeout
	}
	if c.options.DialerKeepAlive != nil {
		d.KeepAlive = *c.options.DialerKeepAlive
	}
	if c.options.DialerDualStack != nil {
		d.DualStack = *c.options.DialerDualStack
	}
	return d
}

func (c *Client) initTransport(t *http.Transport) error {
	t.DialContext = c.buildDialer().DialContext
	if c.options.DialerKeepAlive != nil {
		t.MaxIdleConns = *c.options.TransportMaxIdleConns
	}
	if c.options.TransportIdleConnTimeout != nil {
		t.IdleConnTimeout = *c.options.TransportIdleConnTimeout
	}
	if c.options.DialerDualStack != nil {
		t.TLSHandshakeTimeout = *c.options.TransportTLSHandshakeTimeout
	}
	if c.options.DialerDualStack != nil {
		t.ExpectContinueTimeout = *c.options.TransportExpectContinueTimeout
	}
	return nil
}

func (c *Client) buildTransport(proxy *string) (*http.Transport, error) {
	transport := &http.Transport{}
	if proxy == nil || *proxy == "" {
		return transport, nil
	}

	parsedProxyUrl, err := url.Parse(*proxy)
	if err != nil {
		return nil, err
	}

	switch parsedProxyUrl.Scheme {
	case ProtocolHttp, ProtocolHttps:
		transport = &http.Transport{Proxy: http.ProxyURL(parsedProxyUrl)}
	case ProtocolSocks4:
		transport = &http.Transport{Dial: socks.DialSocksProxy(socks.SOCKS4, parsedProxyUrl.Host)}
	case ProtocolSocks5:
		transport = &http.Transport{Dial: socks.DialSocksProxy(socks.SOCKS5, parsedProxyUrl.Host)}
	default:
		transport = &http.Transport{}
	}
	err = c.initTransport(transport)
	if err != nil {
		return nil, err
	}

	return transport, nil
}
