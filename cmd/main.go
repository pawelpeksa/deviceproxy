package main

import (
	"time"

	"deviceproxy"
)

func main() {
	deviceProxy := deviceproxy.NewDeviceProxy()
	go deviceProxy.Run()
	defer deviceProxy.Shutdown()

	time.Sleep(time.Minute * 15)
}
