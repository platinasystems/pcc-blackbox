package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/platinasystems/pcc-blackbox/models"

	"github.com/lib/pq"
	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

func reimageAllBrownNodes(t *testing.T) {
	t.Run("updateBmcInfo", updateBmcInfo)
	t.Run("reimageAllBrown", reimageAllBrown)
}

func updateBmcInfo(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res, Env.CheckServers, CheckNodes)

	assert := test.Assert{t}

	for _, i := range Env.Servers {
		var (
			addReq pcc.NodeDetailed
			err    error
			key    pcc.SecurityKey
		)
		if key, err = getFirstKey(); err == nil {
			keys := pq.Int64Array{int64(key.Id)}

			pBool := new(bool)
			*pBool = true

			addReq.Host = i.HostIp
			addReq.Id = NodebyHostIP[i.HostIp]
			addReq.Bmc = i.BMCIp
			addReq.BmcUser = i.BMCUser
			addReq.BmcUsers = i.BMCUsers
			addReq.BmcPassword = i.BMCPass
			addReq.AdminUser = "admin"
			addReq.SSHKeys = keys
			addReq.Managed = pBool
			addReq.Console = "ttyS1"

			if err = Pcc.UpdateNode(&addReq); err != nil {
				msg := fmt.Sprintf("Failed to update BMC info: %v", err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
			}
		} else {
			msg := fmt.Sprintf("Failed to get the key %v", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		}
	}
}

func reimageAllBrown(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res, Env.CheckServers, CheckNodes)

	assert := test.Assert{t}

	var (
		nodesToReimage = make(map[uint64]string)
		request        pcc.MaasRequest
		nodesReimaged  []uint64
	)

	fails := 0
	if key, err := getFirstKey(); err == nil {
		keys := []string{key.Alias}

		for _, s := range Env.Servers {
			id := NodebyHostIP[s.HostIp]
			imageName := "ubuntu-bionic"
			if len(Env.Maas.Reimage.Nodes) > 0 { // check if the nodes to reimage have been defined
				if n, ok := Env.Maas.Reimage.Nodes[s.HostIp]; ok {
					imageName = n.Image
				} else {
					continue // skip the node
				}
			}
			nodesToReimage[id] = imageName
		}

		request.Locale = "en-US"
		request.Timezone = "PDT"
		request.AdminUser = "admin"
		request.SSHKeys = keys

		log.AuctaLogger.Infof("Reimaging nodes %+v", nodesToReimage)

		for nodeID, _ := range nodesToReimage {
			request.Nodes = []uint64{nodeID}
			request.Image = nodesToReimage[nodeID]
			if err = Pcc.MaasDeploy(request); err != nil {
				msg := fmt.Sprintf("MaasDeploy failed: %v", err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
			}
		}

		log.AuctaLogger.Infof("Sleep for 8 minutes")
		time.Sleep(8 * time.Minute)

		for {
			for id, _ := range nodesToReimage {
				status, err := Pcc.GetProvisionStatus(id)
				if err != nil {
					log.AuctaLogger.Errorf("Node %v error: %v", id, err)
					fails++
					continue
				}
				if strings.Contains(status, "Ready") {
					log.AuctaLogger.Infof("Node %v has gone Ready", id)
					nodesReimaged = append(nodesReimaged, id)
					continue
				} else if strings.Contains(status, "reimage failed") {
					log.AuctaLogger.Errorf("Node %v has failed reimage", id)
					nodesReimaged = append(nodesReimaged, id)
					fails++
					continue
				}
				log.AuctaLogger.Infof("Node %v: %v", id, status)
			}
			if len(nodesReimaged) == len(nodesToReimage) { // all re-image completed
				if fails == 0 {
					log.AuctaLogger.Infof("Brownfield re-image done")
				} else {
					msg := fmt.Sprintf("Brownfield re-image failed on %v nodes", fails)
					res.SetTestFailure(msg)
					log.AuctaLogger.Error(msg)
					assert.FailNow()
				}
				checkAddNodesStatus(t, nodesReimaged)
				return
			}
			time.Sleep(60 * time.Second)
		}
	} else {
		msg := fmt.Sprintf("Failed to get the key %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}
