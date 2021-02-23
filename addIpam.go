package main

import (
	"fmt"
	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/test"
	"testing"
	"time"
)

func updateIpam(t *testing.T) {
	t.Run("updateIpam", addIpam)
	t.Run("updateIpamConfig", addIpamConfig)
}

func addIpam(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "addIpam")
	assert := test.Assert{t}

	var (
		addSubReq1 pcc.SubnetObj
		newSub     pcc.SubnetObj
	)

	subs, err := Pcc.GetSubnetObj()
	if err != nil {
		msg := fmt.Sprintf("Error getting subnetObjs: %v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}

	for _, sub := range *subs {
		if sub.UsedBy != "{}" {
			log.AuctaLogger.Infof("IPAM [%v] in uses, not deleting\n",
				sub.Name)
			continue
		}
		log.AuctaLogger.Infof("delete IPAM %v [%v] [%v]\n",
			sub.Id, sub.Name, sub.Subnet)
		err = Pcc.DeleteSubnetObj(sub.Id)
		if err != nil {
			msg := fmt.Sprintf("Error deleting subnetObj: %v\n", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
			return
		}
	}

	addSubReq1.Name = "test-cidr"
	addSubReq1.Subnet = "10.0.201.192/26"
	addSubReq1.PubAccess = true
	addSubReq1.Routed = true

	log.AuctaLogger.Infof("Add IPAM  [%+v]\n", addSubReq1)
	err = Pcc.AddSubnetObj(&addSubReq1)
	if err != nil {
		msg := fmt.Sprintf("Error adding subnetObj: %v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}
	log.AuctaLogger.Infof("After add [%+v]\n", addSubReq1)

	newSub.Subnet = "1.1.1.0/25"
	addSubReq1.Subnet = newSub.Subnet
	err = Pcc.UpdateSubnetObj(&addSubReq1)
	if err != nil {
		msg := fmt.Sprintf("Error adding subnetObj: %v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}
	if addSubReq1.Subnet != newSub.Subnet {
		msg := fmt.Sprintf("Error updating subnetObj: %v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}

	err = Pcc.DeleteSubnetObj(addSubReq1.Id)
	if err != nil {
		msg := fmt.Sprintf("Error deleting subnetObj: %v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}

	subs, err = Pcc.GetSubnetObj()
	if err != nil {
		msg := fmt.Sprintf("Error getting subnetObjs: %v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}
}

func addIpamConfig(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "addIpamConfig")
	assert := test.Assert{t}

	if len(Env.NetIpam) == 0 {
		log.AuctaLogger.Info("IPAM: no subnets configured\n")
		return
	}

	for _, ipam := range Env.NetIpam {
		var sub pcc.SubnetObj

		oldSub, err := Pcc.FindSubnetObj(ipam.Name)
		if err == nil {
			log.AuctaLogger.Warn("IPAM  [%v] already exists\n", oldSub.Name)
			continue
		}
		sub.Name = ipam.Name
		sub.SetSubnet(ipam.Subnet)
		sub.PubAccess = ipam.PubAccess
		sub.Routed = ipam.Routed

		log.AuctaLogger.Infof("Add IPAM  [%+v]\n", sub)
		err = Pcc.AddSubnetObj(&sub)
		if err != nil {
			msg := fmt.Sprintf("Error adding subnetObj: %v\n", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
			return
		}
		log.AuctaLogger.Infof("After add [%+v]\n", sub)
	}
}

func delAllIpams(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "delAllIpams")
	assert := test.Assert{t}

	subs, err := Pcc.GetSubnetObj()
	if err != nil {
		msg := fmt.Sprintf("Error getting subnetObjs: %v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}

	for _, sub := range *subs {
		log.AuctaLogger.Infof("Delete IPAM %v [%v] [%v]\n",
			sub.Id, sub.Name, sub.Subnet)
		err = Pcc.DeleteSubnetObj(sub.Id)
		if err != nil {
			msg := fmt.Sprintf("Error deleting subnetObj: %v\n", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
			return
		}
	}
}
