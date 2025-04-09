package kubernetes

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// PortForwardOptions defines options for port forwarding
type PortForwardOptions struct {
	Namespace    string
	ResourceName string
	APIVersion   string
	Kind         string
	Ports        []string      // Format: "localPort:containerPort"
	ReadyChan    chan struct{} // Signal when port forwarding is ready
	StopChan     chan struct{}
	Out          io.Writer
	ErrOut       io.Writer
}

// PortForward forwards ports from a Kubernetes resource (pod) to the local machine
func (k *Kubernetes) PortForward(ctx context.Context, options PortForwardOptions) error {
	// For now, only support port forwarding to pods since that's what client-go supports natively
	if strings.ToLower(options.Kind) != "pod" {
		return fmt.Errorf("port forwarding is only supported for pod resources, got %s", options.Kind)
	}

	// If namespace is not specified, use the configured one
	namespace := namespaceOrDefault(options.Namespace)

	// Create the request URL
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		namespace, options.ResourceName)
	u, err := url.Parse(fmt.Sprintf("%s%s", k.cfg.Host, path))
	if err != nil {
		return err
	}

	// Configure transport for SPDY
	transport, upgrader, err := spdy.RoundTripperFor(k.cfg)
	if err != nil {
		return err
	}

	// Create dialer
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, u)

	// Initialize the port forwarder
	fw, err := portforward.New(dialer, options.Ports, options.StopChan, options.ReadyChan, options.Out, options.ErrOut)
	if err != nil {
		return err
	}

	// Start port forwarding
	return fw.ForwardPorts()
}
