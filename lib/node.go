// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"fmt"
	"sync"
	"time"

	"github.com/platinasystems/pcc-models/avro"
	"github.com/platinasystems/tiles/pccserver/models"
)

type NodeAvailability struct {
	avro.NodeAvailability
}

type Node struct {
	models.Node
}

type NodeDetailed struct {
	models.NodeDetailed
}

// fixme
type NodeDetail struct {
	//models.NodeDetail
	NodeDetailed
	//HardwareInventory *HardwareInventory
	//Apps              []ProvisionedApp
	//Roles             []Role             `gorm:"many2many:node_roles"`
	// SystemData        *avro.Status       `json:"systemData"`
	Interfaces []*InterfaceDetail `json:"interfaces"`
}

func (p *PccClient) GetNodesDetail() (nodes []*NodeDetailed, err error) {
	err = p.Get("pccserver/node", &nodes)
	return
}

func (p *PccClient) GetNodesId(id uint64) (node NodeDetailed, err error) {
	endpoint := fmt.Sprintf("pccserver/node/%v", id)
	err = p.Get(endpoint, &node)
	return
}

func (p *PccClient) FindNodeId(name string) (id uint64, err error) {
	var nodes *[]NodeDetailed

	if nodes, err = p.GetNodes(); err == nil {
		for _, n := range *nodes {
			if n.Name == name {
				id = n.Id
				return
			}
		}
	}
	err = fmt.Errorf("node [%v] not found", name)
	return
}

func (p *PccClient) FindNodeAddress(addr string) (id uint64, err error) {
	var nodes *[]NodeDetailed

	if nodes, err = p.GetNodes(); err == nil {
		for _, n := range *nodes {
			if n.Host == addr {
				id = n.Id
				return
			}
		}
	}
	err = fmt.Errorf("node [%v] not found", addr)
	return
}

func (p *PccClient) GetProvisionStatus(id uint64) (status string, err error) {
	var node *NodeDetailed
	if node, err = p.GetNode(id); err == nil {
		status = node.ProvisionStatus
	}
	return
}

func (p *PccClient) GetNodeSummary(id uint64, node *Node) (err error) {
	endpoint := fmt.Sprintf("pccserver/node/summary/%v", id)
	err = p.Get(endpoint, node)
	return
}

func (p *PccClient) GetNodeConnectionStatus(nodeId uint64) (status string, err error) {
	var node *NodeDetailed
	if node, err = p.GetNode(nodeId); err == nil {
		if node.NodeStatus != nil {
			status = node.NodeStatus.ConnectionStatus
		}
	}
	return
}

func (p *PccClient) IsNodeOnline(nodeId uint64) bool {
	if status, err := p.GetNodeConnectionStatus(nodeId); err == nil {
		return status == "online"
	} else {
		return false
	}
}

func (pcc *PccClient) GetNodes() (nodes *[]NodeDetailed, err error) {
	var n []NodeDetailed
	if err = pcc.Get("pccserver/node", &n); err == nil {
		nodes = &n
	}
	return
}

func (pcc *PccClient) GetNodesOptions(page int, limit int, sortBy string, sortDir string, search string) (nodes *[]NodeDetailed, err error) {
	var n []NodeDetailed
	endpoint := fmt.Sprintf("pccserver/node?page=%d&limit=%d&sortBy=%s&sortDir=%s&search=%s",
		page, limit, sortBy, sortDir, search)
	if err = pcc.Get(endpoint, &n); err == nil {
		nodes = &n
	}
	return
}

func (pcc *PccClient) GetNode(id uint64) (node *NodeDetailed, err error) {
	var n NodeDetailed
	if err = pcc.Get(fmt.Sprintf("pccserver/node/%d", id), &n); err == nil {
		node = &n
	}
	return
}

// List of all invaders
func (pcc *PccClient) GetInvaders() (nodes *[]NodeDetailed, err error) {
	return pcc.filterNodes(true)
}

// List of all invaders
func (pcc *PccClient) GetNodeIds() (nodes []uint64, err error) {
	var n *[]NodeDetailed
	if n, err = pcc.GetNodes(); err == nil {
		for i := range *n {
			nodes = append(nodes, (*n)[i].Id)
		}
	}

	return
}

// List of all invaders
func (pcc *PccClient) GetInvaderIds() (nodes []uint64, err error) {
	var n *[]NodeDetailed
	if n, err = pcc.GetInvaders(); err == nil {
		for i := range *n {
			nodes = append(nodes, (*n)[i].Id)
		}
	}

	return
}

// List of all servers
func (pcc *PccClient) GetServers() (nodes *[]NodeDetailed, err error) {
	return pcc.filterNodes(false)
}

func (pcc *PccClient) filterNodes(invader bool) (nodes *[]NodeDetailed, err error) {
	if nodes, err = pcc.GetNodes(); err == nil {
		var inv []NodeDetailed

		for i := range *nodes {
			if (*nodes)[i].Invader == invader {
				inv = append(inv, (*nodes)[i])
			}
		}

		nodes = &inv
	}
	return
}

func (pcc *PccClient) GetEnvironment(nodeID *uint64) (env map[string]interface{}, err error) {

	env = make(map[string]interface{})
	nodeStr := ""
	if nodeID != nil && *nodeID > 0 {
		nodeStr = fmt.Sprintf("%d", *nodeID)
	}
	err = pcc.Get(fmt.Sprintf("pccserver/environment/%s", nodeStr), &env)
	return
}

// Fetch the node from the DB
func (pcc *PccClient) GetNodeFromDB(nodeId uint64) (node *NodeDetailed, err error) {
	var n NodeDetailed
	if err = pcc.DBHandler().FetchWhere(&n, "id = ?", nodeId); err == nil {
		node = &n
	} else {
		fmt.Println("Not able to fetch the node", err)
	}
	return
}

// Fetch the nodes from the DB
func (pcc *PccClient) GetNodesFromDB() (nodes *[]NodeDetailed, err error) {
	var n []NodeDetailed
	if err = pcc.DBHandler().FetchMany(&n); err == nil {
		nodes = &n
	} else {
		fmt.Println("Not able to fetch the nodes", err)
	}
	return
}

// Fetch the invaders from the DB
func (pcc *PccClient) GetInvadersFromDB() (nodes *[]NodeDetailed, err error) {
	var n []NodeDetailed
	if err = pcc.DBHandler().FetchManyWhere(&n, "invader = true"); err == nil {
		nodes = &n
	} else {
		fmt.Println("Not able to fetch the invader nodes", err)
	}
	return
}

// Fetch the servers from the DB
func (pcc *PccClient) GetServersFromDB() (nodes *[]NodeDetailed, err error) {
	var n []NodeDetailed
	if err = pcc.DBHandler().FetchManyWhere(&n, "invader = false"); err == nil {
		nodes = &n
	} else {
		fmt.Println("Not able to fetch the invader nodes", err)
	}
	return
}

// Delete all the nodes (servers and invaders). Returns the deleted node
func (pcc *PccClient) DeleteNodes(wait bool) (nodes *[]NodeDetailed, err error) {
	if nodes, err = pcc.GetNodes(); err == nil {
		err = pcc.deleteNodes(nodes, wait)
	}

	return
}

// Delete the invaders
func (pcc *PccClient) DeleteInvaders(wait bool) (err error) {
	var nodes *[]NodeDetailed
	if nodes, err = pcc.GetInvaders(); err == nil {
		err = pcc.deleteNodes(nodes, wait)
	}
	return
}

// Delete the servers
func (pcc *PccClient) DeleteServers(wait bool) (err error) {
	var nodes *[]NodeDetailed
	if nodes, err = pcc.GetServers(); err == nil {
		err = pcc.deleteNodes(nodes, wait)
	}

	return
}

// Delete all the nodes
func (pcc *PccClient) deleteNodes(nodes *[]NodeDetailed, wait bool) (err error) {
	var ids []uint64
	for i := range *nodes {
		node := (*nodes)[i]
		ids = append(ids, node.Id)
	}

	for _, nodeId := range ids {
		if err = pcc.DeleteNode(nodeId); err != nil {
			return
		}
	}

	if len(ids) > 0 && wait {
		var running []uint64
	sleep:
		for i := 1; i <= 20; i++ { // FIXME add the notification
			toDelete := ids
			time.Sleep(time.Second * time.Duration(15))
			if running, err = pcc.GetNodeIds(); err == nil {
				ids = make([]uint64, 0)

				for _, nodeId := range toDelete {
					for _, id := range running {
						if id == nodeId {
							ids = append(ids, nodeId) // node still running
							continue sleep
						}
					}
					fmt.Println(fmt.Sprintf("Node %d has been deleted", nodeId))
				}
				return
			}
		}

		err = fmt.Errorf("an error occurs deleting the servers")
	}

	return
}

// Add a node
func (pcc *PccClient) AddNode(node *NodeDetailed) (err error) {
	err = pcc.Post("pccserver/node/add", node, node)
	return
}

// Delete a node
func (pcc *PccClient) DeleteNode(id uint64) (err error) {
	fmt.Println(fmt.Sprintf("deleting the node %d", id))
	err = pcc.Delete(fmt.Sprintf("pccserver/node/%d", id), nil, nil)
	return
}

// Update a node
func (pcc *PccClient) UpdateNode(node *NodeDetailed) (err error) {
	fmt.Println(fmt.Sprintf("updating the node %d", node.Id))
	if len(node.RoleIds) > 0 { // FIXME fix pcc-side
		m := make(map[uint64]bool)
		for _, k := range node.RoleIds {
			m[k] = true
		}
		var roles []uint64
		for k := range m {
			roles = append(roles, k)
		}
		node.RoleIds = roles
	}

	err = pcc.Put("pccserver/node/update", node, node)
	return
}

// Update MaaS role
func (pcc *PccClient) UpdateMaas(node *NodeDetailed) (err error) {
	var timeout time.Duration = 1 * time.Minute

	endpoint := fmt.Sprintf("pccserver/node/updateMaas/%v", node.Id)
	err = pcc.Put(endpoint, node, node, timeout)
	return
}

// Install in parallel on nodes. Keep previous roles
// TODO wait for the installation by using the requestID
func (pcc *PccClient) AddRolesToNodes(nodes []uint64, roles []uint64) (installed []uint64, installing []uint64, err error) {
	var (
		lock sync.Mutex
		wg   sync.WaitGroup
	)
	wg.Add(len(nodes))

	addRoles := func(nodeId uint64) {
		defer wg.Done()
		var (
			errNode          error
			alreadyInstalled bool
			node             *NodeDetailed
		)
		if node, errNode = pcc.GetNode(nodeId); err == nil {
			node.RoleIds = append(node.RoleIds, roles...) // keep other roles
			if alreadyInstalled, errNode = pcc.AreRoleInstalled(nodeId, roles); errNode == nil {
				if alreadyInstalled {
					lock.Lock()
					installed = append(installed, node.Id)
					lock.Unlock()
					fmt.Printf("roles %v already set on node:%v\n", roles, nodeId)
				} else {
					fmt.Printf("setting roles %v on node %d\n", roles, nodeId)
					lock.Lock()
					installing = append(installing, nodeId)
					lock.Unlock()
					errNode = pcc.UpdateNode(node)
				}
			} else {
				err = errNode
			}
		} else {
			err = errNode
		}
	}

	for _, id := range nodes {
		go addRoles(id)
	}

	wg.Wait()
	return
}

// node group

type NodeGroup struct {
	models.Cluster
}

type NodeGroupRequest struct {
	Id          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   uint64 `json:"createAt"`
	ModifiedAt  uint64 `json:"modifiedAt"`
	Owner       uint64 `json:"owner"`
}

func (pcc *PccClient) GetNodeGroups() (groups []NodeGroup, err error) {
	err = pcc.Get("pccserver/cluster", &groups)
	return
}

func (pcc *PccClient) GetNodeGroupId(id uint64) (group NodeGroup, err error) {
	endpoint := fmt.Sprintf("pccserver/cluster/%v", id)
	err = pcc.Get(endpoint, &group)
	return
}

func (pcc *PccClient) AddNodeGroup(req *NodeGroupRequest) (err error) {
	err = pcc.Post("pccserver/cluster/add", req, req)
	return
}

func (pcc *PccClient) UpdateNodeGroup(group *NodeGroup) (err error) {
	endpoint := fmt.Sprintf("pccserver/cluster/%v", group.ID)
	err = pcc.Put(endpoint, group, group)
	return
}

func (pcc *PccClient) DeleteNodeGroup(id uint64) (err error) {
	endpoint := fmt.Sprintf("pccserver/cluster/%v", id)
	err = pcc.Delete(endpoint, nil, nil)
	return
}
