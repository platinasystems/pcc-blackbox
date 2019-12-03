package main

import (
	"encoding/json"
	"fmt"
	"github.com/platinasystems/test"
	//"github.com/platinasystems/tiles/pccserver/models"
	"testing"
)

type simpleKClusterRequest struct {
	ID         uint64
	Name       string         `json:"name" validate:"required"`
	K8sVersion string         `json:"k8sVersion" validate:"required"`
	CniPlugin  string         `json:"cniPlugin" validate:"required"`
	Nodes      []simpleKNodes `json:"nodes"`
}

type simpleKNodes struct {
	ID uint64
}

func createK8sCluster(t *testing.T) {
	t.Run("CreateK8sCluster", createK8s_3nodes)
}

func createK8s_3nodes(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	const DIM = 3
	var (
		body              []byte
		data              []byte
		resp              HttpResp
		err               error
		sk8sRequest       simpleKClusterRequest
		sNodes                 = make([]simpleKNodes, DIM)
		nodesSetCompleted bool = false
	)
	var j = 0
	for _, i := range Env.Invaders {
		sNodes[j] = simpleKNodes{ID: NodebyHostIP[i.HostIp]}
		j++
		if j == DIM {
			nodesSetCompleted = true
			continue
		}
	}
	for _, i := range Env.Servers {
		if nodesSetCompleted {
			continue
		}
		sNodes[j] = simpleKNodes{ID: NodebyHostIP[i.HostIp]}
		j++
		if j == DIM {
			nodesSetCompleted = true
			continue
		}
	}
	sk8sRequest = simpleKClusterRequest{
		ID:         0,         //todo dynamic counter
		Name:       "k8stest", //todo dynamic
		K8sVersion: "v1.14.3", //todo dynamic
		CniPlugin:  "kube-router",
		Nodes:      sNodes,
	}
	endpoint := fmt.Sprintf("pccserver/kubernetes")
	if data, err = json.Marshal(sk8sRequest); err != nil {
		assert.Fatalf("invalid struct for K8s creation")
	}
	if resp, body, err = pccGateway("POST", endpoint, data); err != nil {
		assert.Fatalf("%v\n%v\n", string(body), err)
		return
	}
	if resp.Status != 200 {
		assert.Fatalf("%v\n", string(body))
		fmt.Printf("K8s cretion failed:\n%v\n", string(body))
		return
	}
}

//todo check k8s status
