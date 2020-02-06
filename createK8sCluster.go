package main

import (
	"fmt"
	"testing"
	"time"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

var k8sname string = "k8stest"

func createK8sCluster(t *testing.T) {
	t.Run("CreateK8sCluster", createK8s_3nodes)
	t.Run("ValidateK8sCluster", validateK8sCluster)
}

func createK8s_3nodes(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	const DIM = 3
	var (
		err               error
		k8sRequest        pcc.K8sClusterRequest
		k8sNodes               = make([]pcc.K8sNodes, DIM)
		nodesSetCompleted bool = false
	)
	var j = 0
	for _, i := range Env.Invaders {
		k8sNodes[j] = pcc.K8sNodes{ID: NodebyHostIP[i.HostIp]}
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
		k8sNodes[j] = pcc.K8sNodes{ID: NodebyHostIP[i.HostIp]}
		j++
		if j == DIM {
			nodesSetCompleted = true
			continue
		}
	}
	k8sRequest = pcc.K8sClusterRequest{
		ID:         0,         //todo dynamic counter
		Name:       k8sname,   //todo dynamic
		K8sVersion: "v1.14.3", //todo dynamic
		CniPlugin:  "kube-router",
		Nodes:      k8sNodes,
	}
	err = Pcc.CreateKubernetes(k8sRequest)
	if err != nil {
		assert.Fatalf("%v", err)
	}
}

func validateK8sCluster(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	id, err := Pcc.FindKubernetesId(k8sname)
	if err != nil {
		assert.Fatalf("%v", err)
	}

	timeout := time.After(30 * time.Minute)
	tick := time.Tick(1 * time.Minute)
	done := false
	loops := 0
	for !done {
		loops++
		select {
		case <-timeout:
			assert.Fatalf("Timed out waiting for Kubernetes")
		case <-tick:
			status, err := Pcc.GetKubernetesDeployStatus(id)
			if err != nil {
				assert.Fatalf("Failed to get deploy status "+
					"%v\n", err)
			}
			switch status {
			case pcc.K8S_DEPLOY_STATUS_PROGRESS:
				if loops%5 == 0 {
					fmt.Printf("Cluster %v = %v\n",
						id, status)
				}
			case pcc.K8S_DEPLOY_STATUS_COMPLETED:
				fmt.Println("Kubernetes cluster installed")
				done = true
			case pcc.K8S_DEPLOY_STATUS_FAILED:
				assert.Fatalf("Kubernetes cluster install " +
					"failed")
			case pcc.K8S_DEPLOY_APP_STATUS_PROGRESS:
				fmt.Println("Kubernetes app progress")
			default:
				assert.Fatalf("Unexpected status - %v\n",
					status)
			}
		}
	}
	health, err := Pcc.GetKubernetesHealth(id)
	if err != nil {
		assert.Fatalf("Error geting K8s health\n")
	}
	fmt.Printf("Kubernetes health = %v\n", health)
}
