package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type HTTP struct {
	http   *http.Client
	url    string
	header http.Header
}

func New(url string) *HTTP {
	return &HTTP{
		http: http.DefaultClient,
		url:  url,
	}
}

func (h *HTTP) SetHeader(header http.Header) {
	h.header = header
}

func (h *HTTP) Request(method string, path string, req interface{}, res interface{}) (int, error) {
	var data io.Reader = nil
	if req != nil {
		b, err := json.Marshal(req)
		if err != nil {
			return 0, err
		}
		data = bytes.NewReader(b)
	}

	fullurl := h.url + path
	// if err != nil {
	// 	return 0, err
	// }

	request, err := http.NewRequest(method, fullurl, data)
	if err != nil {
		return 0, err
	}

	if h.header != nil {
		request.Header = h.header
	}

	response, err := h.http.Do(request)
	if err != nil {
		return 0, err
	}

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return response.StatusCode, err
	}

	// fmt.Println("---", response, string(b))

	err = json.Unmarshal(b, res)
	if err != nil {
		return response.StatusCode, err
	}

	return response.StatusCode, nil
}
