package deviceproxy

import (
	"fmt"
	"log"
	"net/http"

	"deviceproxy/api"
	"deviceproxy/resources"
	"deviceproxy/server"
)

// DeviceProxy ...
type DeviceProxy struct {
	server *server.Server
}

// NewDeviceProxy ...
func NewDeviceProxy() *DeviceProxy {
	return &DeviceProxy{}
}

// Run ...
func (p *DeviceProxy) Run() {

	p.server = server.NewServer()
	msgQueue := resources.NewServiceMsgQueue()
	api := api.NewAPI(p.server, msgQueue)
	p.server.Listener = api

	log.Printf("Starting DeviceProxy %v on port %v\n", resources.ServiceName, resources.ServicePort)

	err := p.server.Serve(resources.ServicePort)

	if err == http.ErrServerClosed {
		log.Println("DeviceProxy stopped")
		return
	}

	if err != nil {
		panic(fmt.Sprintf("DeviceProxy error when serving:%v", err))
	}
}

// Shutdown ...
func (p *DeviceProxy) Shutdown() error {
	return p.server.Shutdown()
}
