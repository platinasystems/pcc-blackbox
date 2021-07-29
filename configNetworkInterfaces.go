package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/platinasystems/pcc-blackbox/models"

	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

var nodeIntfMap = make(map[uint64][]int64)

func configServerInterfaces(t *testing.T) {
	t.Run("configNetworkInterfaces", configNetworkInterfaces)
	t.Run("verifyNetworkInterfaces", verifyNetworkInterfaces)
	if t.Failed() {
		return
	}
	t.Run("verifyNetworkConfig", verifyNetworkConfig)
	t.Run("verifyNetworkUp", verifyNetworkUp)
}

// FIXME lump the functions in an single one
func configNetworkInterfaces(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res, Env.CheckInvaders, Env.CheckServers, CheckNodes)

	assert := test.Assert{t}

	configureNode := func(node node, sever bool) {
		var (
			err       error
			ifaces    []*pcc.InterfaceDetail
			nodeIntfs []int64
		)

		id := NodebyHostIP[node.HostIp]
		ifaces, err = Pcc.GetIfacesByNodeId(id)
		if err != nil {
			msg := fmt.Sprintf("Error retrieving node %s id[%d] interfaces %v", node.HostIp, id, err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		}
	l2:
		for j := range node.NetInterfaces { // skip the check of the interfaces are not declared in Env
			iface := node.NetInterfaces[j]
			for _, intf := range ifaces {
				if iface.Name == intf.Interface.Name || iface.MacAddr == intf.Interface.MacAddress {
					nodeIntfs = append(nodeIntfs, intf.Interface.Id)
					continue l2
				}
			}

			log.AuctaLogger.Infof("SKIP interface %v for node %d", iface, id)
		}
		nodeIntfMap[id] = nodeIntfs
		configNodeInterfaces(t, sever, id, node.HostIp, node.NetInterfaces, ifaces)
	}

	for i := range Env.Invaders {
		configureNode(Env.Invaders[i].node, false)
	}
	for i := range Env.Servers {
		configureNode(Env.Servers[i].node, true)
	}

}

func prepIfaceRequest(nodeId uint64, iface *pcc.InterfaceDetail, configIface netInterface) (ifaceRequest pcc.InterfaceRequest) {

	var digitCheck = regexp.MustCompile(`^[0-9]+$`)

	ifaceRequest.NodeId = nodeId
	ifaceRequest.InterfaceId = iface.Interface.Id
	ifaceRequest.Name = iface.Interface.Name
	ifaceRequest.Ipv4Addresses = configIface.Cidrs
	ifaceRequest.MacAddress = configIface.MacAddr
	ifaceRequest.ManagedByPcc = configIface.ManagedByPcc
	if configIface.IsManagement {
		ifaceRequest.Gateway = configIface.Gateway
	}
	ifaceRequest.Autoneg = configIface.Autoneg
	switch configIface.Autoneg {
	case "true", "on", "":
		ifaceRequest.Autoneg = pcc.INTERFACE_AUTONEG_ON
	case "false", "off":
		ifaceRequest.Autoneg = pcc.INTERFACE_AUTONEG_OFF

		if digitCheck.MatchString(configIface.Speed) {
			ifaceRequest.Speed = json.Number(configIface.Speed)
		}
	default:
		log.AuctaLogger.Errorf("Error: invalid autoneg [%v] using ON",
			configIface.Autoneg)
		ifaceRequest.Autoneg = pcc.INTERFACE_AUTONEG_ON
	}
	ifaceRequest.Mtu = json.Number(configIface.Mtu)

	adminStatus := configIface.AdminStatus
	switch adminStatus {
	case pcc.INTERFACE_STATUS_UP:
	case pcc.INTERFACE_STATUS_DOWN:
	default:
		adminStatus = pcc.INTERFACE_STATUS_UP
	}
	ifaceRequest.AdminStatus = adminStatus

	if configIface.IsManagement {
		ifaceRequest.IsManagement = "true"
	} else {
		ifaceRequest.IsManagement = "false"
	}

	var fec string
	switch configIface.Fec {
	case "cl91", "fec91", "rs":
		fec = pcc.INTERFACE_FEC_CL91
	case "cl74", "fec74", "baser":
		fec = pcc.INTERFACE_FEC_CL74
	case "none", "", "off":
		fec = pcc.INTERFACE_FEC_NONE
	default:
		log.AuctaLogger.Errorf("Error: invalid fec [%v] using none",
			configIface.Fec)
		fec = pcc.INTERFACE_FEC_NONE
	}
	ifaceRequest.FecType = fec
	return
}

func configNodeInterfaces(t *testing.T, skipManagement bool, nodeId uint64, HostIp string,
	serverInterfaces []netInterface, ifaces []*pcc.InterfaceDetail) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	assert := test.Assert{t}
	var (
		iface        *pcc.InterfaceDetail
		ifaceRequest pcc.InterfaceRequest
		err          error
	)

	for j := 0; j < len(serverInterfaces); j++ {
		mac := serverInterfaces[j].MacAddr
		if serverInterfaces[j].IsManagement && !skipManagement {
			continue // don't mess with management
		}
		iface, err = Pcc.GetIfaceByMacAddress(mac, ifaces)
		if err != nil {
			msg := fmt.Sprintf("Error in retrieving interface having "+
				"MacAddress: %v for node %v id[%v]",
				mac, HostIp, nodeId)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		}

		ifaceRequest = prepIfaceRequest(nodeId, iface,
			serverInterfaces[j])

		log.AuctaLogger.Infof("Configuring node %v interface %v %v", nodeId,
			iface.Interface.Name, ifaceRequest)

		if err := Pcc.SetIface(ifaceRequest); err != nil {
			msg := fmt.Sprintf("Error setting interface %v for node "+
				"%v id[%v]: %v", ifaceRequest, HostIp,
				nodeId, err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		}
	}

	log.AuctaLogger.Infof("Apply interface changes for node %d", nodeId)
	if err = Pcc.ApplyIface(nodeId); err != nil {
		msg := fmt.Sprintf("Interface apply failed: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

}

func validateIfaceConfig(intfReq pcc.InterfaceRequest) (err error) {

	iface, err := Pcc.GetIfaceById(intfReq.NodeId, intfReq.InterfaceId)
	if err != nil {
		return
	}

	err = fmt.Errorf("config mismatch for %v", intfReq.Name)

	log.AuctaLogger.Infof("  Validating config for %v", intfReq.Name)

	// chop off ",<metric>"
	rGateway := strings.Split(intfReq.Gateway, ",")
	cGateway := strings.Split(iface.Interface.Gateway, ",")
	if rGateway[0] != cGateway[0] {
		if rGateway[0] == "" && cGateway[0] != "" {
			log.AuctaLogger.Info("    skipping configured gateway")
		} else {
			log.AuctaLogger.Errorf("    gateway mismatch [%v] [%v]",
				intfReq.Gateway, iface.Interface.Gateway)
			return
		}
	}
	switch intfReq.Autoneg {
	case pcc.INTERFACE_AUTONEG_ON:
		if iface.Interface.Autoneg != true {
			log.AuctaLogger.Errorf("    autoneg mismatch [%v] [%v]",
				intfReq.Autoneg, iface.Interface.Autoneg)
			return
		}
	case pcc.INTERFACE_AUTONEG_OFF:
		if iface.Interface.Autoneg != false {
			log.AuctaLogger.Errorf("    autoneg mismatch [%v] [%v]",
				intfReq.Autoneg, iface.Interface.Autoneg)
			return
		}
		if intfReq.AdminStatus == pcc.INTERFACE_STATUS_UP {
			if intfReq.Speed != json.Number(iface.Interface.Speed) {
				log.AuctaLogger.Errorf("    speed mismatch [%v] [%v]",
					intfReq.Speed, iface.Interface.Speed)
				return
			}
		}
	default:
		log.AuctaLogger.Errorf("Error: unexpected autoneg [%v]", intfReq.Autoneg)
		return
	}

	mtu := fmt.Sprintf("%v", iface.Interface.Mtu)
	if intfReq.Mtu != json.Number(mtu) {
		log.AuctaLogger.Errorf("    mtu mismatch [%v] [%v]", intfReq.Mtu, mtu)
		return
	}
	if iface.Interface.AdminStatus != "" {
		if intfReq.AdminStatus != iface.Interface.AdminStatus {
			log.AuctaLogger.Errorf("    adminStatus mismatch [%v] [%v]",
				intfReq.AdminStatus, iface.Interface.AdminStatus)
			return
		}
	}
	if iface.Interface.IsManagement {
		if intfReq.IsManagement != "true" {
			log.AuctaLogger.Errorf("    IsManagement mismatch [%v] [%v]",
				intfReq.IsManagement,
				iface.Interface.IsManagement)
			return
		}
	} else {
		if intfReq.IsManagement != "false" {
			log.AuctaLogger.Errorf("    IsManagement mismatch [%v] [%v]",
				intfReq.IsManagement,
				iface.Interface.IsManagement)
			return
		}
	}
	if intfReq.ManagedByPcc != iface.Interface.ManagedByPcc {
		log.AuctaLogger.Errorf("    ManagedByPcc mismatch [%v] [%v]",
			intfReq.ManagedByPcc, iface.Interface.ManagedByPcc)
		return
	}

	var desireIpMap = make(map[string]int, len(intfReq.Ipv4Addresses))
	for _, addr := range intfReq.Ipv4Addresses {
		desireIpMap[addr] = 1
	}
	for _, addr := range iface.Interface.Ipv4Addresses {
		if _, ok := desireIpMap[addr]; ok {
			delete(desireIpMap, addr)
		}
	}
	if len(desireIpMap) != 0 {
		log.AuctaLogger.Errorf("    Ipv4 mismatch ")
		for k, _ := range desireIpMap {
			log.AuctaLogger.Errorf("  %v", k)
		}
		return
	}

	var desire6IpMap = make(map[string]int,
		len(iface.Interface.Ipv6Addresses))

	for _, addr := range iface.Interface.Ipv6Addresses {
		if strings.HasPrefix(strings.ToLower(addr), "fe80") {
			continue // skip link local
		}
		desire6IpMap[addr] = 1
	}

	for _, addr := range intfReq.Ipv6Addresses {
		if _, ok := desire6IpMap[addr]; ok {
			delete(desire6IpMap, addr)
		}
	}
	if len(desire6IpMap) != 0 {
		log.AuctaLogger.Errorf("    Ipv6 mismatch ")
		for k, _ := range desire6IpMap {
			log.AuctaLogger.Infof("  %v", k)
		}
		return
	}

	err = nil
	return
}

func serverConfigLoop(id uint64, serverIntfs []netInterface) (done bool, err error) {
	var ifaces []*pcc.InterfaceDetail

	done = false
	if ifaces, err = Pcc.GetIfacesByNodeId(id); err != nil {
		return
	}

	log.AuctaLogger.Infof("Validating config on server %v", id)

	var intfsToCheck = make(map[string]netInterface, len(serverIntfs))
	for _, intf := range serverIntfs {
		if intf.IsManagement {
			continue // don't mess with management
		}
		intfsToCheck[intf.MacAddr] = intf
	}

	for _, intf := range serverIntfs {
		var (
			intfConfig netInterface
			mac        string
			ok         bool
			iface      *pcc.InterfaceDetail
		)

		mac = intf.MacAddr
		if intfConfig, ok = intfsToCheck[mac]; !ok {
			continue
		}

		iface, err = Pcc.GetIfaceByMacAddress(mac, ifaces)
		if err != nil {
			return
		}

		ifaceRequest := prepIfaceRequest(id, iface, intfConfig)
		err = validateIfaceConfig(ifaceRequest)
		if err == nil {
			delete(intfsToCheck, mac)
		} else {
			Pcc.SetIfaceApply(ifaceRequest)
		}
	}
	err = nil
	if len(intfsToCheck) == 0 {
		done = true
		return
	}
	return
}

func verifyNetworkConfig(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res, Env.CheckInvaders, Env.CheckServers)

	assert := test.Assert{t}

	var serverMap = make(map[uint64][]netInterface)
	for _, i := range Env.Servers {
		id := NodebyHostIP[i.HostIp]
		serverMap[id] = i.NetInterfaces
	}
	for _, i := range Env.Invaders {
		id := NodebyHostIP[i.HostIp]
		serverMap[id] = i.NetInterfaces
	}

	allDone := false
	loop := 0
	loopLimit := 50
	for !allDone {
		loop++
		log.AuctaLogger.Info("Interface config validation, loop %v", loop)
		for _, node := range Nodes {
			var (
				done bool
				err  error
			)
			ifaces, ok := serverMap[node.Id]
			if !ok {
				continue
			}
			done, err = serverConfigLoop(node.Id, ifaces)
			if err != nil {
				msg := fmt.Sprintf("Failed serverConfigLoop: %v", err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
			}
			if done {
				delete(serverMap, node.Id)
			}
		}
		if len(serverMap) == 0 {
			allDone = true
		} else {
			time.Sleep(10 * time.Second)
		}
		if loop >= loopLimit {
			msg := "Timed out verifying intferface config"
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		}
	}

}

func verifyNetworkInterfaces(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	assert := test.Assert{t}
	var nodesToCheck = make(map[uint64]uint64, len(nodeIntfMap))

	for id := range nodeIntfMap {
		nodesToCheck[id] = 1
	}

	timeout := time.After(10 * time.Minute)
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <-timeout:
			for id, _ := range nodesToCheck {
				intfs, found := nodeIntfMap[id]
				if !found {
					msg := fmt.Sprintf("map lookup failed %v", id)
					res.SetTestFailure(msg)
					log.AuctaLogger.Error(msg)
					assert.FailNow()
				}
				for _, i := range intfs {
					intf, _ := Pcc.GetIfaceById(id, i)
					if intf.Interface.IsManagement {
						continue
					}
					state := intf.Interface.IntfState
					if state != pcc.Ready {
						log.AuctaLogger.Errorf("failed to update"+
							" %v %v %v", id,
							intf.Interface.Name,
							state)
					}
				}
			}
			msg := "time out updating interfaces"
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		case <-tick:
			for id, intfs := range nodeIntfMap {
				if _, found := nodesToCheck[id]; !found {
					continue
				}
				intf_count := len(intfs)
				intf_up := 0
				for _, i := range intfs {
					intf, err := Pcc.GetIfaceById(id, i)
					if err != nil {
						return
					}
					if intf.Interface.IsManagement {
						intf_count--
						continue
					}

					state := intf.Interface.IntfState
					switch state {
					case pcc.Ready:
						intf_up++
					case pcc.Queued:
					case pcc.Updating:
					case pcc.Unknown:
					case pcc.Offline:
					default:
						msg := fmt.Sprintf("unexpected "+
							"IntfState %v",
							state)
						res.SetTestFailure(msg)
						log.AuctaLogger.Error(msg)
						assert.FailNow()
					}
				}
				if intf_count == intf_up {
					log.AuctaLogger.Infof("Node %v interfaces "+
						"updated", id)
					delete(nodesToCheck, id)
				} else {
					// work around for PCC-3012
					log.AuctaLogger.Infof("Pcc.ApplyIface(%d)", id)
					err := Pcc.ApplyIface(id)
					if err != nil {
						msg := fmt.Sprintf("Interface "+
							"apply failed: %v",
							err)
						res.SetTestFailure(msg)
						log.AuctaLogger.Error(msg)
						assert.FailNow()
					}
				}
			}
			if len(nodesToCheck) == 0 {
				//res.SetTestSkipped("No node to verify")
				//assert.SkipNow()
				return
			}
		}
	}

}

func verifyNetworkUp(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	assert := test.Assert{t}
	var nodesToCheck = make(map[uint64]uint64, len(nodeIntfMap))

	for id := range nodeIntfMap {
		nodesToCheck[id] = 1
	}

	timeout := time.After(5 * time.Minute)
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <-timeout:
			for id, _ := range nodesToCheck {
				intfs, found := nodeIntfMap[id]
				if !found {
					msg := fmt.Sprintf("map lookup failed %v", id)
					res.SetTestFailure(msg)
					log.AuctaLogger.Error(msg)
					assert.FailNow()
				}
				for _, i := range intfs {
					intf, err := Pcc.GetIfaceById(id, i)
					if err != nil {
						msg := fmt.Sprintf("getIfaceById: %v", err)
						res.SetTestFailure(msg)
						log.AuctaLogger.Error(msg)
						assert.FailNow()
					}
					if !intf.Interface.ManagedByPcc {
						continue
					}
					ifName := intf.Interface.Name
					carrier := intf.Interface.CarrierStatus
					admin := intf.Interface.AdminStatus
					log.AuctaLogger.Infof("  %v %v admin %v carrier "+
						"%v",
						id, ifName, admin, carrier)
				}
			}
			msg := "time out updating interfaces"
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		case <-tick:
			for id, intfs := range nodeIntfMap {
				if _, found := nodesToCheck[id]; !found {
					continue
				}
				intf_count := len(intfs)
				intf_up := 0
				admin_down := 0
				for _, i := range intfs {
					intf, _ := Pcc.GetIfaceById(id, i)
					if !intf.Interface.ManagedByPcc {
						intf_count--
						continue
					}
					status := intf.Interface.AdminStatus
					if status == pcc.INTERFACE_STATUS_DOWN {
						admin_down++
					}
				}

				for _, i := range intfs {
					intf, _ := Pcc.GetIfaceById(id, i)
					if !intf.Interface.ManagedByPcc {
						continue
					}
					ifName := intf.Interface.Name
					carrier := intf.Interface.CarrierStatus
					admin := intf.Interface.AdminStatus
					if admin == pcc.INTERFACE_STATUS_DOWN {
						log.AuctaLogger.Infof("  %v %v"+
							" admin down",
							id, ifName)
					} else {
						log.AuctaLogger.Infof("  %v %v carrier "+
							"%v", id, ifName,
							carrier)
						if carrier == pcc.INTERFACE_STATUS_UP {
							intf_up++
						}
					}

				}
				if intf_up+admin_down == intf_count {
					log.AuctaLogger.Infof("Node %v interfaces "+
						"all UP", id)
					delete(nodesToCheck, id)
				}
			}
			if len(nodesToCheck) == 0 {
				//res.SetTestSkipped("No node to verify")
				//assert.SkipNow()
				return
			}
		}
	}

}
