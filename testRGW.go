package main

import (
	"fmt"
	"github.com/platinasystems/pcc-models/ceph"
	"testing"
)

////
// TEST Rados Gateway
////

// add, get and delete an entity. get all entities
func deployRadosGateway(t *testing.T) {
	var (
		err   error
		added ceph.RadosGateway
	)
	//TODO create the pool or check existing

	gw := ceph.RadosGateway{Name: "blackbox", CephPoolID: 0}

	// TODO check if the gw is still running
	if added, err = Pcc.AddRadosGW(&gw); err == nil {
		fmt.Println("\nAdded the Rados Gateway ", added.ID)

	} else {
		t.Fatal(err)
	}
}
