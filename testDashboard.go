package main

import (
	"math/rand"
	"net/url"
	"testing"
	"time"

	log "github.com/platinasystems/go-common/logs"
	"github.com/platinasystems/pcc-blackbox/models"
)

type IdList []uint64

// get whole PCC ObjectsList
func testDashboardGetAllPCCObjects(t *testing.T) {
	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Info("Get full PCC Objects list with no sort or pagination")
	pccObjects, err := Pcc.TestDashboardObjectList(nil, nil)
	if err != nil {
		log.AuctaLogger.Error(err.Error())
	} else {
		log.AuctaLogger.Infof("Received [%d] PCC Objects", len(*pccObjects))
	}
	checkError(t, res, err)
}

// get Index of PCC Objects (concise Objects List)
func testDashboardGetAllConcisePCCObjects(t *testing.T) {
	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Infof("Get Index of PCC Objects (concise format) with no sort or pagination")
	pccObjectsIndex, err := Pcc.TestDashboardConciseObjectList(nil, nil)
	if err != nil {
		log.AuctaLogger.Error(err.Error())
	} else {
		log.AuctaLogger.Infof("Received [%d] Concise PCC Objects", len(*pccObjectsIndex))
	}
	checkError(t, res, err)
}

func testDashboardGetPCCObjectByRandomId(t *testing.T) {
	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Infof("Get PCCObject by Random Id")
	pccObjects, err := Pcc.TestDashboardObjectList(nil, nil)
	idList := make(IdList, 0)
	if err == nil {
		for _, obj := range *pccObjects {
			idList = append(idList, obj.Id)
		}
		rand.Seed(time.Now().UnixNano())
		id := idList[rand.Uint64()%uint64(len(idList))]
		pccObject, err := Pcc.TestDashboardObjectById(id)
		if err == nil {
			log.AuctaLogger.Infof("Received PCC Object with id=[%d] & name=[%s]", pccObject.Id, pccObject.PccObjectName)
		}
	}

	checkError(t, res, err)
}

func testDashboardGetPCCObjectById(t *testing.T) {
	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Info("Get PCCObject by Id")
	pccObject, err := Pcc.TestDashboardObjectById(4)
	_ = pccObject
	if err != nil {
		log.AuctaLogger.Error(err.Error())
	} else {
		log.AuctaLogger.Info("Received 1 PCC Object")
	}
	checkError(t, res, err)
}

func testDashboardGetChildrenObjectsByRandomId(t *testing.T) {
	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Info("Get Children Objects by Random Id")
	pccObjects, err := Pcc.TestDashboardObjectList(nil, nil)
	idList := make(IdList, 0)
	if err == nil {
		for _, obj := range *pccObjects {
			idList = append(idList, obj.Id)
		}
		rand.Seed(time.Now().UnixNano())
		id := idList[rand.Uint64()%uint64(len(idList))]
		childObjects, err := Pcc.TestDashboardChildrenObjectList(id, nil, nil)
		if err == nil {
			if len(*childObjects) != 0 {
				log.AuctaLogger.Infof("Received [%d] Children Objects for PCC Object with id=[%d]:%s", len(*childObjects), id, Pcc.HelperExtractIdAndNameOf(childObjects))
			} else {
				log.AuctaLogger.Infof("PCC Object with id=[%d] has no children", id)
			}
		}
	}

	checkError(t, res, err)
}

func testDashboardGetParentObjectsByRandomId(t *testing.T) {
	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Info("Get Parent Objects by Random Id")
	pccObjects, err := Pcc.TestDashboardObjectList(nil, nil)
	idList := make(IdList, 0)
	if err == nil {
		for _, obj := range *pccObjects {
			idList = append(idList, obj.Id)
		}
		rand.Seed(time.Now().UnixNano())
		id := idList[rand.Uint64()%uint64(len(idList))]
		parentObjects, err := Pcc.TestDashboardParentsObjectList(id, nil, nil)
		if err == nil {
			if len(*parentObjects) != 0 {
				log.AuctaLogger.Infof("Received [%d] Parent Objects for PCC Object with id=[%d]:%s", len(*parentObjects), id, Pcc.HelperExtractIdAndNameOf(parentObjects))
			} else {
				log.AuctaLogger.Infof("PCC Object with id=[%d] has no parents", id)
			}
		}
	}

	checkError(t, res, err)
}

func testDashboardGetFilteredObjects(t *testing.T) {
	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Info("Get Objects filtered by health and type")
	pccObjects, err := Pcc.TestDashboardFilteredObjectList("health", "OK", nil, nil)
	if err != nil {
		log.AuctaLogger.Error(err.Error())
	} else {
		log.AuctaLogger.Infof("1. Received [%d] PCC Objects with Health=OK", len(*pccObjects))
	}
	pccObjects, err = Pcc.TestDashboardFilteredObjectList("health", "Warning", nil, nil)
	if err != nil {
		log.AuctaLogger.Error(err.Error())
	} else {
		log.AuctaLogger.Infof("2. Received [%d] PCC Objects with Health=Warning", len(*pccObjects))
	}
	pccObjects, err = Pcc.TestDashboardFilteredObjectList("type", "node", nil, nil)
	if err != nil {
		log.AuctaLogger.Error(err.Error())
	} else {
		log.AuctaLogger.Infof("3. Received [%d] PCC Objects with Type=Node", len(*pccObjects))
	}
	pccObjects, err = Pcc.TestDashboardFilteredObjectList("type", "cephcluster", nil, nil)
	if err != nil {
		log.AuctaLogger.Error(err.Error())
	} else {
		log.AuctaLogger.Infof("4. Received [%d] PCC Objects with Type=CephCluster", len(*pccObjects))
	}
	checkError(t, res, err)
}

func testDashboardGetAdvSearchedObjects(t *testing.T) {
	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Info("Get Objects matching search criteria")
	pccObjects, err := Pcc.TestDashboardAdvSearchedObjectList(url.QueryEscape("health:OK{X}type~ceph"), nil, nil)
	if err != nil {
		log.AuctaLogger.Error(err.Error())
	} else {
		log.AuctaLogger.Infof("1. Received [%d] PCC Objects with Health=OK and Type contains 'Ceph'", len(*pccObjects))
	}
	pccObjects, err = Pcc.TestDashboardAdvSearchedObjectList(url.QueryEscape("type:node{X}group~video"), nil, nil)
	if err != nil {
		log.AuctaLogger.Error(err.Error())
	} else {
		log.AuctaLogger.Infof("2. Received [%d] PCC Objects with Type=Node and Group contains 'Video'", len(*pccObjects))
	}
	pageParams := "page=0&limit=25"
	sortParams := "sortBy=Name&sortDir=asc"
	pccObjects, err = Pcc.TestDashboardAdvSearchedObjectList(url.QueryEscape("tag~storage{X}name~srv2"), &pageParams, nil)
	if err != nil {
		log.AuctaLogger.Error(err.Error())
	} else {
		log.AuctaLogger.Infof("3. Received [%d] PCC Objects with Tag contains Storage and Name contains 'srv2' with pagination=page=0&limit=25", len(*pccObjects))
	}
	pccObjects, err = Pcc.TestDashboardAdvSearchedObjectList(url.QueryEscape("any~avail{X}health:OK"), nil, &sortParams)
	if err != nil {
		log.AuctaLogger.Error(err.Error())
	} else {
		log.AuctaLogger.Infof("4. Received [%d] PCC Objects with Any contains 'Avail' and Health=OK with sort by Name", len(*pccObjects))
	}
	pccObjects, err = Pcc.TestDashboardAdvSearchedObjectList(url.QueryEscape("name~srv4,tenantName~root"), &pageParams, &sortParams)
	if err != nil {
		log.AuctaLogger.Error(err.Error())
	} else {
		log.AuctaLogger.Infof("5. Received [%d] PCC Objects with Name contains 'srv4' and Tenant=ROOT with pagination=page=0&limit=25 and sort by Name", len(*pccObjects))
	}
	pccObjects, err = Pcc.TestDashboardAdvSearchedObjectList(url.QueryEscape("type~network{U}name~ceph"), &pageParams, &sortParams)
	if err != nil {
		log.AuctaLogger.Error(err.Error())
	} else {
		log.AuctaLogger.Infof("6. Received [%d] PCC Objects with Type conatins Network and Name contains 'ceph'", len(*pccObjects))
	}
	checkError(t, res, err)
}

func testDashboardGetAggrHealthCountByType(t *testing.T) {
	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Info("Get Aggregate Health Count grouped by Type")
	health, err := Pcc.TestDashboardAggrHealthCountByType()
	if err != nil {
		log.AuctaLogger.Error(err.Error())
	} else {
		log.AuctaLogger.Infof("Received Aggregate Health Count grouped by Type:%s", Pcc.HelperExtractHealthByTypeFrom(health))
	}
	checkError(t, res, err)
}

func testDashboardGetMetadataEnumStrings(t *testing.T) {
	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Info("Get Dashboard Metadata Enum Strings")
	metadata, err := Pcc.TestDashboardMetadataEnumStrings()
	if err != nil {
		log.AuctaLogger.Error(err.Error())
	} else {
		log.AuctaLogger.Infof("Received Dashboard Metadata Enum Strings:[%+v]", *metadata)
	}
	checkError(t, res, err)
}
