package connector

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"io/ioutil"
	"net/http"

	"github.com/connctd/api-go/crypto"
)

type signatureValidationHandler struct {
	preProcessor ValidationPreProcessor
	next         http.HandlerFunc
	publicKey    ed25519.PublicKey
}

// NewSignatureValidationHandler creates a new handler capable of verifying the signature header
// Validation can be influenced by passing a ValidationPreProcessor. Quite common
// functionalities are offered by DefaultValidationPreProcessor and ProxiedRequestValidationPreProcessor
func NewSignatureValidationHandler(validationPreProcessor ValidationPreProcessor, publicKey ed25519.PublicKey, next http.HandlerFunc) http.Handler {
	return &signatureValidationHandler{preProcessor: validationPreProcessor, publicKey: publicKey, next: next}
}

func (h *signatureValidationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	signature := r.Header.Get(crypto.SignatureHeaderKey)

	decodedSignature, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`Failed to decode signature signature`))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`Failed to read body`))
		return
	}

	defer r.Body.Close()

	// apply preprocess and pass values to signing function
	extractedValues := h.preProcessor(r)
	expectedSignature, err := crypto.SignablePayload(r.Method, extractedValues.Scheme, extractedValues.Host, extractedValues.RequestURI, r.Header, body)

	// lets check the signature manually
	if ed25519.Verify(h.publicKey, expectedSignature, decodedSignature) {
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
		h.next.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`Bad signature`))
		return
	}
}

// ValidationPreProcessor can be used to influence the signature validation algorithm
// by returning a modified url struct. This becomes handy if your service is sitting behind
// a proxy that modifies the original request headers which normally would lead to a
// validation error
type ValidationPreProcessor func(r *http.Request) ValidationParameters

// DefaultValidationPreProcessor extracts all relevant values from request fields
// Use this processor if there are no proxies between connctd platform and your connector
func DefaultValidationPreProcessor() ValidationPreProcessor {
	return func(r *http.Request) ValidationParameters {
		// on server side scheme is not populated: https://github.com/golang/go/issues/28940
		// manually picking https since this is currently the only supported protocol by connctd for callbacks
		return ValidationParameters{
			Scheme:     "https",
			Host:       r.URL.Host,
			RequestURI: r.URL.RequestURI(),
		}
	}
}

// ProxiedRequestValidationPreProcessor extracts all relevant values from request fields
func ProxiedRequestValidationPreProcessor(host string, endpoint string) ValidationPreProcessor {
	return func(r *http.Request) ValidationParameters {
		return ValidationParameters{
			Scheme:     "https",
			Host:       host,
			RequestURI: endpoint,
		}
	}
}

// ValidationParameters reflects list of parameters that are relevant for request signature validation
type ValidationParameters struct {
	Scheme     string
	Host       string
	RequestURI string
}
