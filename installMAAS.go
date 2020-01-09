package main

import (
	"encoding/json"
	"fmt"
	"github.com/platinasystems/test"
	"github.com/platinasystems/tiles/pccserver/models"
	"testing"
	"time"
)

func updateNodes_installMAAS(t *testing.T) {
	t.Run("installMAAS", installMAAS)
}

func installMAAS(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		body       []byte
		resp       HttpResp
		err        error
		check      bool
	)

	from := time.Now()
	var isMAASInNodes = make(map[uint64]bool)
	var nodesToCheck []uint64
	for _, i := range Env.Invaders {
		var (
			addReq models.NodeWithAdditionalFields
			node   models.NodeWithKubernetes
			data   []byte
			maas   []uint64 = []uint64{2, 6}
		)

		addReq.Host = i.HostIp
		addReq.Id = NodebyHostIP[i.HostIp]
		addReq.RoleIds = maas

		isMAASInNodes[addReq.Id] = IsAppInstalled(addReq.Id, "lldpd")
		if !isMAASInNodes[addReq.Id] {
			endpoint := fmt.Sprintf("pccserver/node/update")
			if data, err = json.Marshal(addReq); err != nil {
				assert.Fatalf("invalid struct for node update request")
			}
			if resp, body, err = pccGateway("PUT", endpoint, data); err != nil {
				assert.Fatalf("%v\n%v\n", string(body), err)
				return
			}
			if resp.Status != 200 {
				assert.Fatalf("%v\n", string(body))
				fmt.Printf("Update node %v failed\n%v\n", i.HostIp, string(body))
				return
			}
			if err := json.Unmarshal(resp.Data, &node); err != nil {
				assert.Fatalf("%v\n%v\n", string(resp.Data), err)
				return
			}
			nodesToCheck = append(nodesToCheck, NodebyHostIP[i.HostIp])

		} else {
			fmt.Printf("MAAS already installed in nodeId:%v\n", addReq.Id)
		}
	}

	//Check MAAS installation
	for i := 0; i < len(nodesToCheck); i++ {
		fmt.Printf("Checking MAAS installation for nodeId:%v\n", nodesToCheck[i])
		check, err = checkMAASInstallation(nodesToCheck[i], from)
		if err != nil {
			fmt.Printf("%v", err)
		}
		if check {
			fmt.Printf("MAAS correctly installed on nodeId:%v\n", nodesToCheck[i])
		}
	}
}
