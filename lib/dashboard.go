package pcc

import (
	_ "fmt"

	dashboardctl "github.com/platinasystems/tiles/pccserver/controllers"
	"github.com/platinasystems/tiles/pccserver/pccobject"
)

type DashboardObj struct {
    dashboardctl.PccObjectOutput
}

type DashboardObjList struct {
    dashboardctl.PccObjectOutputList
}

func (client *PccClient) TestDashboardObjectList(sortParam *string, pageParam *string) (objects *DashboardObjList, e error) {
    // sort & pagination params may not be specified, i.e. the params are either nil or empty string
    var (
		result DashboardObjList
		err error
	)
    endPoint := "pccserver/dashboard/objects"
    if sortParam != nil {
        endPoint = endPoint + "?" + endPoint
    }
    if pageParam != nil {
        endPoint = endPoint + "&" + endPoint
    }
    if err = client.Get(endPoint, &result); err == nil {
		objects = &result
	} else {
	    objects = nil
	}
    return objects, err
}

func (client *PccClient) TestDashboardObjectById(id uint64) (*DashboardObj, error) {
    return nil, nil
}

func (client *PccClient) TestDashboardFilteredObjectList(sortParam *string, pageParam *string) (*DashboardObjList, error) {
    // sort & pagination params may not be specified, i.e. the params are either nil or empty string

    return nil, nil
}

func (client *PccClient) TestDashboardAdvSearchedObjectList(sortParam *string, pageParam *string) (*DashboardObjList, error) {
    // sort & pagination params may not be specified, i.e. the params are either nil or empty string

    return nil, nil
}

func (client *PccClient) TestDashboardChildrenObjectList(id uint64, sortParam *string, pageParam *string) (*DashboardObjList, error) {
    // sort & pagination params may not be specified, i.e. the params are either nil or empty string

    return nil, nil
}

func (client *PccClient) TestDashboardParentsObjectList(id uint64, sortParam *string, pageParam *string) (*DashboardObjList, error) {
    // sort & pagination params may not be specified, i.e. the params are either nil or empty string

    return nil, nil
}

func (client *PccClient) TestDashboardAggrHealthCountByType() (*pccobject.AggrHealthCountByType, error) {

    return nil, nil
}


