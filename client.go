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
	"time"

	stdlog "log"

	"path"

	"github.com/connctd/restapi-go"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

const (
	// APIBaseURL defines how to reach connctd api
	APIBaseURL = "https://connectors.connctd.io/api/v1/"

	connectorThingsEndpoint            = "connectorhub/callback/instances/things"
	connectorInstanceStateEndpoint     = "connectorhub/callback/instances/state"
	connectorInstallationStateEndpoint = "connectorhub/callback/installations/state"
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
	UpdateThingPropertyValue(ctx context.Context, token InstantiationToken, thingID string, componentID string, propertyID string, value string, lastUpdate time.Time) error
	UpdateInstallationState(ctx context.Context, token InstallationToken, state InstallationState, details json.RawMessage) error
	UpdateInstanceState(ctx context.Context, token InstantiationToken, state InstantiationState, details json.RawMessage) error
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
		return restapi.Thing{}, fmt.Errorf("failed to marshal thing: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, a.baseURL.String()+connectorThingsEndpoint, bytes.NewBuffer(payload))
	if err != nil {
		return restapi.Thing{}, fmt.Errorf("failed to create new request: %w", err)
	}

	// set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+string(token))

	resp, err := a.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		a.logger.Error(err, "Failed to create thing", "name", thing.Name)
		return restapi.Thing{}, fmt.Errorf("failed to create thing: %w", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return restapi.Thing{}, fmt.Errorf("could not read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		a.logger.Error(ErrorUnexpectedStatusCode, "Could not create thing", "expectedStatusCode", http.StatusCreated, "givenStatusCode", resp.StatusCode, "body", string(body))
		return restapi.Thing{}, ErrorUnexpectedStatusCode
	}

	var res AddThingResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return restapi.Thing{}, fmt.Errorf("unable to unmarshal AddThingResponse: %w", err)
	}

	thing.ID = res.ID
	return thing, nil
}

// UpdateThingPropertyValue implements interface definition
func (a *APIClient) UpdateThingPropertyValue(ctx context.Context, token InstantiationToken, thingID string, componentID string, propertyID string, value string, lastUpdate time.Time) error {
	message := UpdateThingPropertyValueRequest{
		Value:      value,
		LastUpdate: lastUpdate,
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal thing update message: %w", err)
	}

	endpoint := path.Join(connectorThingsEndpoint, thingID, "components", componentID, "properties", propertyID)
	req, err := http.NewRequest(http.MethodPut, a.baseURL.String()+endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create new thing property value update request: %w", err)
	}

	// set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+string(token))

	resp, err := a.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		a.logger.Error(err, "Failed to update thing property", "thingId", thingID, "componentId", componentID, "propertyId", propertyID)
		return fmt.Errorf("failed to update thing property: %w", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read response body of update message: %w", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		a.logger.Error(ErrorUnexpectedStatusCode, "Could not update thing property", "expectedStatusCode", http.StatusNoContent, "givenStatusCode", resp.StatusCode, "body", string(body))
		return ErrorUnexpectedStatusCode
	}

	return nil
}

// UpdateInstallationState implements interface definition
func (a *APIClient) UpdateInstallationState(ctx context.Context, token InstallationToken, state InstallationState, details json.RawMessage) error {
	message := InstallationStateUpdateRequest{
		State:   state,
		Details: details,
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal installation state update message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, a.baseURL.String()+connectorInstallationStateEndpoint, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create installation state update message: %w", err)
	}

	// set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+string(token))

	resp, err := a.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		a.logger.Error(err, "Failed to send installation state update")
		return fmt.Errorf("failed to send installation state update: %w", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read response body of installation state update message: %w", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		a.logger.Error(ErrorUnexpectedStatusCode, "Could not update installation state", "expectedStatusCode", http.StatusNoContent, "givenStatusCode", resp.StatusCode, "body", string(body))
		return ErrorUnexpectedStatusCode
	}

	return nil
}

// UpdateThingPropertyValue implements interface definition
func (a *APIClient) UpdateInstanceState(ctx context.Context, token InstantiationToken, state InstantiationState, details json.RawMessage) error {
	message := InstanceStateUpdateRequest{
		State:   state,
		Details: details,
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal intance state update message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, a.baseURL.String()+connectorInstanceStateEndpoint, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create instance state update message: %w", err)
	}

	// set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+string(token))

	resp, err := a.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		a.logger.Error(err, "Failed to send instance state update")
		return fmt.Errorf("failed to send instance state update: %w", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read response body of instance state update message: %w", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		a.logger.Error(ErrorUnexpectedStatusCode, "Could not update instance state", "expectedStatusCode", http.StatusNoContent, "givenStatusCode", resp.StatusCode, "body", string(body))
		return ErrorUnexpectedStatusCode
	}

	return nil
}

// some error defintions
var (
	ErrorInvalidBaseURL       = errors.New("the base url needs to end with a slash")
	ErrorMissingLogger        = errors.New("a logger needs to be passed")
	ErrorUnexpectedStatusCode = errors.New("the resulting status code does not match with expectation")
)
