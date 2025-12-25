package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Default gRPC connection values.
const (
	DefaultGRPCPort       = 50051
	DefaultConnectTimeout = 5
	DefaultRequestTimeout = 30
	DefaultKeepaliveTime  = 30
	DefaultMaxRetries     = 3
	DefaultRetryBackoff   = 1000
)

// SecretReference contains the reference to a Secret.
type SecretReference struct {
	// Name is the name of the Secret.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Namespace is the namespace of the Secret.
	// If empty, the Secret is assumed to be in the same namespace as the referencing resource.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// TLSConfig configures TLS for gRPC connection to Pingora proxy.
type TLSConfig struct {
	// Enabled controls whether TLS is used for the gRPC connection.
	// +optional
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// SecretRef references a Secret containing TLS certificates.
	// The Secret must contain "tls.crt" and "tls.key" keys.
	// If CA validation is needed, include "ca.crt" key.
	// +optional
	SecretRef *SecretReference `json:"secretRef,omitempty"`

	// InsecureSkipVerify skips TLS certificate verification.
	// WARNING: This should only be used for testing.
	// +optional
	// +kubebuilder:default=false
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`

	// ServerName overrides the server name used for TLS verification.
	// +optional
	ServerName string `json:"serverName,omitempty"`
}

// ConnectionConfig configures the gRPC connection parameters.
type ConnectionConfig struct {
	// ConnectTimeoutSeconds is the timeout for establishing the connection.
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=5
	ConnectTimeoutSeconds *int32 `json:"connectTimeoutSeconds,omitempty"`

	// RequestTimeoutSeconds is the timeout for individual gRPC requests.
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=30
	RequestTimeoutSeconds *int32 `json:"requestTimeoutSeconds,omitempty"`

	// KeepaliveTimeSeconds is the interval for keepalive pings.
	// +optional
	// +kubebuilder:validation:Minimum=10
	// +kubebuilder:default=30
	KeepaliveTimeSeconds *int32 `json:"keepaliveTimeSeconds,omitempty"`

	// MaxRetries is the maximum number of retries for failed requests.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=3
	MaxRetries *int32 `json:"maxRetries,omitempty"`

	// RetryBackoffMs is the backoff duration between retries in milliseconds.
	// +optional
	// +kubebuilder:validation:Minimum=100
	// +kubebuilder:default=1000
	RetryBackoffMs *int32 `json:"retryBackoffMs,omitempty"`
}

// PingoraConfigSpec defines the desired state of PingoraConfig.
type PingoraConfigSpec struct {
	// Address is the gRPC endpoint address of the Pingora proxy.
	// Format: "host:port" (e.g., "pingora-proxy.pingora-system.svc.cluster.local:50051")
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Address string `json:"address"`

	// TLS configures TLS for the gRPC connection.
	// +optional
	TLS *TLSConfig `json:"tls,omitempty"`

	// Connection configures the gRPC connection parameters.
	// +optional
	Connection *ConnectionConfig `json:"connection,omitempty"`
}

// PingoraConfigStatus defines the observed state of PingoraConfig.
type PingoraConfigStatus struct {
	// Conditions describe the current state of the PingoraConfig.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Connected indicates whether the controller has successfully connected to the proxy.
	// +optional
	Connected bool `json:"connected,omitempty"`

	// LastSyncTime is the timestamp of the last successful route sync.
	// +optional
	LastSyncTime *metav1.Time `json:"lastSyncTime,omitempty"`

	// ConfigVersion is the current configuration version applied to the proxy.
	// +optional
	ConfigVersion uint64 `json:"configVersion,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=pgconfig
// +kubebuilder:printcolumn:name="Address",type=string,JSONPath=`.spec.address`
// +kubebuilder:printcolumn:name="TLS",type=boolean,JSONPath=`.spec.tls.enabled`
// +kubebuilder:printcolumn:name="Connected",type=boolean,JSONPath=`.status.connected`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// PingoraConfig is the Schema for the pingoraconfigs API.
// It provides configuration for connecting to a Pingora proxy.
type PingoraConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"` //nolint:modernize // kubebuilder standard

	Spec   PingoraConfigSpec   `json:"spec,omitempty"`   //nolint:modernize // kubebuilder standard
	Status PingoraConfigStatus `json:"status,omitempty"` //nolint:modernize // kubebuilder standard
}

// +kubebuilder:object:root=true

// PingoraConfigList contains a list of PingoraConfig.
type PingoraConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"` //nolint:modernize // kubebuilder standard

	Items []PingoraConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PingoraConfig{}, &PingoraConfigList{})
}

// IsTLSEnabled returns whether TLS is enabled for the connection.
func (c *PingoraConfigSpec) IsTLSEnabled() bool {
	return c.TLS != nil && c.TLS.Enabled
}

// GetConnectTimeout returns the connect timeout, defaulting to DefaultConnectTimeout.
func (c *PingoraConfigSpec) GetConnectTimeout() int32 {
	if c.Connection == nil || c.Connection.ConnectTimeoutSeconds == nil {
		return DefaultConnectTimeout
	}

	return *c.Connection.ConnectTimeoutSeconds
}

// GetRequestTimeout returns the request timeout, defaulting to DefaultRequestTimeout.
func (c *PingoraConfigSpec) GetRequestTimeout() int32 {
	if c.Connection == nil || c.Connection.RequestTimeoutSeconds == nil {
		return DefaultRequestTimeout
	}

	return *c.Connection.RequestTimeoutSeconds
}

// GetKeepaliveTime returns the keepalive time, defaulting to DefaultKeepaliveTime.
func (c *PingoraConfigSpec) GetKeepaliveTime() int32 {
	if c.Connection == nil || c.Connection.KeepaliveTimeSeconds == nil {
		return DefaultKeepaliveTime
	}

	return *c.Connection.KeepaliveTimeSeconds
}

// GetMaxRetries returns the max retries, defaulting to DefaultMaxRetries.
func (c *PingoraConfigSpec) GetMaxRetries() int32 {
	if c.Connection == nil || c.Connection.MaxRetries == nil {
		return DefaultMaxRetries
	}

	return *c.Connection.MaxRetries
}

// GetRetryBackoff returns the retry backoff, defaulting to DefaultRetryBackoff.
func (c *PingoraConfigSpec) GetRetryBackoff() int32 {
	if c.Connection == nil || c.Connection.RetryBackoffMs == nil {
		return DefaultRetryBackoff
	}

	return *c.Connection.RetryBackoffMs
}
