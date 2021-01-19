package main

import (
	"fmt"
	"testing"
_	"time"
_	"github.com/platinasystems/pcc-blackbox/lib"
)

// get whole PCC ObjectsList
func testDashboardGetAllPCCObjects(t *testing.T) {
	fmt.Println("Get whole PCCObjects list with no sort or pagination")
	pccObjects, err := Pcc.TestDashboardObjectList(nil, nil)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Received # [%+v] objects", *pccObjects)
	}
}
