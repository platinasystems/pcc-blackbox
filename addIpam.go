package main

import (
	"fmt"
	"testing"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

func updateIpam(t *testing.T) {
	t.Run("updateIpam", addIpam)
}

func addIpam(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	var (
		addSubReq1 pcc.SubnetObj
		newSub     pcc.SubnetObj
	)

	subs, err := Pcc.GetSubnetObj()
	if err != nil {
		assert.Fatalf("Error getting subnetObjs: %v\n", err)
		return
	}

	for _, sub := range *subs {
		fmt.Printf("Delete IPAM %v [%v] [%v]\n",
			sub.Id, sub.Name, sub.Subnet)
		err = Pcc.DeleteSubnetObj(sub.Id)
		if err != nil {
			assert.Fatalf("Error deleting subnetObj: %v\n", err)
			return
		}
	}

	addSubReq1.Name = "test-cidr"
	addSubReq1.Subnet = "10.0.201.192/26"
	addSubReq1.PubAccess = true
	addSubReq1.Routed = true

	fmt.Printf("Add IPAM  [%+v]\n", addSubReq1)
	err = Pcc.AddSubnetObj(&addSubReq1)
	if err != nil {
		assert.Fatalf("Error adding subnetObj: %v\n", err)
		return
	}
	fmt.Printf("After add [%+v]\n", addSubReq1)

	newSub.Subnet = "1.1.1.0/25"
	addSubReq1.Subnet = newSub.Subnet
	err = Pcc.UpdateSubnetObj(&addSubReq1)
	if err != nil {
		assert.Fatalf("Error adding subnetObj: %v\n", err)
		return
	}
	if addSubReq1.Subnet != newSub.Subnet {
		assert.Fatalf("Error updating subnetObj: %v\n", err)
		return
	}

	err = Pcc.DeleteSubnetObj(addSubReq1.Id)
	if err != nil {
		assert.Fatalf("Error deleting subnetObj: %v\n", err)
		return
	}

	subs, err = Pcc.GetSubnetObj()
	if err != nil {
		assert.Fatalf("Error getting subnetObjs: %v\n", err)
		return
	}
	if len(*subs) != 0 {
		assert.Fatalf("Error expecting 0 subnetObj: %v\n", len(*subs))
		return
	}
}
