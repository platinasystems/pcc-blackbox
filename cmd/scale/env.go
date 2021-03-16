package main

import pcc "github.com/platinasystems/pcc-blackbox/lib"

type Env struct {
	PccIp       string
	PccUser     string
	PccPassword string
	Nodes       []string
	DockerStats pcc.DockerStatsConfig
}
