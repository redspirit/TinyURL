package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	thttp "tinyurl/internal/transport/http"
)

func TestHTTP_Short_Redirect_Stats(t *testing.T) {
	d := NewTestDeps(t)
	h := thttp.NewHandlers(d.Svc, d.Log, "http://localhost:8080")
	router := thttp.NewRouter(h, d.Log)

	srv := httptest.NewServer(router)
	defer srv.Close()

	// 1) POST /shorten
	reqBody := `{"url":"https://example.com"}`
	resp, err := http.Post(srv.URL+"/shorten", "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	var res struct {
		Code     string `json:"code"`
		ShortURL string `json:"short_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if res.Code == "" || res.ShortURL == "" {
		t.Fatal("empty code or short_url")
	}

	// 2) GET /r/{code} -> 302 (не следуем редиректу)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
	}
	rResp, err := client.Get(srv.URL + "/r/" + res.Code)
	if err != nil {
		t.Fatal(err)
	}
	defer rResp.Body.Close()
	if rResp.StatusCode != http.StatusFound {
		t.Fatalf("expected 302, got %d", rResp.StatusCode)
	}
	loc := rResp.Header.Get("Location")
	if loc != "https://example.com" {
		t.Fatalf("unexpected redirect location: %s", loc)
	}

	// 3) GET /stats/{code}
	sResp, err := http.Get(srv.URL + "/stats/" + res.Code)
	if err != nil {
		t.Fatal(err)
	}
	defer sResp.Body.Close()
	if sResp.StatusCode != http.StatusOK {
		t.Fatalf("stats status: %d", sResp.StatusCode)
	}
	var stats struct {
		URL       string  `json:"url"`
		CreatedAt string  `json:"created_at"`
		ExpiresAt *string `json:"expires_at"`
		HitCount  int     `json:"hit_count"`
	}
	if err := json.NewDecoder(sResp.Body).Decode(&stats); err != nil {
		t.Fatal(err)
	}
	if stats.URL != "https://example.com" || stats.HitCount != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}
