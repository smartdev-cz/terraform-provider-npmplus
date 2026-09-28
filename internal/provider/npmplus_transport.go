package provider

import (
	"bytes"
	"io"
	"net/http"
	"sync"
)

type npmplusCookieTransport struct {
	base          http.RoundTripper
	mu            sync.RWMutex
	cookie        string
	lastErrorBody []byte
}

func (t *npmplusCookieTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.RLock()
	cookie := t.cookie
	t.mu.RUnlock()

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	response, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode >= 400 {
		body, readErr := io.ReadAll(response.Body)
		if readErr == nil {
			t.mu.Lock()
			t.lastErrorBody = append(t.lastErrorBody[:0], body...)
			t.mu.Unlock()
			response.Body.Close()
			response.Body = io.NopCloser(bytes.NewReader(body))
		}
	}

	for _, setCookie := range response.Cookies() {
		if setCookie.Name == "__Host-Http-token" {
			t.mu.Lock()
			t.cookie = setCookie.Name + "=" + setCookie.Value
			t.mu.Unlock()
			break
		}
	}

	return response, nil
}

func (t *npmplusCookieTransport) LastErrorBody() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return string(t.lastErrorBody)
}
