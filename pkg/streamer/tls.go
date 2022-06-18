package streamer

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// InitTlSConfig is a utility function to create a standard *tls.Config from path
// It is intentionally separated from NewStreamer to preserve re-usability
func InitTlSConfig(caCertPath, clientCertPath, clientKeyPath string) (*tls.Config, error) {
	tlsClientConfig := &tls.Config{}

	// Read the CA Certificates
	if len(caCertPath) > 0 {
		caCert, err := ioutil.ReadFile(caCertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificates bundle: %v", err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsClientConfig.ClientCAs = caCertPool
	}

	// Read the client certificate & key
	if len(clientCertPath) > 0 && len(clientKeyPath) > 0 {
		clientCertificate, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read client certificate and key pair: %v", err)
		}
		tlsClientConfig.Certificates = []tls.Certificate{clientCertificate}
	}

	return tlsClientConfig, nil
}
