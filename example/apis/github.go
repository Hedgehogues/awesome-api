package apis

import (
	"io"
	"io/ioutil"
	"net/http"

	cl "github.com/hedgehogues/xxx"
	"github.com/sirupsen/logrus"
)

type CatAPI struct {
	cl.BaseMethod
}

type RequestCatAPI struct {
}

type ResponseCatAPI struct {
	ID      int     `json:"id"`
	URL     string  `json:"url"`
	WebpUrl string  `json:"webpurl"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
}

func NewCatAPI(req *RequestCatAPI) (*CatAPI, error) {
	m, err := cl.NewBaseMethod("/catapi/rest/", 0)
	if err != nil {
		return nil, err
	}
	m.Method = http.MethodGet
	return &CatAPI{
		BaseMethod: *m,
	}, nil
}

func (m *CatAPI) ResponseProcess(body io.ReadCloser, h http.Header, s cl.StatusCode, retries uint) (*cl.Response, error) {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		logrus.WithError(err).Error(cl.ReadBodyError)
		return nil, cl.ReadBodyError
	}
	err = body.Close()
	if err != nil {
		logrus.WithError(err).Error(cl.CloseBodyError)
		return nil, cl.CloseBodyError
	}
	var headers []cl.KVPair
	for k, vs := range h {
		for _, v := range vs {
			headers = append(headers, cl.KVPair{k, v})
		}
	}
	return &cl.Response{Bytes: b, StatusCode: s, Headers: headers, CountRetry: retries}, nil
}
