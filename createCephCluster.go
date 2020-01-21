package main

import (
	"fmt"
	"github.com/platinasystems/test"
	"testing"
	"github.com/lib/pq"
	"time"
	"os"
	"github.com/platinasystems/pcc-blackbox/lib"
	"strings"
)

func createCephCluster(t *testing.T) {
	t.Run("create3NodesCephCluster", create3NodesCephCluster)
}

//TODO: Move this function to lib/ceph.go
func getCephClusterName() string{
	var serv0 string
	if len(Env.Invaders) > 0 {
		serv0 = Env.Invaders[0].HostIp
	} else if len(Env.Servers) > 0{
		serv0 = Env.Servers[0].HostIp
	}
	return fmt.Sprintf("%s_%s", pcc.CEPH_CLUSTER_NAME, serv0)
}

func create3NodesCephCluster(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	const DIM = 3
	var (
		err               error
		clusterId	  int
	)
	sNodes := getNodesList(DIM)

	createRequest := pcc.CreateCephClusterRequest{
		Name:       pcc.CEPH_CLUSTER_NAME, //todo dynamic
		Nodes:      sNodes,
		Tags: pq.StringArray{"ROTATIONAL","SATA","PCIe"},
	}
	createRequest.PublicNetwork = os.Getenv("CEPH_PUBLIC_NETWORK")
	createRequest.ClusterNetwork = os.Getenv("CEPH_CLUSTER_NETWORK")
	if createRequest.ClusterNetwork == "" || createRequest.PublicNetwork == "" {
		assert.Fatalf("Invalid Public or Cluster Network")
		return
	}
	deleteCluster := func(id int) (err error){
		time.Sleep(time.Second * 10)
		fmt.Println("Ceph cluster uninstallation is starting")
		from := time.Now()
		err = Pcc.DeleteCephCluster(uint64(id))
		if err != nil {
			err = fmt.Errorf("Ceph cluster deletion failed..ERROR: %v", err)
		}
		fmt.Println("Ceph cluster uninstallation has started. Cluster id:", id)
		fmt.Printf("Verifying ceph uninstallation...Timeout:[%v sec]\n", pcc.CEPH_3_NODE_UNINSTALLATION_TIMEOUT)
		err = verifyCephInstallation(from, "uninstall")
		if err != nil {
			err = fmt.Errorf("Ceph cluster undeployment failed ERROR: %v", err)
		}
		return
	}
	//TODO: Delete existing ceph cluster with same name if any

	fmt.Println("Ceph cluster installation is starting")
	from := time.Now()
	clusterId, err = Pcc.CreateCephCluster(createRequest)
	if err != nil {
		fmt.Println("Ceoh cluster deployment failed..ERROR:", err)
		if clusterId != -1 {
			err = deleteCluster(clusterId)
			if err != nil {
				fmt.Printf("%v\n", err)
			}
		}
		assert.Fatalf("Ceph cluster creation failed.")
		return
	}
	fmt.Println("Ceph cluster installation has started. Cluster id:", clusterId)
	fmt.Printf("Verifying ceph installation...Timeout:[%v sec]\n", pcc.CEPH_3_NODE_INSTALLATION_TIMEOUT)
	err = verifyCephInstallation(from, "install")
	if err != nil {
		fmt.Println("Ceoh cluster deployment failed..ERROR:", err)
		if clusterId != -1 {
			err = deleteCluster(clusterId)
			if err != nil {
				fmt.Printf("%v\n", err)
			}
		}
		assert.Fatalf("Ceph cluster deployment failed")
		return
	}

	//TODO: Create/Delete CephPool
	//TODO: Create/Delete CephFS

	err = deleteCluster(clusterId)
	if err != nil {
		assert.Fatalf("%v", err)
		return
	}
}

func verifyCephInstallation(from time.Time, action string) (err error){
	done := make(chan status)
	breakLoop := make([]chan bool, 2)

	if action == "install" {
		go func() {
			breakLoop[0] = make(chan bool)
			syncCheckGenericInstallation(0, pcc.CEPH_3_NODE_INSTALLATION_TIMEOUT, pcc.CEPH_INSTALLATION_SUCCESS_NOTIFICATION, from, done, breakLoop[0])
		}()
		go func() {
			breakLoop[1] = make(chan bool)
			syncCheckGenericInstallation(0, pcc.CEPH_3_NODE_INSTALLATION_TIMEOUT, pcc.CEPH_INSTALLATION_FAILED_NOTIFICATION, from, done, breakLoop[1])
		}()
	} else if action == "uninstall" {
		go func() {
			breakLoop[0] = make(chan bool)
			syncCheckGenericInstallation(0, pcc.CEPH_3_NODE_UNINSTALLATION_TIMEOUT, pcc.CEPH_UNINSTALLATION_SUCCESS_NOTIFICATION, from, done, breakLoop[0])
		}()
		go func() {
			breakLoop[1] = make(chan bool)
			syncCheckGenericInstallation(0, pcc.CEPH_3_NODE_UNINSTALLATION_TIMEOUT, pcc.CEPH_UNINSTALLATION_FAILED_NOTIFICATION, from, done, breakLoop[1])
		}()
	}
	s := <- done
	go func() {
		for i := 0; i < 2; i++ {
			breakLoop[i] <- true
		}
	}()
	if !s.isError {
		if strings.Contains(s.msg, pcc.CEPH_INSTALLATION_SUCCESS_NOTIFICATION) {
			fmt.Println("Ceph cluster deployed properly..\n", s.msg)
		} else if strings.Contains(s.msg, pcc.CEPH_INSTALLATION_FAILED_NOTIFICATION) {
			err = fmt.Errorf("%v", s.msg)
		} else if strings.Contains(s.msg, pcc.CEPH_UNINSTALLATION_SUCCESS_NOTIFICATION) {
			fmt.Println("Ceph cluster undeployed properly..\n", s.msg)
		} else if strings.Contains(s.msg, pcc.CEPH_UNINSTALLATION_FAILED_NOTIFICATION) {
			err = fmt.Errorf("%v", s.msg)
		}
	} else {
		err = fmt.Errorf("%v", s.msg)
	}
	return
}

/*
// Another way to verify ceph installation
func verifyCephInstallation(from time.Time) (err error) {
	var (
		cluster *models.CephCluster
		timeoutDuration = time.Duration(time.Minute * 10)
	)
	loop:
	for {
		cluster, err = Pcc.GetCephCluster(pcc.CEPH_CLUSTER_NAME)
		if err == nil {
			fmt.Println("Cluster deploy status:", cluster.DeployStatus)
			if cluster.DeployStatus == ceph.CEPH_DEPLOY_STATUS_COMPLTETED {
				break loop
			} else if cluster.DeployStatus == ceph.CEPH_DEPLOY_STATUS_INSTALLING && time.Now().Sub(from) > timeoutDuration {
				err = fmt.Errorf("Cluster verification failed. timeout (%v min) exceeded", timeoutDuration.Minutes())
				break loop
			} else if cluster.DeployStatus == ceph.CEPH_DEPLOY_STATUS_FAILED {
				err = fmt.Errorf("Cluster verification failed")
				break loop
			}
		}
		time.Sleep(time.Second * 5)
	}
	return
}
*/

func getNodesList(num int) []pcc.CephNodes{
	var(
		DIM = num
		sNodes = make([]pcc.CephNodes, DIM)
		nodesSetCompleted bool = false
	)
	var j = 0
	for _, i := range Env.Invaders {
		sNodes[j] = pcc.CephNodes{ID: NodebyHostIP[i.HostIp]}
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
		sNodes[j] = pcc.CephNodes{ID: NodebyHostIP[i.HostIp]}
		j++
		if j == DIM {
			nodesSetCompleted = true
			continue
		}
	}
	return sNodes
}
