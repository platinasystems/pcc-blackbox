package main

import ()

type testEnv struct {
	PccIp    string
	Invaders []invader
	Servers  []server
}

type node struct {
	HostIp        string
	BMCIp         string
	BMCUser       string
	BMCUsers      []string
	BMCPass       string
	KeyId         []uint64
	KeyAlias      []string
	NetInterfaces []netInterface
}

type netInterface struct {
	Name         string
	Cidrs        []string
	MacAddr      string
	IsManagement bool
	ManagedByPcc bool
	Speed        string
	Autoneg      string
	Fec          string
	Media        string
	Mtu          uint
}

type invader struct {
	node
}

type server struct {
	node
}

var exampleEnv = testEnv{
	PccIp: "172.17.2.238",
	Invaders: []invader{
		invader{
			node{
				HostIp:   "172.17.2.60",
				BMCIp:    "172.17.3.60",
				BMCUser:  "ADMIN",
				BMCUsers: []string{"ADMIN"},
				BMCPass:  "ADMIN",
				KeyId:    []uint64{3},
				KeyAlias: []string{"test"},
				NetInterfaces: []netInterface{
					netInterface{
						Name:  "eth0",
						Cidrs: []string{"172.17.2.60/23"},
					},
				},
			},
		},
	},
}
