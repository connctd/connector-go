package connector

import (
	"encoding/json"
	"time"
)

// definition of generic constants
const (
	StepText     = 1
	StepMarkdown = 2
	StepRedirect = 3
)

// InstallationToken can be used by a connector installation to propagte e.g. state changes
type InstallationToken string

// InstallationState reflects the current state of an installation
type InstallationState int

// definition of instantiation related constants
const (
	InstallationStateInitialized = 1
	// all further states are set internally by connctd
)

// InstallationRequest sent by connctd in order to signalise a new installation
type InstallationRequest struct {
	ID            string            `json:"id"`
	Token         InstallationToken `json:"token"`
	State         InstallationState `json:"state"`
	Configuration []Configuration   `json:"configuration"`
	MessageID     string            `json:"messageId"`
	Sent          time.Time         `json:"sent"`
}

// InstallationStateUpdateRequest can be sent by a connector to indicate new state
type InstallationStateUpdateRequest struct {
	State   InstallationState `json:"state"`
	Details json.RawMessage   `json:"details,omitempty"`
}

// InstallationResponse defines the struct of a potential response
type InstallationResponse struct {
	Details     json.RawMessage `json:"details,omitempty"`
	FurtherStep Step            `json:"furtherStep,omitempty"`
}

// InstantiationToken can be used by a connector instance to send things or updates
type InstantiationToken string

// InstantiationState reflects the current state of an instantiation
type InstantiationState int

// definition of instantiation states
const (
	InstantiationStateInitialized = 1
)

// InstantiationRequest sent by connctd in order to signalise a new instantiation
type InstantiationRequest struct {
	ID             string             `json:"id"`
	InstallationID string             `json:"installation_id"`
	Token          InstantiationToken `json:"token"`
	State          InstantiationState `json:"state"`
	Configuration  []Configuration    `json:"configuration"`
	Timestamp      time.Time          `json:"timestamp"`
}

// InstantiationResponse defines the struct of a potential response
type InstantiationResponse struct {
	Details     json.RawMessage `json:"details,omitempty"`
	FurtherStep Step            `json:"furtherStep,omitempty"`
}

// InstanceStateUpdateRequest can be sent by a connector to indicate a new state
type InstanceStateUpdateRequest struct {
	State   InstantiationState `json:"state"`
	Details json.RawMessage    `json:"details,omitempty"`
}

// StepType defines the type of a further installation or instantiation step
type StepType int

// Step defines a further installation or instantiation step
type Step struct {
	Type    StepType `json:"type"`
	Content string   `json:"content"`
}

// Configuration is key value pair
type Configuration struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}
