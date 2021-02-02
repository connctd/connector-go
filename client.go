package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	stdlog "log"

	"github.com/connctd/restapi-go"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

const (
	// APIBaseURL defines how to reach connctd api
	APIBaseURL = "https://api.connctd.io/api/v1/"

	connectorThingsEndpoint = "connectorhub/callback/things"
)

// DefaultOptions returns default client options
func DefaultOptions() *ClientOptions {
	url, _ := url.Parse(APIBaseURL)

	return &ClientOptions{
		ConnctdBaseURL: url,
		HTTPClient:     http.DefaultClient,
	}
}

// DefaultLogger uses go standard logging capabilities
var DefaultLogger = stdr.New(stdlog.New(os.Stderr, "", stdlog.LstdFlags|stdlog.Lshortfile))

// Client defines api client functionalities
type Client interface {
	// CreateThing can be used to create a thing. A thingID is returned if
	// operation was successul. Otherwise an error is thrown.
	CreateThing(ctx context.Context, token InstantiationToken, thing restapi.Thing) (result restapi.Thing, err error)
}

// ClientOptions allow modification of api client behaviour
type ClientOptions struct {
	ConnctdBaseURL *url.URL
	HTTPClient     *http.Client
}

// APIClient implements Client interface
type APIClient struct {
	httpClient *http.Client
	baseURL    url.URL
	logger     logr.Logger
}

// NewClient creates a new api client
func NewClient(opts *ClientOptions, logger logr.Logger) (Client, error) {
	httpClient := http.DefaultClient
	url, _ := url.Parse(APIBaseURL)

	if logger == nil {
		return nil, ErrorMissingLogger
	}

	if opts != nil {
		if opts.HTTPClient != nil {
			httpClient = opts.HTTPClient
		}

		if opts.ConnctdBaseURL != nil {
			// url needs to end with slash
			if !strings.HasSuffix(opts.ConnctdBaseURL.String(), "/") {
				return nil, ErrorInvalidBaseURL
			}

			url = opts.ConnctdBaseURL
		}
	}

	return &APIClient{httpClient: httpClient, baseURL: *url, logger: logger.WithName("connector-go-client")}, nil
}

// CreateThing implements interface definition
func (a *APIClient) CreateThing(ctx context.Context, token InstantiationToken, thing restapi.Thing) (result restapi.Thing, err error) {
	message := AddThingRequest{
		Thing: thing,
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return restapi.Thing{}, fmt.Errorf("Failed to marshal thing: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, a.baseURL.String()+connectorThingsEndpoint, bytes.NewBuffer(payload))
	if err != nil {
		return restapi.Thing{}, fmt.Errorf("Failed to create new request: %w", err)
	}

	// set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+string(token))

	resp, err := a.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		a.logger.Error(err, "Failed to create thing", "name", thing.Name)
		return restapi.Thing{}, fmt.Errorf("Failed to create thing: %w", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return restapi.Thing{}, fmt.Errorf("Could not read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		a.logger.Error(ErrorUnexpectedStatusCode, "Could not create thing", "expectedStatusCode", http.StatusCreated, "givenStatusCode", resp.StatusCode, "body", string(body))
		return restapi.Thing{}, ErrorUnexpectedStatusCode
	}

	var res AddThingResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return restapi.Thing{}, fmt.Errorf("Unable to unmarshal AddThingResponse: %w", err)
	}

	thing.ID = res.ID
	return thing, nil
}

// some error defintions
var (
	ErrorInvalidBaseURL       = errors.New("The base url needs to end with a slash")
	ErrorMissingLogger        = errors.New("A logger needs to be passed")
	ErrorUnexpectedStatusCode = errors.New("The resulting status code does not match with expectation")
)
