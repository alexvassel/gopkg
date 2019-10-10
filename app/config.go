package app

import "github.com/severgroup-tt/gopkg-app/background"

type Config struct {
	Name       string
	Version    string
	Env        string
	Host       string
	HostAdmin  string
	Listener   ConfigListener
	Background []background.IService
}

type ConfigListener struct {
	Host          string
	HttpPort      int32
	HttpAdminPort int32
	GrpcPort      int32
}
