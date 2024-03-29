// Package service implements a default service that can be used to quickly implement a connector.
// Together with connector.ConnectorHandler it implements the whole connector protocol.
// Developers can therefore focus on implementing technology specific details by implementing the Provider interface.
// It will listen to the provider.UpdateChannel() and update thing properties and action requests.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/connctd/connector-go/connctd"

	"github.com/connctd/connector-go"
	"github.com/go-logr/logr"
)

// ConnectorService provides the callback functions used by the HTTP handler.
type DefaultConnectorService struct {
	logger         logr.Logger
	db             connector.Database
	connctdClient  connector.Client
	provider       connector.Provider
	thingTemplates connector.ThingTemplates
	options        ConnectorServiceOptions
}

type ConnectorServiceOptions struct {
	// if true connector will immediately respond to instance creation requests and asynchronously
	// creates thing descriptions. Useful if there is a high chance that an instance needs to create
	// multiple things
	AsyncInstanceCreation bool

	// if true instance creation will fail if at least one thing can not be created. You cannot
	// enforce thing creation if asyncInstanceCreation is enabled
	EnforceThingCreation bool
}

var DefaultConnectorServiceOptions = ConnectorServiceOptions{
	AsyncInstanceCreation: false,
	EnforceThingCreation:  true,
}

// NewConnectorService returns a new instance of the default connector.
func NewConnectorService(dbClient connector.Database, connctdClient connector.Client, provider connector.Provider, thingTemplates connector.ThingTemplates, options ConnectorServiceOptions, logger logr.Logger) (*DefaultConnectorService, error) {
	// check for invalid settings
	if options.AsyncInstanceCreation && options.EnforceThingCreation {
		return nil, errors.New("enforced thing creation cant be enabled when async instance creation is enabled")
	}

	connector := &DefaultConnectorService{
		logger,
		dbClient,
		connctdClient,
		provider,
		thingTemplates,
		options,
	}

	err := connector.init()

	return connector, err
}

// init is called once during startup of the connector.
// It will register existing installations and instances with the provider.
func (s *DefaultConnectorService) init() error {
	installations, err := s.db.GetInstallations(context.Background())
	if err != nil {
		s.logger.Error(err, "Failed to retrieve instances from db")
		return fmt.Errorf("failed to retrieve installations from db: %v", err)
	}
	s.provider.RegisterInstallations(installations...)

	instances, err := s.db.GetInstances(context.Background())
	if err != nil {
		s.logger.Error(err, "Failed to retrieve instances from db")
		return fmt.Errorf("failed to retrieve instance from db: %v", err)
	}

	s.provider.RegisterInstances(instances...)

	return nil
}

// AddInstallation is called by the HTTP handler when it receives an installation request.
// It will persist the new installation and its configuration and register the new installation with the provider.
func (s *DefaultConnectorService) AddInstallation(ctx context.Context, request connector.InstallationRequest) (*connector.InstallationResponse, error) {
	s.logger.WithValues("installationRequest", request).Info("Received an installation request")

	if err := s.db.AddInstallation(ctx, request); err != nil {
		s.logger.WithValues("installationRequest", request).Error(err, "Failed to add installation")
		return nil, err
	}

	if len(request.Configuration) > 0 {
		if err := s.db.AddInstallationConfiguration(ctx, request.ID, request.Configuration); err != nil {
			s.logger.WithValues("config", request.Configuration).Error(err, "Failed to add installation configuration")
			return nil, err
		}
	}

	s.provider.RegisterInstallations(&connector.Installation{
		ID:            request.ID,
		Token:         request.Token,
		Configuration: request.Configuration,
	})

	return nil, nil
}

// RemoveInstallation is called by the HTTP handler when it receives an installation removal request.
// It will remove the installation from the database (including the installation token) and from the provider.
// Note that we will not be able to communicate with the connctd platform about the removed installation after this, since the token is deleted.
func (s *DefaultConnectorService) RemoveInstallation(ctx context.Context, installationId string) error {
	s.logger.WithValues("installationId", installationId).Info("Received an installation removal request")

	if err := s.provider.RemoveInstallation(installationId); err != nil {
		s.logger.WithValues("installationID", installationId).Error(err, "tried to remove installation that is not registered")
	}

	if err := s.db.RemoveInstallation(ctx, installationId); err != nil {
		s.logger.WithValues("installationId", installationId).Error(err, "failed to remove installation from db")
		return err
	}
	return nil
}

// AddInstantiation is called by the HTTP handler when it receives an instantiation request.
// It will persist the new instance, create new things for the instance
// and register the new instance with the provider.
func (s *DefaultConnectorService) AddInstance(ctx context.Context, request connector.InstantiationRequest) (*connector.InstantiationResponse, error) {
	s.logger.WithValues("instantiationRequest", request).Info("Received an instantiation request")

	if err := s.db.AddInstance(ctx, request); err != nil {
		s.logger.WithValues("instantiationRequest", request).Error(err, "Failed to add instance")
		return nil, err
	}

	if len(request.Configuration) > 0 {
		if err := s.db.AddInstanceConfiguration(ctx, request.ID, request.Configuration); err != nil {
			s.logger.WithValues("config", request.Configuration).Error(err, "Failed to add instance configuration")
			return nil, err
		}
	}

	thingTemplates := s.thingTemplates(request)

	if s.options.AsyncInstanceCreation {
		go s.synchronizeThings(context.Background(), request.ID, request.InstallationID, request.Token, request.Configuration, thingTemplates)
	} else {
		if err := s.synchronizeThings(ctx, request.ID, request.InstallationID, request.Token, request.Configuration, thingTemplates); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (s *DefaultConnectorService) synchronizeThings(ctx context.Context, instanceID string, installationID string, token connector.InstantiationToken, configuration []connector.Configuration, thingTemplates []connector.ThingTemplate) error {
	thingMapping := []connector.ThingMapping{}
	for _, template := range thingTemplates {
		thing, err := s.CreateThing(ctx, instanceID, template.Thing, template.ExternalID)
		if err != nil {
			s.logger.WithValues("thing", template).Error(err, "Failed to create new thing")

			// return error and abort instance creation
			if s.options.EnforceThingCreation && !s.options.AsyncInstanceCreation {
				s.logger.Info("Cancelling instance creation since enforeThingCreation is enabled")
				return err
			}

			continue
		}

		thingMapping = append(thingMapping, connector.ThingMapping{
			InstanceID: instanceID,
			ThingID:    thing.ID,
			ExternalID: template.ExternalID,
		})
	}

	s.provider.RegisterInstances(&connector.Instance{
		ID:             instanceID,
		InstallationID: installationID,
		Token:          token,
		ThingMapping:   thingMapping,
		Configuration:  configuration,
	})

	return nil
}

// RemoveInstance is called by the HTTP handler when it receives an instance removal request.
// It will remove the instance from the database (including the instance token) and from the provider.
// Note that we will not be able to communicate with the connctd platform about the removed instance after this, since the token is deleted.
func (s *DefaultConnectorService) RemoveInstance(ctx context.Context, instanceId string) error {
	s.logger.WithValues("instanceId", instanceId).Info("Received an instance removal request")

	if err := s.provider.RemoveInstance(instanceId); err != nil {
		s.logger.WithValues("instanceId", instanceId).Error(err, "tried to remove instance that is not registered")
	}

	if err := s.db.RemoveInstance(ctx, instanceId); err != nil {
		s.logger.WithValues("instanceId", instanceId).Error(err, "failed to remove instance from db")
		return err
	}
	return nil
}

// PerformAction is called by the HTTP handler when it receives an action request.
func (s *DefaultConnectorService) PerformAction(ctx context.Context, actionRequest connector.ActionRequest) (*connector.ActionResponse, error) {
	s.logger.WithValues("actionRequest", actionRequest).Info("Received an action request")

	instance, err := s.db.GetInstanceByThingId(ctx, actionRequest.ThingID)
	if err != nil {
		s.logger.WithValues("actionRequest", actionRequest).Error(err, "Could not retrieve the instance for thing ID")
		return &connector.ActionResponse{Status: connector.ActionRequestStatusFailed, Error: "thing ID was not found at connector"}, nil
	}

	status, err := s.provider.RequestAction(ctx, instance, actionRequest)
	if err != nil {
		s.logger.WithValues("actionRequest", actionRequest).Error(err, "failed to perform action")
		return &connector.ActionResponse{Status: status, Error: err.Error()}, err
	}

	switch status {
	case connector.ActionRequestStatusCompleted:
		// The action is completed.
		// We send no error and no response body and the handler will return status code 204.
		return nil, nil
	case connector.ActionRequestStatusPending:
		// The action is not completed yet.
		// We send no error but an ActionResponse and the handler will return status code 200.
		// We have to send an status update when the action is completed.
		return &connector.ActionResponse{Status: status}, nil
	case connector.ActionRequestStatusFailed:
		// This should not happen.
		// The provider is expected to return an error if the action failed, which we catch above.
		s.logger.WithValues("actionRequest", actionRequest).Error(errors.New("implementation should return an error if action failed"), "connector did not send an error but set action state to FAILED")
	}

	return nil, nil
}

// EventHandler handles events coming from the provider.
func (s *DefaultConnectorService) EventHandler(ctx context.Context) {
	// wait for update events
	go func() {
		for update := range s.provider.UpdateChannel() {
			var err error
			if update.PropertyUpdateEvent != nil {
				propertyUpdate := update.PropertyUpdateEvent
				err = s.UpdateProperty(ctx, propertyUpdate.InstanceId, propertyUpdate.ThingId, propertyUpdate.ComponentId, propertyUpdate.PropertyId, propertyUpdate.Value)
				if err != nil {
					s.logger.WithValues("propertyUpdate", propertyUpdate).Error(err, "failed to update property")
				}
			}
			if update.ActionEvent != nil {
				actionEvent := update.ActionEvent
				if err != nil {
					actionEvent.Response.Status = connector.ActionRequestStatusFailed
					actionEvent.Response.Error = fmt.Sprintf("failed to update property %v", err)
					s.logger.WithValues("actionEvent", actionEvent).Error(err, "action failed: failed to update property")
				}
				err := s.UpdateActionStatus(ctx, actionEvent.InstanceId, actionEvent.RequestId, actionEvent.Response)
				if err != nil {
					s.logger.WithValues("actionEvent", actionEvent).Error(err, "Failed to update action status")
				}
			}
		}
	}()
}

// CreateThing can be called by the connector to register a new thing for the given instance.
// It retrieves the instance token from the database and uses the token to create a new thing via the connctd API client.
// The new thing ID is then stored in the database referencing the instance id.
func (s *DefaultConnectorService) CreateThing(ctx context.Context, instanceId string, thing connctd.Thing, externalId string) (*connctd.Thing, error) {
	instance, err := s.db.GetInstance(ctx, instanceId)
	if err != nil {
		s.logger.WithValues("instanceId", instanceId).Error(err, "failed to retrieve instance from database")
		return nil, err
	}

	// CreateThing() will create the thing at the connctd platform.
	// Since the platform will manage the thing, we only need to store its ID.
	createdThing, err := s.connctdClient.CreateThing(ctx, instance.Token, thing)
	if err != nil {
		s.logger.WithValues("thing", thing).Error(err, "failed to register new Thing")
		return nil, err
	}

	// Save the thing ID with the instance, so we have a mapping of things to instances.
	err = s.db.AddThingMapping(ctx, instanceId, createdThing.ID, externalId)
	if err != nil {
		s.logger.WithValues("thing", thing).Error(err, "failed to insert new Thing into database")
		return nil, err
	}

	s.logger.WithValues("thing", createdThing).Info("Created new thing")

	return &createdThing, nil
}

// UpdateProperty can be called by the connector to update a component property of a thing belonging to an instance.
func (s *DefaultConnectorService) UpdateProperty(ctx context.Context, instanceId, thingId, componentId, propertyId, value string) error {
	instance, err := s.db.GetInstance(ctx, instanceId)
	if err != nil {
		s.logger.WithValues("instanceId", instanceId).Error(err, "failed to retrieve instance")
		return err
	}

	timestamp := time.Now()

	// Use the client from the SDK to update the action status
	err = s.connctdClient.UpdateThingPropertyValue(ctx, instance.Token, thingId, componentId, propertyId, value, timestamp)

	return err
}

// UpdateActionStatus can be called by the connector to update the status of an action request.
func (s *DefaultConnectorService) UpdateActionStatus(ctx context.Context, instanceId string, actionRequestId string, actionResponse *connector.ActionResponse) error {
	instance, err := s.db.GetInstance(ctx, instanceId)
	if err != nil {
		s.logger.WithValues("instanceId", instanceId).Error(err, "failed to retrieve instance")
		return err
	}

	// Use the client from the SDK to update the action status
	return s.connctdClient.UpdateActionStatus(ctx, instance.Token, actionRequestId, actionResponse.Status, actionResponse.Error)
}
