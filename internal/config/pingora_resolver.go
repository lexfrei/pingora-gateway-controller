package config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"time"

	"github.com/cockroachdb/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/lexfrei/pingora-gateway-controller/api/v1alpha1"
	routingv1 "github.com/lexfrei/pingora-gateway-controller/pkg/api/routing/v1"
)

const (
	// PingoraParametersRefGroup is the API group for PingoraConfig.
	PingoraParametersRefGroup = "pingora.k8s.lex.la"
	// PingoraParametersRefKind is the kind for PingoraConfig.
	PingoraParametersRefKind = "PingoraConfig"
)

// ResolvedPingoraConfig contains all configuration resolved from PingoraConfig and Secrets.
type ResolvedPingoraConfig struct {
	// gRPC endpoint address
	Address string

	// TLS configuration
	TLSEnabled            bool
	TLSCert               []byte
	TLSKey                []byte
	TLSCA                 []byte
	TLSInsecureSkipVerify bool
	TLSServerName         string

	// Connection parameters
	ConnectTimeout time.Duration
	RequestTimeout time.Duration
	KeepaliveTime  time.Duration
	MaxRetries     int32
	RetryBackoff   time.Duration

	// Reference to the source config for watch purposes
	ConfigName string
}

// PingoraResolver resolves PingoraConfig from GatewayClass parametersRef.
type PingoraResolver struct {
	client           client.Client
	defaultNamespace string
}

// NewPingoraResolver creates a new PingoraResolver.
func NewPingoraResolver(c client.Client, defaultNamespace string) *PingoraResolver {
	return &PingoraResolver{
		client:           c,
		defaultNamespace: defaultNamespace,
	}
}

// ResolveFromGatewayClass resolves PingoraConfig from a GatewayClass.
func (r *PingoraResolver) ResolveFromGatewayClass(
	ctx context.Context,
	gatewayClass *gatewayv1.GatewayClass,
) (*ResolvedPingoraConfig, error) {
	if gatewayClass.Spec.ParametersRef == nil {
		return nil, errors.New("GatewayClass has no parametersRef")
	}

	ref := gatewayClass.Spec.ParametersRef
	if string(ref.Group) != PingoraParametersRefGroup {
		//nolint:wrapcheck // errors.Newf creates a new error, not wrapping
		return nil, errors.Newf("unsupported parametersRef group: %s (expected %s)", ref.Group, PingoraParametersRefGroup)
	}

	if string(ref.Kind) != PingoraParametersRefKind {
		//nolint:wrapcheck // errors.Newf creates a new error, not wrapping
		return nil, errors.Newf("unsupported parametersRef kind: %s (expected %s)", ref.Kind, PingoraParametersRefKind)
	}

	config := &v1alpha1.PingoraConfig{}

	err := r.client.Get(ctx, types.NamespacedName{Name: ref.Name}, config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get PingoraConfig %s", ref.Name)
	}

	return r.resolveConfig(ctx, config)
}

// ResolveFromGatewayClassName resolves configuration by GatewayClass name.
func (r *PingoraResolver) ResolveFromGatewayClassName(
	ctx context.Context,
	gatewayClassName string,
) (*ResolvedPingoraConfig, error) {
	gatewayClass := &gatewayv1.GatewayClass{}

	err := r.client.Get(ctx, types.NamespacedName{Name: gatewayClassName}, gatewayClass)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get GatewayClass %s", gatewayClassName)
	}

	return r.ResolveFromGatewayClass(ctx, gatewayClass)
}

//nolint:funcorder // private helper
func (r *PingoraResolver) resolveConfig(ctx context.Context, config *v1alpha1.PingoraConfig) (*ResolvedPingoraConfig, error) {
	// Validate required address
	if config.Spec.Address == "" {
		return nil, errors.New("address is required in PingoraConfig")
	}

	resolved := &ResolvedPingoraConfig{
		Address:        config.Spec.Address,
		TLSEnabled:     config.Spec.IsTLSEnabled(),
		ConnectTimeout: time.Duration(config.Spec.GetConnectTimeout()) * time.Second,
		RequestTimeout: time.Duration(config.Spec.GetRequestTimeout()) * time.Second,
		KeepaliveTime:  time.Duration(config.Spec.GetKeepaliveTime()) * time.Second,
		MaxRetries:     config.Spec.GetMaxRetries(),
		RetryBackoff:   time.Duration(config.Spec.GetRetryBackoff()) * time.Millisecond,
		ConfigName:     config.Name,
	}

	// Resolve TLS configuration if enabled
	//nolint:nestif // TLS configuration requires checking multiple optional fields
	if resolved.TLSEnabled && config.Spec.TLS != nil {
		resolved.TLSInsecureSkipVerify = config.Spec.TLS.InsecureSkipVerify
		resolved.TLSServerName = config.Spec.TLS.ServerName

		if config.Spec.TLS.SecretRef != nil {
			secretRef := config.Spec.TLS.SecretRef

			secret, err := r.getSecret(ctx, secretRef.Name, secretRef.Namespace)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get TLS secret")
			}

			// Load TLS certificate and key
			if cert, ok := secret.Data["tls.crt"]; ok {
				resolved.TLSCert = cert
			}

			if key, ok := secret.Data["tls.key"]; ok {
				resolved.TLSKey = key
			}

			if ca, ok := secret.Data["ca.crt"]; ok {
				resolved.TLSCA = ca
			}
		}
	}

	return resolved, nil
}

//nolint:funcorder // private helper
func (r *PingoraResolver) getSecret(ctx context.Context, name, namespace string) (*corev1.Secret, error) {
	if namespace == "" {
		namespace = r.defaultNamespace
	}

	secret := &corev1.Secret{}

	err := r.client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, secret)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get secret %s/%s", namespace, name)
	}

	return secret, nil
}

// CreateGRPCConnection creates a gRPC connection to the Pingora proxy.
func (r *PingoraResolver) CreateGRPCConnection(_ context.Context, resolved *ResolvedPingoraConfig) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	// Set up keepalive
	opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                resolved.KeepaliveTime,
		Timeout:             resolved.ConnectTimeout,
		PermitWithoutStream: true,
	}))

	// Set up TLS or insecure
	if resolved.TLSEnabled {
		tlsConfig, err := r.buildTLSConfig(resolved)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build TLS config")
		}

		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Create connection using NewClient (DialContext is deprecated)
	conn, err := grpc.NewClient(resolved.Address, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to Pingora proxy at %s", resolved.Address)
	}

	return conn, nil
}

// CreateRoutingClient creates a gRPC routing service client.
func (r *PingoraResolver) CreateRoutingClient(conn *grpc.ClientConn) routingv1.RoutingServiceClient {
	return routingv1.NewRoutingServiceClient(conn)
}

//nolint:funcorder // private helper
func (r *PingoraResolver) buildTLSConfig(resolved *ResolvedPingoraConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: resolved.TLSInsecureSkipVerify, //nolint:gosec // user-configurable
	}

	if resolved.TLSServerName != "" {
		tlsConfig.ServerName = resolved.TLSServerName
	}

	// Load client certificate if provided
	if len(resolved.TLSCert) > 0 && len(resolved.TLSKey) > 0 {
		cert, err := tls.X509KeyPair(resolved.TLSCert, resolved.TLSKey)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load TLS certificate")
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate if provided
	if len(resolved.TLSCA) > 0 {
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(resolved.TLSCA) {
			return nil, errors.New("failed to parse CA certificate")
		}

		tlsConfig.RootCAs = caPool
	}

	return tlsConfig, nil
}

// GetConfigForGatewayClass returns the PingoraConfig for a GatewayClass.
//
//nolint:wrapcheck // errors.Newf creates new errors
func (r *PingoraResolver) GetConfigForGatewayClass(
	ctx context.Context,
	gatewayClass *gatewayv1.GatewayClass,
) (*v1alpha1.PingoraConfig, error) {
	if gatewayClass.Spec.ParametersRef == nil {
		return nil, errors.New("GatewayClass has no parametersRef")
	}

	ref := gatewayClass.Spec.ParametersRef
	if string(ref.Group) != PingoraParametersRefGroup || string(ref.Kind) != PingoraParametersRefKind {
		return nil, errors.Newf("unsupported parametersRef: %s/%s", ref.Group, ref.Kind)
	}

	config := &v1alpha1.PingoraConfig{}

	err := r.client.Get(ctx, types.NamespacedName{Name: ref.Name}, config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get PingoraConfig %s", ref.Name)
	}

	return config, nil
}
