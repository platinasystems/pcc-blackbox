package main

import (
    "fmt"
    "math/rand"
    "testing"
    "time"
)

// get whole PCC ObjectsList
func testDashboardGetAllPCCObjects(t *testing.T) {
    fmt.Println("Get full PCC Objects list with no sort or pagination")
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
        id := idList[rand.Uint64() % uint64(len(idList))]
        pccObject, err := Pcc.TestDashboardObjectById(id)
        if err == nil {
            fmt.Printf("Received PCC Object with id=[%d] & name=[%s]\n", pccObject.Id, pccObject.PccObjectName)
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

func testDashboardGetChildrenObjectsByRandomId(t *testing.T) {
    fmt.Println("Get Children Objects by Random Id")
    pccObjects, err := Pcc.TestDashboardObjectList(nil, nil)
    idList := make(IdList, 0)
    if err == nil {
        for _, obj := range *pccObjects {
            idList = append(idList, obj.Id)
        }
        rand.Seed(time.Now().UnixNano())
        id := idList[rand.Uint64() % uint64(len(idList))]
        childObjects, err := Pcc.TestDashboardChildrenObjectList(id, nil, nil)
        if err == nil {
            if len(*childObjects) != 0 {
                fmt.Printf("Received [%d] Children Objects for PCC Object with id=[%d]:\n%s\n", len(*childObjects), id, Pcc.HelperGetIdAndNameOf(childObjects))
            } else {
                fmt.Printf("PCC Object with id=[%d] has no children\n", id)
            }
        }
    }

    checkError(t, err)
}

func testDashboardGetParentObjectsByRandomId(t *testing.T) {
    fmt.Println("Get Parent Objects by Random Id")
    pccObjects, err := Pcc.TestDashboardObjectList(nil, nil)
    idList := make(IdList, 0)
    if err == nil {
        for _, obj := range *pccObjects {
            idList = append(idList, obj.Id)
        }
        rand.Seed(time.Now().UnixNano())
        id := idList[rand.Uint64() % uint64(len(idList))]
        parentObjects, err := Pcc.TestDashboardParentsObjectList(id, nil, nil)
        if err == nil {
            if len(*parentObjects) != 0 {
                fmt.Printf("Received [%d] Parent Objects for PCC Object with id=[%d]:\n%s\n", len(*parentObjects), id, Pcc.HelperGetIdAndNameOf(parentObjects))
            } else {
                fmt.Printf("PCC Object with id=[%d] has no parents\n", id)
            }
        }
    }

    checkError(t, err)
}

func testDashboardGetFilteredObjects(t *testing.T) {
    fmt.Println("Get Objects filtered by health and type")
    pccObjects, err := Pcc.TestDashboardFilteredObjectList("health", "OK", nil, nil)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Printf("1. Received [%d] PCC Objects with Health=OK\n", len(*pccObjects))
    }
    pccObjects, err = Pcc.TestDashboardFilteredObjectList("health", "Warning", nil, nil)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Printf("2. Received [%d] PCC Objects with Health=Warning\n", len(*pccObjects))
    }
    pccObjects, err = Pcc.TestDashboardFilteredObjectList("type", "node", nil, nil)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Printf("3. Received [%d] PCC Objects with Type=Node\n", len(*pccObjects))
    }
    pccObjects, err = Pcc.TestDashboardFilteredObjectList("type", "cephcluster", nil, nil)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Printf("4. Received [%d] PCC Objects with Type=CephCluster\n", len(*pccObjects))
    }
    checkError(t, err)
}

func testDashboardGetAdvSearchedObjects(t *testing.T) {
    fmt.Println("Get Objects matching search criteria")
    pccObjects, err := Pcc.TestDashboardAdvSearchedObjectList("health:OK{X}type~ceph", nil, nil)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Printf("1. Received [%d] PCC Objects with Health=OK and Type contains 'Ceph'\n", len(*pccObjects))
    }
    checkError(t, err)
}
