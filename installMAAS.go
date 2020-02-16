package main

import (
	"fmt"
	"testing"
	"time"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

func updateNodes_installMAAS(t *testing.T) {
	t.Run("installMAAS", installMAAS)
	t.Run("verifyMAAS", verifyMAAS)
}

var nodesToCheck []uint64
var from time.Time

func installMAAS(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	from = time.Now()
	var isMAASInNodes = make(map[uint64]bool)
	for _, i := range Env.Invaders {
		var (
			addReq pcc.NodeWithKubernetes
			maas   []uint64 = []uint64{2, 6}
			err    error
		)

		id := NodebyHostIP[i.HostIp]
		addReq.Host = i.HostIp
		addReq.Id = id
		addReq.RoleIds = maas

		isMAASInNodes[id] = Pcc.IsAppInstalled(id, "maas")
		if !isMAASInNodes[id] {
			_, err = Pcc.UpdateNode(addReq)
			if err != nil {
				assert.Fatalf("Failed to install MaaS on id "+
					"%v : %v", id, err)
				return
			}
			nodesToCheck = append(nodesToCheck, id)
		} else {
			fmt.Printf("MAAS already installed in nodeId:%v\n",
				addReq.Id)
		}
	}
}

func verifyMAAS(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	var (
		err   error
		check bool
	)

	//Check MAAS installation
	for i := 0; i < len(nodesToCheck); i++ {
		id := nodesToCheck[i]
		fmt.Printf("Checking MAAS installation for nodeId:%v\n", id)

		check, err = checkMAASInstallation(id, from)
		if err != nil {
			assert.Fatalf("Failed checking MaaS on %v"+
				": %v", id, err)
			return
		}
		if check {
			fmt.Printf("MAAS correctly installed on nodeId:%v\n",
				id)
		}
	}
}
