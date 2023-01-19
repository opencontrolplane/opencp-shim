package setup

import (
	"context"
	"log"

	config "github.com/opencontrolplane/opencp-shim/internal/config"
	etcd "github.com/opencontrolplane/opencp-shim/internal/etcd"
	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/klog/v2"
)

type OpenCPApp struct {
	Config            config.Config
	Token             string
	Context           context.Context
	EtcdClient        *clientv3.Client
	Namespace         opencpspec.NamespaceServiceClient
	LoginClient       opencpspec.LoginClient
	VirtualMachine    opencpspec.VirtualMachineServiceClient
	KubernetesCluster opencpspec.KubernetesClusterServiceClient
	Domain            opencpspec.DomainServiceClient
	SSHkey            opencpspec.SSHKeyServiceClient
	Firewall          opencpspec.FirewallServiceClient
}

func NewAOpenCP() *OpenCPApp {
	return &OpenCPApp{}
}

// Config returns global config struct
func Config(configPath string) config.Config {
	config, err := config.LoadConfig(configPath)
	if err != nil {
		panic(err)
	}
	return config
}

// Etcd returns etcd client
func Etcd(config config.Config) (*clientv3.Client, error) {
	etcdConfig := storagebackend.TransportConfig{
		ServerList: config.EtcdServer.Host,
	}
	etcdClient, _, err := etcd.GetEtcdClients(etcdConfig)
	if err != nil {
		log.Println(err)
	}
	defer etcdClient.Close()

	return etcdClient, err
}

func Login(config config.Config) opencpspec.LoginClient {
	conn, err := grpc.Dial(config.GrpcServer.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.Fatalf("could not connect: %v", err)
	}

	authClient := opencpspec.NewLoginClient(conn)
	return authClient
}

func Namespace(config config.Config) opencpspec.NamespaceServiceClient {
	conn, err := grpc.Dial(config.GrpcServer.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.Fatalf("could not connect: %v", err)
	}

	namespaceClient := opencpspec.NewNamespaceServiceClient(conn)
	return namespaceClient
}

func VirtualMachine(config config.Config) opencpspec.VirtualMachineServiceClient {
	conn, err := grpc.Dial(config.GrpcServer.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.Fatalf("could not connect: %v", err)
	}

	virtualMachineClient := opencpspec.NewVirtualMachineServiceClient(conn)
	return virtualMachineClient
}

func KubernetesCluster(config config.Config) opencpspec.KubernetesClusterServiceClient {
	conn, err := grpc.Dial(config.GrpcServer.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.Fatalf("could not connect: %v", err)
	}

	KubernetesClusterClient := opencpspec.NewKubernetesClusterServiceClient(conn)
	return KubernetesClusterClient
}

func Domain(config config.Config) opencpspec.DomainServiceClient {
	conn, err := grpc.Dial(config.GrpcServer.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.Fatalf("could not connect: %v", err)
	}

	DomainClient := opencpspec.NewDomainServiceClient(conn)
	return DomainClient
}

func SSHKey(config config.Config) opencpspec.SSHKeyServiceClient {
	conn, err := grpc.Dial(config.GrpcServer.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.Fatalf("could not connect: %v", err)
	}

	SSHkeyClient := opencpspec.NewSSHKeyServiceClient(conn)
	return SSHkeyClient
}

func Firewall(config config.Config) opencpspec.FirewallServiceClient {
	conn, err := grpc.Dial(config.GrpcServer.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.Fatalf("could not connect: %v", err)
	}

	FirewallClient := opencpspec.NewFirewallServiceClient(conn)
	return FirewallClient
}
