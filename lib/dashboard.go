package pcc

import (
     "fmt"

    dashboardctl "github.com/platinasystems/tiles/pccserver/controllers"
    "github.com/platinasystems/tiles/pccserver/pccobject"
)

// Are these wrapper types helpful? CHECK LATER...
// >>>>>>
type DashboardObj struct {
    dashboardctl.PccObjectOutput    `json:"PccObjectOutput"`
}

type DashboardObjList []DashboardObj
// <<<<<<

func (client *PccClient) TestDashboardObjectList(sortParam *string, pageParam *string) (objects *[]dashboardctl.PccObjectOutput, err error) {
    // sort & pagination params may not be specified, i.e. the params are either nil or empty string
    endPoint := "pccserver/dashboard/objects"
    var result []dashboardctl.PccObjectOutput = make([]dashboardctl.PccObjectOutput, 0)

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

func (client *PccClient) TestDashboardObjectById(id uint64) (obj *dashboardctl.PccObjectOutput, err error) {
    endPoint := fmt.Sprintf("pccserver/dashboard/objects/%d", id)
    var result dashboardctl.PccObjectOutput

    if err = client.Get(endPoint, &result); err == nil {
        obj = &result
    } else {
        obj = nil
    }
    return obj, err
}

func (client *PccClient) TestDashboardFilteredObjectList(field string, value string, sortParam *string, pageParam *string) (objects *[]dashboardctl.PccObjectOutput, err error) {
    // sort & pagination params may not be specified, i.e. the params are either nil or empty string
    endPoint := fmt.Sprintf("pccserver/dashboard/objects/%s/%s", field, value)
    var result []dashboardctl.PccObjectOutput = make([]dashboardctl.PccObjectOutput, 0)

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

func (client *PccClient) TestDashboardAdvSearchedObjectList(filter string, sortParam *string, pageParam *string) (objects *[]dashboardctl.PccObjectOutput, err error) {
    // sort & pagination params may not be specified, i.e. the params are either nil or empty string 
    endPoint := fmt.Sprintf("pccserver/dashboard/objects/search?filter=%s", filter)
    var result []dashboardctl.PccObjectOutput = make([]dashboardctl.PccObjectOutput, 0)

    if sortParam != nil {
        endPoint = endPoint + "&" + endPoint
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

func (client *PccClient) TestDashboardChildrenObjectList(id uint64, sortParam *string, pageParam *string) (objects *[]dashboardctl.PccObjectOutput, err error) {
    // sort & pagination params may not be specified, i.e. the params are either nil or empty string
    endPoint := fmt.Sprintf("pccserver/dashboard/objects/childrenOf/%d", id)
    var result []dashboardctl.PccObjectOutput = make([]dashboardctl.PccObjectOutput, 0)

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

func (client *PccClient) TestDashboardParentsObjectList(id uint64, sortParam *string, pageParam *string) (objects *[]dashboardctl.PccObjectOutput, err error) {
    // sort & pagination params may not be specified, i.e. the params are either nil or empty string
    endPoint := fmt.Sprintf("pccserver/dashboard/objects/parentsOf/%d", id)
    var result []dashboardctl.PccObjectOutput = make([]dashboardctl.PccObjectOutput, 0)

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

func (client *PccClient) TestDashboardAggrHealthCountByType() (healthCount *pccobject.AggrHealthCountByType, err error) {
    endPoint := fmt.Sprintf("pccserver/dashboard/stats/health/countByType")
    var result pccobject.AggrHealthCountByType

    if err = client.Get(endPoint, &result); err == nil {
        healthCount = &result
    } else {
        healthCount = nil
    }
    return healthCount, err
}

func (client *PccClient) TestDashboardMetadataEnumStrings() (metadataStrings *dashboardctl.DashboardStringsList, err error) {
    endPoint := fmt.Sprintf("pccserver/dashboard/metadata/enum/strings")
    var result dashboardctl.DashboardStringsList

    if err = client.Get(endPoint, &result); err == nil {
        metadataStrings = &result
    } else {
        metadataStrings = nil
    }
    return metadataStrings, err
}
