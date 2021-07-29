package main

import (
	"fmt"
	"testing"
	"time"

	log "github.com/platinasystems/go-common/logs"
	"github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/tiles/pccserver/executor"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

var k8sname string = "k8stest"

func createK8sCluster(t *testing.T) {
	t.Run("CreateK8sCluster", createK8s_3nodes)
	t.Run("ValidateK8sCluster", validateK8sCluster)
	if t.Failed() {
		return
	}
	t.Run("AddNodeK8sCluster", addNodeK8sCluster)
	t.Run("ValidateK8sCluster", validateK8sCluster)
}

var k8sAddDelNodes []pcc.K8sNodes

func createK8s_3nodes(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res, CheckNetClusterExists)

	assert := test.Assert{t}

	const DIM = 3
	var (
		err               error
		k8sRequest        pcc.K8sClusterRequest
		k8sNodes               = make([]pcc.K8sNodes, DIM)
		nodesSetCompleted bool = false
	)

	if (len(Env.Invaders) + len(Env.Servers)) < DIM {
		msg := fmt.Sprintf("Needed at least %d nodes for creating a cluster", DIM)
		res.SetTestSkipped(msg)
		assert.SkipNow()
	}

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
			extraNode := pcc.K8sNodes{
				ID: NodebyHostIP[i.HostIp],
			}
			k8sAddDelNodes = append(k8sAddDelNodes, extraNode)
			break
		}
		k8sNodes[j] = pcc.K8sNodes{ID: NodebyHostIP[i.HostIp]}
		j++
		if j == DIM {
			nodesSetCompleted = true
			continue
		}
	}

	netClusterId, err := Pcc.FindNetClusterId(netClusterName)
	if err != nil {
		msg := fmt.Sprintf("FindNetClusterId failed: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	k8sRequest = pcc.K8sClusterRequest{
		ID:               0,         //todo dynamic counter
		Name:             k8sname,   //todo dynamic
		K8sVersion:       "v1.14.3", //todo dynamic
		CniPlugin:        "kube-router",
		Nodes:            k8sNodes,
		NetworkClusterId: netClusterId,
	}
	err = Pcc.CreateKubernetes(k8sRequest)
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}

func validateK8sCluster(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	assert := test.Assert{t}

	var (
		id  uint64
		err error
	)

	id, err = Pcc.FindKubernetesId(k8sname)
	if err != nil {
		msg := fmt.Sprintf("Failed to find cluster %v: %v", k8sname, err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	timeout := time.After(45 * time.Minute)
	tick := time.Tick(1 * time.Minute)
	done := false
	var last_percent int8 = -1
	for !done {
		select {
		case <-timeout:
			msg := "Timed out waiting for Kubernetes"
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		case <-tick:
			status, percent, err := Pcc.GetKubernetesDeployStatus(id)
			if err != nil {
				msg := fmt.Sprintf("Failed to get deploy status "+
					"%v", err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
			}

			switch status {
			case pcc.K8S_DEPLOY_STATUS_PROGRESS:
				fallthrough
			case pcc.K8S_DEPLOY_STATUS_COMPLETED:
				// When adding a new k8s cluster the deploy
				// status goes to this state when complete.
				// However on update k8s cluster it's
				// already at this state and only the percent
				// changes.
				if percent != last_percent {
					log.AuctaLogger.Infof("Cluster %v = %v  %v%%",
						id, status, percent)
					last_percent = percent
				}
				if percent == 100 {
					log.AuctaLogger.Info("Kubernetes cluster done")
					done = true
				}
			case pcc.K8S_DEPLOY_STATUS_FAILED:
				msg := "Kubernetes cluster install failed"
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
			case pcc.K8S_DEPLOY_APP_STATUS_PROGRESS:
				log.AuctaLogger.Info("Kubernetes app progress")
			default:
				msg := fmt.Sprintf("Unexpected status - %v",
					status)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
			}
		}
	}

	timeout = time.After(5 * time.Minute)
	tick = time.Tick(5 * time.Second)
	done = false
	for !done {
		select {
		case <-timeout:
			msg := "health check timed out"
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		case <-tick:
			health, err := Pcc.GetKubernetesHealth(id)
			if err != nil {
				msg := "Error geting K8s health"
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
			}
			log.AuctaLogger.Infof("Kubernetes health = %v", health)
			if health == "good" {
				done = true
				return
			}

		}
	}
}

func addNodeK8sCluster(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res, CheckK8sClusterExists)

	assert := test.Assert{t}

	var (
		id        uint64
		err       error
		updateReq pcc.KClusterUpdateRequest
		delNodes  []pcc.K8sNodes
	)

	id, err = Pcc.FindKubernetesId(k8sname)
	if err != nil {
		msg := fmt.Sprintf("Failed to find cluster %v: %v", k8sname, err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	if len(k8sAddDelNodes) == 0 {
		msg := fmt.Sprintf("No spare nodes to add to Kubernetes cluster %v", id)
		res.SetTestSkipped(msg)
		t.SkipNow()
	}

	updateReq = pcc.KClusterUpdateRequest{
		RolePolicy: "auto",
		ToAdd:      k8sAddDelNodes,
		ToRemove:   delNodes,
	}
	err = Pcc.UpdateKubernetes(id, updateReq)
	if err != nil {
		msg := fmt.Sprintf("Failed to update cluster %v: %v", k8sname, err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}

func deleteK8sCluster(t *testing.T) {
	t.Run("DeleteK8sCluster", deleteAllK8sCluster)
}

func deleteAllK8sCluster(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	assert := test.Assert{t}

	clusters, err := Pcc.GetKubernetes()
	if err != nil {
		msg := fmt.Sprintf("Failed to get kubernetes clusters: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	for _, c := range clusters {
		err := Pcc.DeleteKubernetes(c.ID, false)
		if err != nil {
			log.AuctaLogger.Warnf("delete K8s cluster failed, try force: %v",
				err)
			err := Pcc.DeleteKubernetes(c.ID, true)
			if err != nil {
				msg := fmt.Sprintf("force delete failed: %v", err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
			}
		}

		timeout := time.After(10 * time.Minute)
		tick := time.Tick(5 * time.Second)
		var last_percent int8 = -1
		for {
			select {
			case <-timeout:
				msg := "Time out deleting Kubernetes"
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
			case <-tick:
				var cluster pcc.K8sCluster
				allClusters, err := Pcc.GetKubernetes()
				if err != nil {
					msg := fmt.Sprintf("get cluster failed: %v", err)
					res.SetTestFailure(msg)
					log.AuctaLogger.Error(msg)
					assert.FailNow()
				}
				found := false
				for _, cluster = range allClusters {
					if cluster.ID == c.ID {
						found = true
						break
					}
				}
				if !found {
					log.AuctaLogger.Info("K8s delete OK")
					return
				}
				var percent int8
				task, ok := (cluster.Task).(*executor.Task)
				if ok {
					percent = task.Progress
				}
				if percent != last_percent {
					log.AuctaLogger.Infof("delete status: %v  %v%%",
						cluster.DeployStatus, percent)
					last_percent = percent
				}
			}
		}
	}
}
