package connector

import "time"

// definition of generic constants
const (
	StepText     = 1
	StepMarkdown = 2
)

// InstallationToken can be used by a connector installation to propagte e.g. state changes
type InstallationToken string

// InstallationState reflects the current state of an installation
type InstallationState int

// definition of instantiation related constants
const (
	InstallationStateInitialized = 1
	InstallationStateTimeout     = 2
	InstallationStateOngoing     = 3
	InstallationStateRejected    = 4
	InstallationStateFailed      = 5
)

// InstallationRequest sent by connctd in order to signalise a new installation
type InstallationRequest struct {
	ID            string            `json:"id"`
	Token         InstallationToken `json:"token"`
	State         InstallationState `json:"state"`
	Configuration []Configuration   `json:"configuration"`
	Timestamp     time.Time         `json:"timestamp"`
}

// InstallationStateUpdateRequest can be sent by a connector to indicate new state
type InstallationStateUpdateRequest struct {
	State   InstallationState `json:"state"`
	Details string            `json:"details,omitempty"`
}

// InstallationResponse defines the struct of a potential response
type InstallationResponse struct {
	Details     string `json:"details,omitempty"`
	FurtherStep Step   `json:"furtherStep,omitempty"`
}

// InstanceToken maps to certain extern subject and can be used to e.g. propagate new things or device changes
type InstanceToken string

// InstanceState reflects the current state of an instantiation
type InstanceState int

// definition of instance related constants
const (
	InstanceStateInitialized = 1
	InstanceStateTimeout     = 2
	InstanceStateOngoing     = 3
	InstanceStateRejected    = 4
	InstanceStateFailed      = 5
)

// InstantiationRequest sent by connctd in order to signalise a new instantiation
type InstantiationRequest struct {
	ID            string          `json:"id"`
	Token         InstanceToken   `json:"token"`
	State         InstanceState   `json:"state"`
	Configuration []Configuration `json:"configuration"`
	Timestamp     time.Time       `json:"timestamp"`
}

// InstantiationResponse defines the struct of a potential response
type InstantiationResponse struct {
	Details     string `json:"details,omitempty"`
	FurtherStep Step   `json:"furtherStep,omitempty"`
}

// InstanceStateUpdateRequest can be sent by a connector to indicate a new state
type InstanceStateUpdateRequest struct {
	State   InstanceState `json:"state"`
	Details string        `json:"details,omitempty"`
}

// StepType defines the type of a further installation or instantiation step
type StepType int

// Step defines a further installation or instantiation step
type Step struct {
	Type  StepType `json:"type"`
	Value string   `json:"value"`
}

// Configuration is key value pair
type Configuration struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
}
