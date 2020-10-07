package connector

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestSignatureComposition(t *testing.T) {
	dummyPayload := struct {
		Hello string `json:"hello"`
	}{
		Hello: "world",
	}

	body, err := json.Marshal(dummyPayload)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://foo.com:8080/bar?hello=world", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create new request: %v", err)
	}

	fakeTime := time.Date(2020, 10, 7, 10, 0, 0, 0, time.Now().UTC().Location())
	req.Header.Set("Date", fakeTime.Format(http.TimeFormat))

	toBeSigned, err := SignablePayload(req.Method, req.URL.Scheme, req.Host, req.URL.RequestURI(), req.Header, body)
	if err != nil {
		t.Fatalf("Failed to create signable payload: %v", err)
	}

	expected := "(method):GET\r\n(url):https://foo.com:8080/bar?hello=world\r\n(Date):Wed, 07 Oct 2020 10:00:00 GMT\r\n(body):{\"hello\":\"world\"}"

	if string(toBeSigned) != expected {
		t.Fatalf("Generated payload and expected payload do not match.\nExpctd: %v\n Given: %v", expected, string(toBeSigned))
	}
}

func TestSignatureCompositionWithEmptyBody(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://foo.com:8080/bar?hello=world", nil)
	if err != nil {
		t.Fatalf("Failed to create new request: %v", err)
	}

	fakeTime := time.Date(2020, 10, 7, 10, 0, 0, 0, time.Now().UTC().Location())
	req.Header.Set("Date", fakeTime.Format(http.TimeFormat))

	toBeSigned, err := SignablePayload(req.Method, req.URL.Scheme, req.Host, req.URL.RequestURI(), req.Header, nil)
	if err != nil {
		t.Fatalf("Failed to create signable payload: %v", err)
	}

	expected := "(method):GET\r\n(url):https://foo.com:8080/bar?hello=world\r\n(Date):Wed, 07 Oct 2020 10:00:00 GMT\r\n(body):"

	if string(toBeSigned) != expected {
		t.Fatalf("Generated payload and expected payload do not match.\nExpctd: %v\n Given: %v", expected, string(toBeSigned))
	}
}
