package pcc

import (
	"fmt"
	"github.com/platinasystems/pcc-models/ceph"
)

////
// Rados Rest Client
////

// Get all Rados Gateways
func (pcc *PccClient) GetRadosGWs(serviceType string) (gws []ceph.RadosGateway, err error) {
	err = pcc.Get("pccserver/storage/ceph/rgws", &gws)
	return
}

// Get a Rados Gateway
func (pcc *PccClient) GetRadosGW(id uint64) (gw ceph.RadosGateway, err error) {
	err = pcc.Get(fmt.Sprintf("pccserver/storage/ceph/rgws/%d", id), &gw)
	return
}

// Add a Rados Gateway
func (pcc *PccClient) AddRadosGW(gw *ceph.RadosGateway) (added ceph.RadosGateway, err error) {
	err = pcc.Post("ppccserver/storage/ceph/rgws", gw, &added)
	return
}

// Update a Rados Gateway
func (pcc *PccClient) UpdateRadosGW(gw *ceph.RadosGateway) (added ceph.RadosGateway, err error) {
	err = pcc.Put(fmt.Sprintf("ppccserver/storage/ceph/rgws/%d", gw.ID), gw, &added)
	return
}

// Delete a Rados Gateway
func (pcc *PccClient) DeleteRadosGW(id uint64) (deleted ceph.RadosGateway, err error) {
	err = pcc.Delete(fmt.Sprintf("ppccserver/storage/ceph/rgws/%d", id), nil, &deleted)
	return
}
