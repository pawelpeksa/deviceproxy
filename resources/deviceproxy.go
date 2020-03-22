package resources

import (
	"fmt"
	"log"
	"os"
	"strconv"

	queuewrapper "git.krk.awesome-ind.com/GoUtils/QueueWrapper"
	stan "github.com/nats-io/go-nats-streaming"
)

const (
	ServiceName = "DeviceProxy"

	EnvDeviceProxyEndpoint = "DeviceProxyEndpoint"
	EnvServicePort         = "DeviceProxyServicePort"
	EnvDeviceProxyLogDebug = "DeviceProxyLogDebug"

	NATSEnvURLName     = "nats_URL_deviceproxy"
	NATSEnvClusterName = "nats_cluster_deviceproxy"
	NATSEnvUserName    = "nats_username_deviceproxy"
	NATSEnvPassName    = "nats_password_deviceproxy"
)

var (
	// DeviceProxyEndpoint ...
	DeviceProxyEndpoint = "/deviceproxy"
	// ServicePort ...
	ServicePort = "3001"
	// NATSURL says where to connects for NATS queueing
	NATSURL = ""
	// NATSClusterName holds NATS cluster name to connect to
	NATSClusterName = "FitStationCluster"
	// NATSUsername keeps username for NATS cluser
	NATSUsername = "nats"
	// NATSPassword keeps username for NATS cluser
	NATSPassword = "nats"
	// DeviceProxyLogDebug ...
	DeviceProxyLogDebug = false
)

func init() {
	initNATSEnvs()
	initServiceEnvs()
}

func connectionLostHandler(_ stan.Conn, reason error) {
	log.Fatalf("[FATAL] NATS connection lost and reconnection failed, reason: %v", reason)
}

// NewServiceMsgQueue ...
func NewServiceMsgQueue() queuewrapper.IMsgQueue {
	if NATSURL == "" {
		panic(fmt.Sprintf("[ERROR] Environment variable '%s' is empty or not set. NATS is required for this service, please make sure all enviroement variables (%s, %s, %s, %s) are set",
			NATSEnvURLName, NATSEnvURLName, NATSEnvClusterName, NATSEnvUserName, NATSEnvPassName))
	}

	return queuewrapper.NewMsgQueueNATS(NATSURL, NATSClusterName, ServiceName, NATSUsername, NATSPassword, connectionLostHandler)
}

func initNATSEnvs() {
	if url := os.Getenv(NATSEnvURLName); url != "" {
		NATSURL = url
	}
	if cluster := os.Getenv(NATSEnvClusterName); cluster != "" {
		NATSClusterName = cluster
	}
	if user := os.Getenv(NATSEnvUserName); user != "" {
		NATSUsername = user
	}
	if pwd := os.Getenv(NATSEnvPassName); pwd != "" {
		NATSPassword = pwd
	}
}

func initServiceEnvs() {
	if servicePort := os.Getenv(EnvServicePort); servicePort != "" {
		ServicePort = servicePort
	}

	if deviceEndpoint := os.Getenv(EnvDeviceProxyEndpoint); deviceEndpoint != "" {
		DeviceProxyEndpoint = deviceEndpoint
	}

	if deviceLogDebug := os.Getenv(EnvDeviceProxyLogDebug); deviceLogDebug != "" {
		DeviceProxyLogDebug, _ = strconv.ParseBool(deviceLogDebug)
	}
}
