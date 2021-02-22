package main

import (
	"fmt"
	"sync"
	"testing"
	"time"

	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/test"
)

func updateNodes_installMAAS(t *testing.T) {
	t.Run("installMAAS", installMAAS)
}

func installMAAS(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.SaveTestResult()
	defer res.SetElapsedTime(time.Now(), "installMAAS")

	var (
		err    error
		maasId uint64
		lldpId uint64
	)

	if maasId, err = Pcc.FindRoleId(pcc.ROLE_MAAS, pcc.ROLE_MAAS_NEW); err == nil {
		if lldpId, err = Pcc.FindRoleId(pcc.ROLE_LLDP, pcc.ROLE_DEFAULT); err == nil {
			if nodes, err := Pcc.GetInvaderIds(); err == nil {
				if err = setRolesToNodesAndCheck([]uint64{lldpId, maasId}, "MAAS", nodes, MAAS_INSTALL_TIMEOUT); err != nil {
					res.SetTestFailure(err.Error())
					log.AuctaLogger.Error(err)
					t.FailNow()
				}
			} else {
				res.SetTestFailure(err.Error())
				log.AuctaLogger.Error(err)
				t.FailNow()
			}
		} else {
			res.SetTestFailure(err.Error())
			log.AuctaLogger.Error(err)
			t.FailNow()
		}
	} else {
		res.SetTestFailure(err.Error())
		log.AuctaLogger.Error(err)
		t.FailNow()
	}
	res.SetTestPass()
}

// FIXME move the wait to PccClient
func setRolesToNodesAndCheck(roles []uint64, app string, nodes []uint64, timeoutSec int) (err error) {
	var (
		installed    []uint64
		nodesToCheck []uint64
		check        bool
		wg           sync.WaitGroup
	)
	log.AuctaLogger.Infof("installing %s on nodes:%v\n", app, nodes)
	if installed, nodesToCheck, err = Pcc.AddRolesToNodes(nodes, roles); err == nil {
		if len(installed) > 0 {
			log.AuctaLogger.Warnf("%s already installed on nodes %d\n", app, installed)
		}
		if n := len(nodesToCheck); n > 0 {
			if timeoutSec <= 0 {
				timeoutSec = DEFAULT_TIMEOUT
			}
			timeout := time.Duration(timeoutSec * n)

			wg.Add(n)
			checkInstall := func(id uint64) {
				defer wg.Done()
				log.AuctaLogger.Infof("Checking %q installation for node: %d\n", app, id)

				start := time.Now()
				if check, err = Pcc.WaitForInstallation(id, timeout, app, "", &start); err != nil {
					err = fmt.Errorf("failed checking %s on %v: %v", app, id, err)
					return
				} else if check {
					log.AuctaLogger.Infof("%s correctly installed on nodeId:%v\n", app, id)
				}
			}

			//Check APP installation
			for i := 0; i < n; i++ {
				id := nodesToCheck[i]
				go checkInstall(id)
			}

			wg.Wait()
		}
	}

	return
}
