package pcc

import (
	"fmt"
	"github.com/platinasystems/pcc-models/ceph"
)

////
// Rados Rest Client
////
const (
	RGW_DEPLOY_STATUS_PROGRESS  = "installing"
	RGW_DEPLOY_STATUS_COMPLETED = "completed"
	RGW_DEPLOY_STATUS_FAILED    = "failed"
)

// Get all Rados Gateways
func (pcc *PccClient) GetRadosGWs() (gws []ceph.RadosGateway, err error) {
	err = pcc.Get("pccserver/v2/storage/ceph/rgws", &gws)
	return
}

// Get a Rados Gateway
func (pcc *PccClient) GetRadosGW(id uint64) (gw ceph.RadosGateway, err error) {
	err = pcc.Get(fmt.Sprintf("pccserver/v2/storage/ceph/rgws/%d", id), &gw)
	return
}

// Add a Rados Gateway
func (pcc *PccClient) AddRadosGW(gw *ceph.RadosGateway) (added ceph.RadosGateway, err error) {
	err = pcc.Post("pccserver/v2/storage/ceph/rgws", gw, &added)
	return
}

// Update a Rados Gateway
func (pcc *PccClient) UpdateRadosGW(gw *ceph.RadosGateway) (added ceph.RadosGateway, err error) {
	err = pcc.Put(fmt.Sprintf("pccserver/v2/storage/ceph/rgws/%d", gw.ID), gw, &added)
	return
}

// Delete a Rados Gateway
func (pcc *PccClient) DeleteRadosGW(id uint64) (deleted ceph.RadosGateway, err error) {
	err = pcc.Delete(fmt.Sprintf("pccserver/v2/storage/ceph/rgws/%d", id), nil, &deleted)
	return
}
