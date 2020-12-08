package connector

import (
	"crypto/ed25519"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/connctd/api-go/crypto"

	"github.com/stretchr/testify/assert"
)

func TestSignatureVerificationHandler(t *testing.T) {
	assert := assert.New(t)

	pub, priv, err := ed25519.GenerateKey(nil)
	assert.Nil(err)

	okHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	// initialize with empty handler
	testServer := httptest.NewServer(nil)

	testServerURL, err := url.Parse(testServer.URL)
	assert.Nil(err)

	// create signature validation handler
	handler := http.Handler(NewSignatureValidationHandler(ProxiedRequestValidationPreProcessor(testServerURL.Scheme, testServerURL.Host, "/test"), pub, okHandler))
	testServer.Config = &http.Server{Handler: handler}

	// sign message on our own
	req, err := http.NewRequest("GET", testServer.URL+"/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = signRequest(priv, req, []byte{})
	assert.Nil(err)

	// this should work fine
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	// manipulated signature should lead to rejected request
	req.Header.Set(crypto.SignatureHeaderKey, "foobarsig")
	resp, err = http.DefaultClient.Do(req)
	assert.Nil(err)
	assert.Equal(ErrorBadSignature.Status, resp.StatusCode)
}

func signRequest(privateKey ed25519.PrivateKey, req *http.Request, body []byte) error {
	req.Header.Add("Date", time.Now().UTC().Format(http.TimeFormat))

	signable, err := crypto.SignablePayload(req.Method, req.URL.Scheme, req.Host, req.URL.RequestURI(), req.Header, body)
	if err != nil {
		return err
	}

	signature := ed25519.Sign(privateKey, signable)
	req.Header.Add(crypto.SignatureHeaderKey, base64.StdEncoding.EncodeToString(signature))

	return nil
}
