package main

import (
	"encoding/json"
	"fmt"
	"github.com/platinasystems/tiles/pccserver/models"
)

func getIfacesByNodeId(nodeId uint64) (ifaces []*models.InterfaceDetail, err error) {
	var (
		body     []byte
		resp     HttpResp
		endpoint string
		node     models.NodeDetail
	)
	endpoint = fmt.Sprintf("pccserver/node/%v", nodeId)
	resp, body, err = pccGateway("GET", endpoint, nil)
	if err != nil {
		fmt.Printf("%v\n%v\n", string(body), err)
		return nil, err
	}
	if resp.Status == 200 {
		err := json.Unmarshal(resp.Data, &node)
		if err != nil {
			fmt.Printf("%v\n%v\n", string(body), err)
			return nil, err
		}
		ifaces = node.Interfaces
		return ifaces, nil
	}
	return nil, err
}

func getIfaceById(nodeId uint64, ifaceId int64) (iface *models.InterfaceDetail, err error) {
	ifaces, err := getIfacesByNodeId(nodeId)
	if err != nil {
		return
	}
	for _, i := range ifaces {
		if i.Interface.Id == ifaceId {
			iface = i
			return
		}
	}
	err = fmt.Errorf("error getting interface %v on node %v",
		ifaceId, nodeId)
	return
}

func getIfaceByMacAddress(mAddr string, ifaces []*models.InterfaceDetail) (iface *models.InterfaceDetail, err error) {
	for i := 0; i < len(ifaces); i++ {
		if ifaces[i].Interface.MacAddress == mAddr {
			return ifaces[i], nil
		}
	}
	return nil, err
}

func setIface(iface models.InterfaceRequest) (err error) {
	var (
		body     []byte
		data     []byte
		resp     HttpResp
		endpoint string
	)
	endpoint = fmt.Sprintf("pccserver/interface")
	fmt.Printf("Sending InterfaceRequest [%v] to pccserver/interface\n", iface)
	data, err = json.Marshal(iface)
	if err != nil {
		return fmt.Errorf("Iface format not valid")
	}
	resp, body, err = pccGateway("POST", endpoint, data)
	if err != nil {
		fmt.Printf("%v\n%v\n", string(body), err)
		return err
	}
	if resp.Status == 200 {
		return nil
	}
	endpoint = fmt.Sprintf("pccserver/interface/apply")
	fmt.Printf("Sending confirmation for node [%v]\n", iface.NodeId)
	var jsonData = map[string]string{"nodeId": string(iface.NodeId)}
	data, err = json.Marshal(jsonData)
	if err != nil {
		return fmt.Errorf("Format not valid")
	}
	resp, body, err = pccGateway("POST", endpoint, data)
	if err != nil {
		fmt.Printf("%v\n%v\n", string(body), err)
		return err
	}
	if resp.Status == 200 {
		return nil
	}
	return fmt.Errorf(resp.Message)
}
