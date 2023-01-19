package etcd

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/pkg/transport"
	"google.golang.org/grpc"
	"k8s.io/apiserver/pkg/storage/storagebackend"
)

// GetEtcdClients returns an initialized  clientv3.Client and clientv3.KV.
func GetEtcdClients(config storagebackend.TransportConfig) (*clientv3.Client, clientv3.KV, error) {
	cfg := clientv3.Config{
		Endpoints:   config.ServerList,
		DialTimeout: 20 * time.Second,
		DialOptions: []grpc.DialOption{
			grpc.WithBlock(), // block until the underlying connection is up
		},
	}

	if config.CertFile != "" {
		tlsInfo := transport.TLSInfo{
			CertFile:      config.CertFile,
			KeyFile:       config.KeyFile,
			TrustedCAFile: config.TrustedCAFile,
		}

		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			return nil, nil, err
		}
		cfg.TLS = tlsConfig
	}

	c, err := clientv3.New(cfg)
	if err != nil {
		return nil, nil, err
	}

	return c, clientv3.NewKV(c), nil
}
