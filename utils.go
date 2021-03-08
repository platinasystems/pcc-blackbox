package main

import (
	"errors"
	"fmt"
	log "github.com/platinasystems/go-common/logs"
	"github.com/platinasystems/pcc-blackbox/models"
	"testing"
	"time"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
)

func IsInvader(node *pcc.NodeDetailed) bool {
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

func getNodeFromEnv(id uint64) *node {
	for i := range Env.Invaders {
		node := Env.Invaders[i].node
		if node.Id == id {
			return &node
		}
	}

	for i := range Env.Servers {
		node := Env.Servers[i].node
		if node.Id == id {
			return &node
		}
	}

	return nil
}

func CheckDependencies(t *testing.T, res *models.TestResult, dep ...func() error) {
	for _, fn := range dep {
		if err := fn(); err != nil {
			log.AuctaLogger.Errorf("%s", err.Error())
			res.SetTestSkipped(err.Error())
			t.SkipNow()
		}
	}
}

func CheckInvaders(nodes *[]pcc.NodeDetailed) (err error) {
	for _, node := range Env.Invaders {
		found := false
		isOnline := false
		for i := 0; i < len(*nodes); i++ {
			remoteNode := (*nodes)[i]
			if node.HostIp == remoteNode.Host {
				found = true
				if remoteNode.Status == "OK" {
					isOnline = true
				}
				break
			}
		}
		if !found || !isOnline {
			msg := fmt.Sprintf("%s not found in PCC or not online", node.HostIp)
			err = errors.New(msg)
			return
		}
	}
	return
}

func CountInvadersMatching(nodes *[]pcc.NodeDetailed) (numInvaders int) {
	for _, node := range Env.Invaders {
		for i := 0; i < len(*nodes); i++ {
			remoteNode := (*nodes)[i]
			if node.HostIp == remoteNode.Host && remoteNode.Status == "OK" {
				numInvaders++
			}
		}
	}
	return
}

func CheckServers(nodes *[]pcc.NodeDetailed) (err error) {
	for _, node := range Env.Servers {
		found := false
		isOnline := false
		for i := 0; i < len(*nodes); i++ {
			remoteNode := (*nodes)[i]
			if node.HostIp == remoteNode.Host {
				found = true
				if remoteNode.Status == "online" {
					isOnline = true
				}
				break
			}
		}
		if !found || !isOnline {
			msg := fmt.Sprintf("%s not found in PCC or not online", node.HostIp)
			err = errors.New(msg)
			return
		}
	}
	return
}

func CountServersMatching(nodes *[]pcc.NodeDetailed) (numServers int) {
	for _, node := range Env.Servers {
		for i := 0; i < len(*nodes); i++ {
			remoteNode := (*nodes)[i]
			if node.HostIp == remoteNode.Host && remoteNode.Status == "OK" {
				numServers++
			}
		}
	}
	return
}

func CheckNodes() (err error) {
	var nodes *[]pcc.NodeDetailed
	if nodes, err = Pcc.GetNodes(); err == nil {
		if err = CheckInvaders(nodes); err != nil {
			return
		}
		if err = CheckServers(nodes); err != nil {
			return
		}
	} else {
		msg := fmt.Sprintf("Error getting nodes: %v\n", err)
		err = errors.New(msg)
		return
	}
	return
}

func CheckNumNodes(numNodes int) (err error) {
	var nodes *[]pcc.NodeDetailed
	if nodes, err = Pcc.GetNodes(); err == nil {
		if remoteNumNodes := len(*nodes); remoteNumNodes < numNodes {
			msg := fmt.Sprintf("%d nodes present,but %d nodes required", remoteNumNodes, numNodes)
			err = errors.New(msg)
			return
		}
		numInvaders := CountInvadersMatching(nodes)
		numServers := CountServersMatching(nodes)
		if (numServers + numInvaders) < numNodes {
			msg := fmt.Sprintf("%d matching nodes ,but %d nodes required", numServers+numInvaders, numNodes)
			err = errors.New(msg)
			return
		}

	} else {
		msg := fmt.Sprintf("Error getting nodes: %v\n", err)
		err = errors.New(msg)
		return
	}
	return
}
