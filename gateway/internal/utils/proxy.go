package utils

import (
	"io"
	"net/http"
	"time"
)

func ProxyRequest(w http.ResponseWriter, r *http.Request, targetURL string) error {
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		return err
	}

	copyHeaders(req.Header, r.Header)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		return err
	}
	return nil
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
