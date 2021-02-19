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
	defer res.CheckTestAndSave(t, "updateBmcInfo")

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
				msg := fmt.Sprintf("Failed to update BMC info: %v\n", err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
				return
			}
		} else {
			msg := fmt.Sprintf("Failed to get the key %v\n", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
			return
		}
	}
}

func reimageAllBrown(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, "reimageAllBrown")

	assert := test.Assert{t}

	fails := 0
	if key, err := getFirstKey(); err == nil {
		keys := []string{key.Alias}

		nodesList := make([]uint64, len(Env.Servers))
		for i, s := range Env.Servers {
			nodesList[i] = NodebyHostIP[s.HostIp]
		}

		var request pcc.MaasRequest
		request.Nodes = nodesList
		request.Image = "centos78"
		request.Locale = "en-US"
		request.Timezone = "PDT"
		request.AdminUser = "admin"
		request.SSHKeys = keys

		fmt.Println(request)
		if err = Pcc.MaasDeploy(request); err != nil {
			msg := fmt.Sprintf("MaasDeploy failed: %v\n", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		}

		fmt.Println("Sleep for 8 minutes")
		time.Sleep(8 * time.Minute)

		for {
			for i, id := range nodesList {
				status, err := Pcc.GetProvisionStatus(id)
				if err != nil {
					log.AuctaLogger.Errorf("Node %v error: %v\n", id, err)
					fails++
					continue
				}
				if strings.Contains(status, "Ready") {
					log.AuctaLogger.Infof("Node %v has gone Ready\n", id)
					nodesList = removeIndex(i, nodesList)
					continue
				} else if strings.Contains(status, "reimage failed") {
					log.AuctaLogger.Errorf("Node %v has failed reimage\n", id)
					nodesList = removeIndex(i, nodesList)
					fails++
					continue
				}
				log.AuctaLogger.Infof("Node %v: %v\n", id, status)
			}
			if len(nodesList) == 0 {
				if fails == 0 {
					log.AuctaLogger.Infof("Brownfield re-image done\n")
				} else {
					msg := fmt.Sprintf("Brownfield re-image failed on %v nodes\n", fails)
					res.SetTestFailure(msg)
					log.AuctaLogger.Error(msg)
					assert.FailNow()
				}
				return
			}
			time.Sleep(60 * time.Second)
		}
	} else {
		msg := fmt.Sprintf("Failed to get the key %v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}

func removeIndex(i int, n []uint64) []uint64 {
	if len(n) > 1 {
		n = append(n[:i], n[i+1:]...)
		return n
	}
	return nil
}
