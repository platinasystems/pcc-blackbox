package main

import (
	"errors"
	"fmt"
	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/pcc-blackbox/models"
	"testing"
	"time"
)

func IsInvader(node *pcc.NodeDetailed) bool {
	for i := 0; i < len(Env.Invaders); i++ {
		if Env.Invaders[i].HostIp == node.Host {
			return true
		}
	}
	return false
}

func ConvertToMillis(startTime time.Time) uint64 {
	return uint64(startTime.UnixNano()) / uint64(time.Millisecond)
}

func idInSlice(idToFind uint64, list []uint64) bool {
	for _, str := range list {
		if str == idToFind {
			return true
		}
	}
	return false
}

func getNodeFromEnv(id uint64) *node {
	for i := range Env.Invaders {
		node := Env.Invaders[i].node
		if node.Id == id {
			return &node
		}
	}

	for i := range Env.Servers {
		node := Env.Servers[i].node
		if node.Id == id {
			return &node
		}
	}

	return nil
}

func CheckDependencies(t *testing.T, res *models.TestResult, dep ...func() error) {
	for _, fn := range dep {
		if err := fn(); err != nil {
			log.AuctaLogger.Errorf("%s", err.Error())
			res.SetTestSkipped(err.Error())
			t.SkipNow()
		}
	}
}

func CheckInvaders(nodes *[]pcc.NodeDetailed) (err error) {
	for _, node := range Env.Invaders {
		ok := false
		for i := 0; i < len(*nodes); i++ {
			remoteNode := (*nodes)[i]
			if node.HostIp == remoteNode.Host && remoteNode.Status == "OK" && remoteNode.NodeStatus.ConnectionStatus == "online" {
				ok = true
				break
			}
		}
		if !ok {
			msg := fmt.Sprintf("%s not found in PCC or not online", node.HostIp)
			err = errors.New(msg)
			return
		}
	}
	return
}

func CountInvadersMatching(nodes *[]pcc.NodeDetailed) (numInvaders int) {
	for _, node := range Env.Invaders {
		for i := 0; i < len(*nodes); i++ {
			remoteNode := (*nodes)[i]
			if node.HostIp == remoteNode.Host && remoteNode.Status == "OK" && remoteNode.NodeStatus.ConnectionStatus == "online" {
				numInvaders++
			}
		}
	}
	return
}

func CheckServers(nodes *[]pcc.NodeDetailed) (err error) {
	for _, node := range Env.Servers {
		ok := false
		for i := 0; i < len(*nodes); i++ {
			remoteNode := (*nodes)[i]
			if node.HostIp == remoteNode.Host && remoteNode.Status == "OK" && remoteNode.NodeStatus.ConnectionStatus == "online" {
				ok = true
				break
			}
		}
		if !ok {
			msg := fmt.Sprintf("%s not found in PCC or not online", node.HostIp)
			err = errors.New(msg)
			return
		}
	}
	return
}

func CountServersMatching(nodes *[]pcc.NodeDetailed) (numServers int) {
	for _, node := range Env.Servers {
		for i := 0; i < len(*nodes); i++ {
			remoteNode := (*nodes)[i]
			if node.HostIp == remoteNode.Host && remoteNode.Status == "OK" && remoteNode.NodeStatus.ConnectionStatus == "online" {
				numServers++
			}
		}
	}
	return
}

func CheckNodes() (err error) {
	var nodes *[]pcc.NodeDetailed
	if nodes, err = Pcc.GetNodes(); err == nil {
		if err = CheckInvaders(nodes); err != nil {
			return
		}
		if err = CheckServers(nodes); err != nil {
			return
		}
	} else {
		msg := fmt.Sprintf("Error getting nodes: %v", err)
		err = errors.New(msg)
		return
	}
	return
}

func CheckNumNodes(numNodes int) (err error) {
	var nodes *[]pcc.NodeDetailed
	if nodes, err = Pcc.GetNodes(); err == nil {
		if remoteNumNodes := len(*nodes); remoteNumNodes < numNodes {
			msg := fmt.Sprintf("%d nodes present,but %d nodes required", remoteNumNodes, numNodes)
			err = errors.New(msg)
			return
		}
		numInvaders := CountInvadersMatching(nodes)
		numServers := CountServersMatching(nodes)
		if (numServers + numInvaders) < numNodes {
			msg := fmt.Sprintf("%d matching nodes ,but %d nodes required", numServers+numInvaders, numNodes)
			err = errors.New(msg)
			return
		}

	} else {
		msg := fmt.Sprintf("Error getting nodes: %v", err)
		err = errors.New(msg)
		return
	}
	return
}

func CheckNodesNetCluster() (err error) {
	var envCluster netCluster
	for _, cluster := range Env.NetCluster {
		if cluster.Name == netClusterName {
			envCluster = cluster
			break
		}
	}
	if envCluster.Name == "" {
		err = errors.New("Can't find a Network Cluster with the provided ClusterName in the env file")
		return
	}
	for _, envNode := range envCluster.Nodes {
		nodeId, ok := Pcc.FindNodeAddress(envNode.IpAddr)
		if ok != nil {
			err = ok
			return
		}
		node, ok := Pcc.GetNode(nodeId)
		if ok != nil {
			err = ok
			return
		}
		if node.NodeStatus.ConnectionStatus != "online" {
			msg := fmt.Sprintf("%s connection status not online", node.Host)
			err = errors.New(msg)
			return
		}
	}
	return
}

func CheckCephClusterExists() (err error) {
	cephCluster, ok := Pcc.GetCephCluster(Env.CephConfiguration.ClusterName)
	if ok != nil {
		err = errors.New("Can't find a CephCluster with the provided ClusterName")
		return
	}
	if cephCluster.CephClusterConfig.ClusterNetwork != Env.CephConfiguration.ClusterNetwork ||
		cephCluster.CephClusterConfig.PublicNetwork != Env.CephConfiguration.PublicNetwork ||
		len(cephCluster.Nodes) != Env.CephConfiguration.NumberOfNodes {
		err = errors.New("The CephCluster does not match the specified parameters")
		return
	}

	if status, _ := Pcc.GetCephHealthStatusById(cephCluster.Id); status.Health == "HEALTH_ERR" {
		err = errors.New("The CephCluster status is not OK")
		return
	}
	return
}

func CheckK8sClusterExists() (err error) {
	k8sCluster, ok := Pcc.GetKubernetesClusterByName(k8sname)
	if ok != nil {
		err = errors.New("Can't find a k8sCluster with the provided ClusterName")
		return
	}
	if k8sCluster.HealthStatus != "good" {
		err = errors.New("The K8sCluster status is not OK")
		return
	}
	return
}

func CheckNetClusterExists() (err error) {
	networkCluster, ok := Pcc.FindNetClusterName(netClusterName)
	if ok != nil {
		err = errors.New("Can't find a Network Cluster with the provided ClusterName")
		return
	}
	var envCluster netCluster
	for _, cluster := range Env.NetCluster {
		if cluster.Name == netClusterName {
			envCluster = cluster
			break
		}
	}
	if envCluster.Name == "" {
		err = errors.New("Can't find a Network Cluster with the provided ClusterName in the env file")
		return
	}
	controlSubnet, ok := Pcc.FindSubnetObj(envCluster.ControlCIDR)
	if ok != nil {
		err = ok
		return
	}
	dataSubnet, ok := Pcc.FindSubnetObj(envCluster.DataCIDR)
	if ok != nil {
		err = ok
		return
	}
	if string(controlSubnet.Subnet) != networkCluster.ControlCIDR ||
		string(dataSubnet.Subnet) != networkCluster.DataCIDR ||
		envCluster.IgwPolicy != networkCluster.IgwPolicy {
		err = errors.New("The network Cluster does not match the specified parameters")
		return
	}
	for _, envNode := range envCluster.Nodes {
		found := false
		for _, node := range networkCluster.Nodes {
			remoteNode, ok := Pcc.GetNode(node.NodeId)
			if ok != nil {
				err = ok
				return
			}
			if envNode.IpAddr == remoteNode.Host {
				found = true
				break
			}
		}
		if !found {
			msg := fmt.Sprintf("Node %s not found in the Network %s", envNode.IpAddr, netClusterName)
			err = errors.New(msg)
			return
		}
	}
	if networkCluster.Health != "OK" {
		err = errors.New("The Network Cluster status is not OK")
		return
	}
	return
}
