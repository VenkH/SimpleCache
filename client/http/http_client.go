package httpclient

import (
	"SimpleCache/common/peer"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type HttpGetter struct {
	BaseURL string
}

func (h *HttpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.BaseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

var _ peer.PeerGetter = (*HttpGetter)(nil)
