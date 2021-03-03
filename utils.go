package main

import (
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
