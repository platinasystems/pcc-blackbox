package main

import (
	"fmt"
	"testing"
	"time"

	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/test"
)

func updateIpam(t *testing.T) {
	t.Run("updateIpam", addIpam)
	t.Run("updateIpamConfig", addIpamConfig)
}

func addIpam(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	assert := test.Assert{t}

	var (
		addSubReq1 pcc.SubnetObj
		newSub     pcc.SubnetObj
	)

	subs, err := Pcc.GetSubnetObj()
	if err != nil {
		msg := fmt.Sprintf("Error getting subnetObjs: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	for _, sub := range *subs {
		if sub.UsedBy != "{}" {
			log.AuctaLogger.Infof("IPAM [%v] in uses, not deleting",
				sub.Name)
			continue
		}
		log.AuctaLogger.Infof("delete IPAM %v [%v] [%v]",
			sub.Id, sub.Name, sub.Subnet)
		err = Pcc.DeleteSubnetObj(sub.Id)
		if err != nil {
			msg := fmt.Sprintf("Error deleting subnetObj: %v", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		}
	}

	addSubReq1.Name = "test-cidr"
	addSubReq1.Subnet = "10.0.201.192/26"
	addSubReq1.PubAccess = true
	addSubReq1.Routed = true

	log.AuctaLogger.Infof("Add IPAM  [%+v]", addSubReq1)
	err = Pcc.AddSubnetObj(&addSubReq1)
	if err != nil {
		msg := fmt.Sprintf("Error adding subnetObj: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
	log.AuctaLogger.Infof("After add [%+v]", addSubReq1)

	newSub.Subnet = "1.1.1.0/25"
	addSubReq1.Subnet = newSub.Subnet
	err = Pcc.UpdateSubnetObj(&addSubReq1)
	if err != nil {
		msg := fmt.Sprintf("Error adding subnetObj: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
	if addSubReq1.Subnet != newSub.Subnet {
		msg := fmt.Sprintf("Error updating subnetObj: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	err = Pcc.DeleteSubnetObj(addSubReq1.Id)
	if err != nil {
		msg := fmt.Sprintf("Error deleting subnetObj: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	subs, err = Pcc.GetSubnetObj()
	if err != nil {
		msg := fmt.Sprintf("Error getting subnetObjs: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}

func addIpamConfig(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res, Env.CheckNetIpams)

	assert := test.Assert{t}

	for _, ipam := range Env.NetIpam {
		var sub pcc.SubnetObj

		oldSub, err := Pcc.FindSubnetObj(ipam.Name)
		if err == nil {
			log.AuctaLogger.Warnf("IPAM [%v] already exists",
				oldSub.Name)
			continue
		}
		sub.Name = ipam.Name
		sub.SetSubnet(ipam.Subnet)
		sub.PubAccess = ipam.PubAccess
		sub.Routed = ipam.Routed

		log.AuctaLogger.Infof("Add IPAM [%+v]", sub)
		err = Pcc.AddSubnetObj(&sub)
		if err != nil {
			msg := fmt.Sprintf("Error adding subnetObj: %v", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		}
		log.AuctaLogger.Infof("After add [%+v]", sub)
	}
}

func delAllIpams(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	assert := test.Assert{t}

	subs, err := Pcc.GetSubnetObj()
	if err != nil {
		msg := fmt.Sprintf("Error getting subnetObjs: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	for _, sub := range *subs {
		log.AuctaLogger.Infof("Delete IPAM %v [%v] [%v]",
			sub.Id, sub.Name, sub.Subnet)
		err = Pcc.DeleteSubnetObj(sub.Id)
		if err != nil {
			msg := fmt.Sprintf("Error deleting subnetObj: %v", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		}
	}
}
