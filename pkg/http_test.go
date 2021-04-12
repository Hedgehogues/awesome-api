package pkg

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
)

var EmptyPageError = errors.New("empty page error")

type ResponseProcessMethod struct{ BaseMethod }

func NewResponseProcessMethod() *ResponseProcessMethod {
	countArgs := 0
	m, _ := NewBaseMethod("/", countArgs)
	m.Method = http.MethodGet
	m.Headers = []KVPair{{"cache-control", "no-cache"}}
	return &ResponseProcessMethod{BaseMethod: *m}
}

func (m *ResponseProcessMethod) ResponseProcess(body io.ReadCloser, h http.Header, s StatusCode, retries uint) (*Response, error) {
	buf := new(bytes.Buffer)
	n, err := buf.ReadFrom(body)
	if n <= 0 {
		return &Response{StatusCode: s}, EmptyPageError
	}
	if err != nil {
		return &Response{StatusCode: s}, err
	}
	return &Response{Bytes: buf.Bytes(), StatusCode: s}, nil
}

type DefaultMethod struct{ BaseMethod }

func NewDefaultMethod() *DefaultMethod {
	countArgs := 0
	m, _ := NewBaseMethod("/", countArgs)
	m.Method = http.MethodGet
	m.Headers = []KVPair{{"cache-control", "no-cache"}}
	return &DefaultMethod{BaseMethod: *m}
}

type FailedMethod struct{ BaseMethod }

func NewFailedMethod() *FailedMethod {
	countArgs := 0
	m, _ := NewBaseMethod("/123", countArgs)
	m.Method = http.MethodGet
	m.Headers = []KVPair{{"cache-control", "no-cache"}}
	m.QueryParams = map[string]string{"xxx": "123"}
	return &FailedMethod{BaseMethod: *m}
}

func (m *FailedMethod) ResponseProcess(body io.ReadCloser, h http.Header, s StatusCode, retries uint) (*Response, error) {
	buf := new(bytes.Buffer)
	n, err := buf.ReadFrom(body)
	if n <= 0 {
		return &Response{StatusCode: s}, EmptyPageError
	}
	if err != nil {
		return &Response{StatusCode: s}, err
	}
	return &Response{StatusCode: s}, nil
}

type ValidFailedMethod struct{ BaseMethod }

func NewValidFailedMethod() *ValidFailedMethod {
	countArgs := 0
	m, _ := NewBaseMethod("/123", countArgs)
	m.AcceptStatusCodes = []StatusCode{http.StatusNotFound}
	m.Method = http.MethodGet
	m.Headers = []KVPair{{"cache-control", "no-cache"}}
	m.QueryParams = map[string]string{"xxx": "123"}
	return &ValidFailedMethod{BaseMethod: *m}
}

func (m *ValidFailedMethod) ResponseProcess(body io.ReadCloser, h http.Header, s StatusCode, retries uint) (*Response, error) {
	buf := new(bytes.Buffer)
	n, err := buf.ReadFrom(body)
	if n <= 0 {
		return &Response{StatusCode: s}, EmptyPageError
	}
	if err != nil {
		return &Response{StatusCode: s}, err
	}
	return &Response{StatusCode: s}, nil
}

func TestClient_Request(t *testing.T) {
	type fields struct {
		url     string
		options *Options
	}
	type args struct {
		m Method
	}
	type want struct {
		code       StatusCode
		countRetry uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "ResponseProcessMethod",
			fields: fields{
				url:     "https://yandex.ru",
				options: &Options{},
			},
			args: args{m: NewResponseProcessMethod()},
			want: want{
				code:       http.StatusOK,
				countRetry: 0,
			},
			wantErr: false,
		},
		{
			name: "DefaultMethod",
			fields: fields{
				url:     "https://yandex.ru",
				options: &Options{},
			},
			args:    args{m: NewDefaultMethod()},
			want:    want{
				code: 200,
			},
			wantErr: false,
		},
		{
			name: "FailedMethod",
			fields: fields{
				url:     "https://yandex.ru",
				options: &Options{},
			},
			args: args{m: NewFailedMethod()},
			want: want{
				code:       http.StatusNotFound,
				countRetry: 0,
			},
			wantErr: false,
		},
		{
			name: "RetryMethod",
			fields: fields{
				url: "https://yandex.ru",
				options: &Options{
					CountRetry: 3,
				},
			},
			args: args{m: NewFailedMethod()},
			want: want{
				code:       http.StatusNotFound,
				countRetry: 3,
			},
			wantErr: false,
		},
		{
			name: "NoRetryFailedMethod",
			fields: fields{
				url: "https://yandex.ru",
				options: &Options{
					CountRetry: 3,
				},
			},
			args: args{m: NewValidFailedMethod()},
			want: want{
				code:       http.StatusNotFound,
				countRetry: 0,
			},
			wantErr: false,
		},
		{
			name: "NotFoundHost",
			fields: fields{
				url:     "https://yand1ex.ru",
				options: &Options{},
			},
			args:    args{m: NewDefaultMethod()},
			want:    want{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClientUrl(tt.fields.url, tt.fields.options)
			got, err := c.Request(tt.args.m, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Request() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && (got.StatusCode != tt.want.code || got.CountRetry != tt.want.countRetry) {
				t.Errorf("Client.Request() = %v, want %v", got, tt.want)
			}
		})
	}
}
