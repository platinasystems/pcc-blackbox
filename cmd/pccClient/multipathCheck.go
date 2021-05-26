package main

import (
	"github.com/platinasystems/tiles/pccserver/models"
	"github.com/platinasystems/tiles/pccserver/network"
	"github.com/platinasystems/tiles/pccserver/utility"

	"encoding/json"
	"fmt"
)

type iprt2MetaBr struct {
	Detail rtMeta
}

type rtMeta struct {
	Main  string `json:"route table main"`
	Local string `json:"route table local"`
}

func multipathCheck(input interface{}) (pairs []*network.PktPaths, infos utility.Errors, err error) {
	var (
		netName  string
		netId    uint64
		out      interface{}
		clusters []models.NetworkClusterUI
		cluster  models.NetworkClusterUI
		n        network.NetworkClusterSw
		errs     utility.Errors
	)
	switch t := input.(type) {
	case int:
		netId = uint64(t)
	case uint64:
		netId = t
	case string:
		netName = t
	default:
		err = fmt.Errorf("Unrecognized type %T network ID or name %v", input, input)
		return
	}

	if netId == 0 && netName == "" {
		err = fmt.Errorf("Cannot retrieve network %v", input)
		return
	}

	// Get cluster
	if netId == 0 {
		// get by name
		ep := "pccserver/network/cluster"
		if err = Pcc.Get(ep, &out, nil); err != nil {
			return
		}
		cs := pSprint(out)
		err = json.Unmarshal([]byte(cs), &clusters)
		if err != nil {
			return
		}
		found := false
		for i, c := range clusters {
			if c.Name == netName {
				cluster = clusters[i]
				found = true
				break
			}
		}
		if !found {
			err = fmt.Errorf("Cannot find network with name %v", netName)
			return
		}
	} else {
		ep := fmt.Sprintf("pccserver/network/cluster/%v", netId)
		if err = Pcc.Get(ep, &out, nil); err != nil {
			return
		}
		cs := pSprint(out)
		err = json.Unmarshal([]byte(cs), &cluster)
		if err != nil {
			return
		}
	}

	// update toplogy and fibmap from API
	n.NetworkClusterUI = cluster
	n.Topologies = map[uint64]models.Topology{}
	for i, node := range n.Nodes {
		var t []models.Topology
		// get topology
		ep := fmt.Sprintf("pccserver/topology/%v", node.NodeId)
		if err = Pcc.Get(ep, &out); err == nil {
			tj := pSprint(out)
			if err2 := json.Unmarshal([]byte(tj), &t); err2 == nil && len(t) > 0 {
				n.Topologies[node.NodeId] = t[0]
				n.Nodes[i].NodeName = t[0].NodeName
			} else {
				errs.Append(err2)
			}
		}
		if err != nil {
			errs.Append(err)
		}
		// get fibmap
		ep = fmt.Sprintf("pccserver/v2/node/%v/apps", node.NodeId)
		if err = Pcc.Get(ep, &out); err == nil {
			var apps []map[string]interface{}
			aj := pSprint(out)
			if err = json.Unmarshal([]byte(aj), &apps); err != nil {
				errs.Append(err)
			} else {
				for _, a := range apps {
					var m iprt2MetaBr
					if a["AppName"] != models.AppIdIproute2 {
						continue
					}
					meta, ok := a["MetaData"]
					if !ok {
						continue
					}
					b, _ := json.Marshal(meta)
					if err = json.Unmarshal(b, &m); err == nil {
						errs.Append(n.FibMapMain.FromNode(node.NodeId, m.Detail.Main))
						errs.Append(n.FibMapLocal.FromNode(node.NodeId, m.Detail.Local))
					} else {
						errs.Append(err)
					}
				}
			}
		}
	}

	if errs.IsEmpty() {
		pairs, _, _ = n.CheckPaths()
	} else {
		err = fmt.Errorf("%v", errs)
	}

	return
}
