package main

import (
	"fmt"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
	"sync"
	"testing"
	"time"
)

func updateNodes_installMAAS(t *testing.T) {
	t.Run("installMAAS", installMAAS)
}

func installMAAS(t *testing.T) {
	test.SkipIfDryRun(t)

	var (
		err    error
		maasId uint64
		lldpId uint64
	)

	if maasId, err = Pcc.FindRoleId(pcc.ROLE_MAAS, pcc.ROLE_MAAS_NEW); err == nil {
		if lldpId, err = Pcc.FindRoleId(pcc.ROLE_LLDP, pcc.ROLE_DEFAULT); err == nil {
			if nodes, err := Pcc.GetInvaderIds(); err == nil {
				if err = setRolesToNodesAndCheck([]uint64{lldpId, maasId}, "MAAS", nodes, MAAS_INSTALL_TIMEOUT); err != nil {
					t.Fatal(err)
				}
			} else {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	} else {
		t.Fatal(err)
	}
}

// FIXME move the wait to PccClient
func setRolesToNodesAndCheck(roles []uint64, app string, nodes []uint64, timeoutSec int) (err error) {
	var (
		installed    []uint64
		nodesToCheck []uint64
		check        bool
		wg           sync.WaitGroup
	)
	fmt.Printf("installing %s on nodes:%v\n", app, nodes)
	if installed, nodesToCheck, err = Pcc.AddRolesToNodes(nodes, roles); err == nil {
		if len(installed) > 0 {
			fmt.Printf("%s already installed on nodes %d\n", app, installed)
		}
		if n := len(nodesToCheck); n > 0 {
			if timeoutSec <= 0 {
				timeoutSec = DEFAULT_TIMEOUT
			}
			timeout := time.Duration(timeoutSec*n)

			wg.Add(n)
			checkInstall := func(id uint64) {
				defer wg.Done()
				fmt.Printf("Checking %q installation for node: %d\n", app, id)

				start := time.Now()
				if check, err = Pcc.WaitForInstallation(id, timeout, app, "", &start); err != nil {
					err = fmt.Errorf("failed checking %s on %v: %v", app, id, err)
					return
				} else if check {
					fmt.Printf("%s correctly installed on nodeId:%v\n", app, id)
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
