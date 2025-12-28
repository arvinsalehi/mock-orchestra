// internal/core/types.go
package core

import "time"

type Device struct {
	DeviceID string
	Product  string
	Serial   string
	Addr     string // ip:22
	Group    string
	LastSeen time.Time
	Meta     map[string]string
}

type Discovery interface {
	Start() error
	Stop() error
	Updates() <-chan Device // emits upserts
}

type Registry interface {
	Upsert(Device)
	ByGroup(group string) []Device
	All() []Device
}

type Runner interface {
	Identify(addr string) (product, serial string, err error)
	DeployAndRun(addr string, cmd RunCommand) error
}

type RunCommand struct {
	SessionID   string
	BuildNumber string
	Plan        string
	RemoteDir   string
}

