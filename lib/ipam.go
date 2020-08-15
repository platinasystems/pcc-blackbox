// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"fmt"

	"github.com/platinasystems/tiles/pccserver/models"
)

const subEndpoint string = "pccserver/subnet-objs"

type SubnetStr struct {
	models.SubnetStr
}

type SubnetObj struct {
	models.SubnetObj
}

func (pcc *PccClient) GetSubnetObj() (subnetObjs *[]SubnetObj, err error) {
	var subs []SubnetObj
	if err = pcc.Get(subEndpoint, &subs); err == nil {
		subnetObjs = &subs
	}
	return
}

func (pcc *PccClient) AddSubnetObj(subnetObj *SubnetObj) (err error) {
	if subnetObj.Id != 0 {
		err = fmt.Errorf("Invalid subnetObj Id [%v] for AddSubnetObj",
			subnetObj.Id)
		return
	}
	err = pcc.Post(subEndpoint, subnetObj, subnetObj)
	if err != nil {
		return
	}
	return
}

func (pcc *PccClient) UpdateSubnetObj(subnetObj *SubnetObj) (err error) {
	if subnetObj.Id == 0 {
		err = fmt.Errorf("Invalid subnetObj Id 0 for UpdateSubnetObj")
		return
	}
	err = pcc.Put(subEndpoint, subnetObj, subnetObj)
	return
}

func (pcc *PccClient) DeleteSubnetObj(id uint64) (err error) {
	err = pcc.Delete(fmt.Sprintf("%v/%v", subEndpoint, id), nil, nil)
	return
}
