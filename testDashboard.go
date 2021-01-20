package main

import (
    "fmt"
    "math/rand"
    "testing"
    "time"
)

// get whole PCC ObjectsList
func testDashboardGetAllPCCObjects(t *testing.T) {
    fmt.Println("Get whole PCCObjects list with no sort or pagination")
    pccObjects, err := Pcc.TestDashboardObjectList(nil, nil)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Printf("Received [%d] PCC Objects\n", len(*pccObjects))
    }
    checkError(t, err)
}

type IdList []uint64

func testDashboardGetPCCObjectByRandomId(t *testing.T) {
    fmt.Println("Get PCCObject by Random Id")
    pccObjects, err := Pcc.TestDashboardObjectList(nil, nil)
    idList := make(IdList, 0)
    if err == nil {
        for _, obj := range *pccObjects {
            idList = append(idList, obj.Id)
        }
        rand.Seed(time.Now().UnixNano())
        id := rand.Uint64() % uint64(len(idList))
        pccObject, err := Pcc.TestDashboardObjectById(id)
        if err == nil {
            fmt.Printf("Received PCC Object with id=[%d]\n", pccObject.Id)
        }
    }

    checkError(t, err)
}

func testDashboardGetPCCObjectById(t *testing.T) {
    fmt.Println("Get PCCObject by Id")
    pccObject, err := Pcc.TestDashboardObjectById(4)
    _ = pccObject
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println("Received 1 PCC Object")
    }
    checkError(t, err)
}
