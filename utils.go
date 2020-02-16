package main

import (
	"time"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
)

func IsInvader(node *pcc.NodeWithKubernetes) bool {
	for i := 0; i < len(Env.Invaders); i++ {
		if Env.Invaders[i].HostIp == node.Host {
			return true
		}
	}
	return false
}

func ConvertToMillis(startTime time.Time) uint64 {
	return uint64(startTime.UnixNano()) / uint64(time.Millisecond)
}

func idInSlice(idToFind uint64, list []uint64) bool {
	for _, str := range list {
		if str == idToFind {
			return true
		}
	}
	return false
}
