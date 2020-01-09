package main

import (
	"encoding/json"
	"fmt"
	"github.com/platinasystems/test"
	"github.com/platinasystems/tiles/pccserver/models"
	"testing"
	"time"
)

func updateNodes_installLLDP(t *testing.T) {
	t.Run("installLLDP", installLLDP)
}

func installLLDP(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		body  []byte
		resp  HttpResp
		err   error
		check bool
	)

	var isLLDPInNodes = make(map[uint64]bool)
	for id := range Nodes {
		var (
			addReq models.NodeWithAdditionalFields
			node   models.NodeWithKubernetes
			data   []byte
			lldp   []uint64 = []uint64{2}
		)

		addReq.Host = Nodes[id].Host
		addReq.Id = id
		addReq.RoleIds = lldp

		isLLDPInNodes[id] = IsAppInstalled(id, "lldpd")
		if !isLLDPInNodes[id] {

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
				fmt.Printf("Update node %v failed\n%v\n", Nodes[id].Host, string(body))
				return
			}
			if err := json.Unmarshal(resp.Data, &node); err != nil {
				assert.Fatalf("%v\n%v\n", string(resp.Data), err)
				return
			}
		} else {
			fmt.Printf("LLDP already installed in nodeId:%v\n", id)
		}
	}

	from := time.Now()
	//Check LLDP installation
	for id := range Nodes {
		if !isLLDPInNodes[id] {
			fmt.Printf("Checking LLDP installation for nodeId:%v\n", id)
			//check, err = checkLLDPInstallation(id)
			check, err = checkGenericInstallation(id, LLDP_TIMEOUT, LLDP_NOTIFICATION, from)
			if err != nil {
				fmt.Printf("%v", err)
			}
			if check {
				fmt.Printf("LLDP correctly installed on nodeId:%v\n", id)
			}
		}
	}
}
