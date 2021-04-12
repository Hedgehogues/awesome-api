package pkg

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

type KVPair struct {
	Key   string
	Value string
}

type Method interface {
	GetPath() (string, error)
	GetMethod() (string, error)
	GetHeader() []KVPair
	GetCookies() []KVPair
	GetQueryParams() map[string]string
	GetBody() (io.Reader, error)
	GetAcceptStatusCodes() []StatusCode

	ResponseProcess(body io.ReadCloser, h http.Header, s StatusCode, retries uint) (*Response, error)
}

type BaseMethod struct {
	Path              string
	Count             int
	Method            string // see: http.MethodGet, http.MethodPost ...
	Headers           []KVPair
	Cookies           []KVPair
	QueryParams       map[string]string
	Body              []byte
	AcceptStatusCodes []StatusCode
}

func NewBaseMethod(path string, countArgs int, args ...string) (*BaseMethod, error) {
	if len(args) != countArgs {
		return nil, CountArgsError
	}
	if len(path) == 0 {
		return nil, PathError
	}
	as := make([]interface{}, 0, len(args))
	for _, a := range args {
		as = append(as, a)
	}
	p := fmt.Sprintf(path, as...)
	return &BaseMethod{
		Path:              p,
		Method:            http.MethodGet,
		Headers:           []KVPair{},
		Cookies:           []KVPair{},
		QueryParams:       map[string]string{},
		AcceptStatusCodes: []StatusCode{http.StatusOK},
	}, nil
}

func (m *BaseMethod) GetPath() (string, error) {
	if len(m.Path) == 0 {
		return "", EmptyPathError
	}
	return m.Path, nil
}

// GetMethod is returned
func (m *BaseMethod) GetMethod() (string, error) {
	if len(m.Method) == 0 {
		return "", EmptyMethodError
	}
	return m.Method, nil
}

// GetQueryParams method is returned list of Headers
func (m *BaseMethod) GetHeader() []KVPair {
	return m.Headers
}

// GetQueryParams method is returned list of cookies
func (m *BaseMethod) GetCookies() []KVPair {
	return m.Cookies
}

// GetQueryParams method is returned list of query params of request
func (m *BaseMethod) GetQueryParams() map[string]string {
	return m.QueryParams
}

// ResponseProcess method is returned io.Reader with body of request (maybe empty)
func (m *BaseMethod) GetBody() (io.Reader, error) {
	if m.Method == http.MethodGet && m.Body != nil {
		return nil, NotEmptyBodyError
	}
	if m.Body == nil {
		return nil, nil
	}
	return bytes.NewBuffer(m.Body), nil
}

// ResponseProcess method is returned list of succeed codes after request to server
func (m *BaseMethod) GetAcceptStatusCodes() []StatusCode {
	return m.AcceptStatusCodes
}

// ResponseProcess method is called after request to server
func (m *BaseMethod) ResponseProcess(body io.ReadCloser, h http.Header, s StatusCode, retries uint) (*Response, error) {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		logrus.WithError(err).Error(ReadBodyError)
		return nil, ReadBodyError
	}
	err = body.Close()
	if err != nil {
		logrus.WithError(err).Error(CloseBodyError)
		return nil, CloseBodyError
	}
	var headers []KVPair
	for k, vs := range h {
		for _, v := range vs {
			headers = append(headers, KVPair{k, v})
		}
	}
	return &Response{Bytes: b, StatusCode: s, Headers: headers, CountRetry: retries}, nil
}
