package main

import (
	"fmt"
	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/pcc-models/security"
	"testing"
	"time"
)

////
// TEST
////

// Generic test for more endpoints
func genericTest(
	t *testing.T,
	item string,
	addItem func(name string, description string) (*pcc.GenericModel, error),
	getItem func(id uint64) (*pcc.GenericModel, error),
	deleteItem func(id uint64) error,
	listItems func() ([]pcc.GenericModel, error)) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var (
		items []pcc.GenericModel
		added *pcc.GenericModel
		get   *pcc.GenericModel
		err   error
	)

	if addItem != nil {
		name := fmt.Sprintf("%s-%d", "bb_test", time.Now().Unix())
		log.AuctaLogger.Infof(fmt.Sprintf("adding the [%s] [%s]", item, name))
		if added, err = addItem(name, "blackbox test"); err == nil { // add the item
			log.AuctaLogger.Infof(fmt.Sprintf("added the [%s] [%v]", item, *added))

			if getItem != nil { // get an item and check if it is the same with added
				if get, err = getItem(added.GetId()); err == nil {
					if get.GetDescription() == added.GetDescription() && get.GetName() == added.GetName() {
						log.AuctaLogger.Infof(fmt.Sprintf("the fetched %s [%s] is ok", item, name))
					} else {
						msg := fmt.Sprintf("the %s %d is different than expected", item, added.GetId())
						res.SetTestFailure(msg)
						log.AuctaLogger.Error(msg)
						t.FailNow()
					}
				} else {
					msg := fmt.Sprintf("error getting the %s with id %d", item, added.GetId())
					res.SetTestFailure(msg)
					log.AuctaLogger.Error(msg)
					t.FailNow()
				}
			}

			if deleteItem != nil { // delete the added item
				if err = deleteItem(added.GetId()); err == nil {
					log.AuctaLogger.Infof(fmt.Sprintf("deleted the %s %d %s", item, added.GetId(), added.GetName()))
				} else {
					msg := fmt.Sprintf("error getting the %s with id %d", item, added.GetId())
					res.SetTestFailure(msg)
					log.AuctaLogger.Error(msg)
					t.FailNow()
				}
			}
		} else {
			msg := fmt.Sprintf("%v", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		}
	}

	// get all items
	if listItems != nil {
		if items, err = listItems(); err == nil {
			log.AuctaLogger.Info(fmt.Sprintf("found %d %s", len(items), item))
		} else {
			msg := fmt.Sprintf("%v", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		}
	}
}

// add, get and delete an entity. get all entities
func testUMEntity(t *testing.T) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	genericTest(t,
		"entity",
		func(name string, description string) (genericModel *pcc.GenericModel, err error) {
			return Pcc.AddEntity(name, description)
		},
		func(id uint64) (genericModel *pcc.GenericModel, err error) {
			return Pcc.GetEntity(id)
		},
		nil,
		func() (models []pcc.GenericModel, err error) {
			return Pcc.GetEntities()
		},
	)
}

// add, get and delete a role. get all roles
// test the deletion of a role still used
func testUMRole(t *testing.T) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	genericTest(t,
		"role",
		func(name string, description string) (genericModel *pcc.GenericModel, err error) {
			return Pcc.AddRole(name, description)
		},
		func(id uint64) (genericModel *pcc.GenericModel, err error) {
			return Pcc.GetSecurityRole(id)
		},
		func(id uint64) error {
			return Pcc.DeleteRole(id)
		}, func() (models []pcc.GenericModel, err error) {
			return Pcc.GetSecurityRoles()
		},
	)

	// test the deletion of a role still used
	if user, err := addTestUser(); err == nil {
		defer Pcc.DelUser(user.UserName)
		if errOk := Pcc.DeleteRole(user.Role.Id); errOk == nil {
			msg := fmt.Sprintf("the role [%d] has been deleted", user.Role.Id)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		} else {
			log.AuctaLogger.Infof(fmt.Sprintf("the system does not allowed to delete the role [%d]. [%s]", user.Role.Id, errOk.Error()))
		}
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
}

func testUMOperation(t *testing.T) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	// Convert operation to Generic Model
	convert := func(f func() ([]pcc.Operation, error)) (models []pcc.GenericModel, err error) {
		if operations, e := f(); e == nil {
			for _, operation := range operations {
				models = append(models, operation.GenericModel)
			}
		} else {
			err = e
		}
		return
	}

	genericTest(t,
		"operation",
		nil, nil, nil,
		func() (models []pcc.GenericModel, err error) {
			return convert(Pcc.GetOperations)
		},
	)

	genericTest(t,
		"groupoperation",
		nil, nil, nil,
		func() (models []pcc.GenericModel, err error) {
			return convert(Pcc.GetGroupOperations)
		},
	)
}

// add a test user
func addTestUser() (user *pcc.User, err error) {
	name := fmt.Sprintf("%d@platinasystems.net", time.Now().Unix())
	u := pcc.User{UserName: name, FirstName: "John", LastName: "Wick"}

	var roles []pcc.GenericModel

	if roles, err = Pcc.GetSecurityRoles(); err == nil {
		u.RoleId = roles[0].Id
	} else {
		return
	}

	if user, err = Pcc.AddUser(u); err == nil {
		fmt.Println(fmt.Sprintf("added the user [%d:%s] with role [%+v] and tenant [%+v]", user.Id, user.UserName, user.Role, user.Tenant))
	}
	return
}

// add and delete an user.
// add an existing user and check for the error.
// get all users.
func testUMUser(t *testing.T) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	if added, err := addTestUser(); err == nil {
		user2 := *added
		log.AuctaLogger.Info(fmt.Sprintf("trying add the same user %s", user2.UserName))
		if _, errOk := Pcc.AddUser(user2); errOk == nil {
			msg := "the duplicate user was added without issues"
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		} else {
			log.AuctaLogger.Info(fmt.Sprintf("the system does not allowed to re-add the user [%s]. [%s]", added.UserName, errOk.Error()))
		}

		if err = Pcc.DelUser(added.UserName); err == nil {
			log.AuctaLogger.Info(fmt.Sprintf("deleted the user %d %s", added.Id, added.UserName))
		} else {
			msg := fmt.Sprintf("%v", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		}
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	if users, err := Pcc.GetUsers(); err == nil {
		log.AuctaLogger.Info(fmt.Sprintf("found %d users", len(users)))
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

}

// add, get and delete a tenant. get all tenants
// try to delete a tenant still used and check for the error
func testUMTenant(t *testing.T) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var err error

	name := fmt.Sprintf("%s-%d", "bb_test", time.Now().Unix())
	tenant := &(security.Tenant{Name: name, Description: "blackbox test"})
	log.AuctaLogger.Info(fmt.Sprintf("adding the tenant %s", name))
	if tenant, err = Pcc.AddTenant(*tenant); err == nil {
		log.AuctaLogger.Info(fmt.Sprintf("added the tenant %v", *tenant))

		if tenant, err = Pcc.GetTenant(tenant.ID); err == nil {
			if tenant.Name == name {
				log.AuctaLogger.Info(fmt.Sprintf("the fetched tenant %s is ok", name))
			} else {
				msg := fmt.Sprintf("the tenant %d is different than expected", tenant.ID)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
			}
		} else {
			msg := fmt.Sprintf("%v", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		}

		if err = Pcc.DelTenant(tenant.ID); err == nil {
			log.AuctaLogger.Info(fmt.Sprintf("deleted the tenant %d %s", tenant.ID, tenant.Name))
		} else {
			msg := fmt.Sprintf("%v", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		}
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	// get all tenants
	if tenants, err := Pcc.GetTenants(); err == nil {
		log.AuctaLogger.Info(fmt.Sprintf("found %d tenants", len(tenants)))
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	// test the deletion of a tenant still used
	if user, err := addTestUser(); err == nil {
		defer Pcc.DelUser(user.UserName)
		if errOk := Pcc.DelTenant(user.Tenant.ID); errOk == nil {
			msg := fmt.Sprintf("the tenant has been deleted [%d]", user.TenantId)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		} else {
			log.AuctaLogger.Info(fmt.Sprintf("the system does not allowed to delete the tenant [%d]. [%s]", user.TenantId, errOk.Error()))
		}
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
}

// add, get and delete an user space.
func testUMUserSpace(t *testing.T) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	scope := "bb-test"
	content := "blackbox test"
	if err := Pcc.SetUserSpace("bb-test", content); err == nil {
		log.AuctaLogger.Info(fmt.Sprintf("added the user space [%s]", scope))

		var get string
		if get, err = Pcc.GetUserSpace(scope); err == nil {
			if get == content {
				log.AuctaLogger.Info(fmt.Sprintf("the fetched user space %s is ok", scope))
			} else {
				msg := fmt.Sprintf("the user space %s is different than expected %s", scope, get)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
			}
		} else {
			msg := fmt.Sprintf("%v", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		}
		if err = Pcc.DeleteUserSpace(scope); err == nil {
			log.AuctaLogger.Info(fmt.Sprintf("deleted the user space %s", scope))
		} else {
			msg := fmt.Sprintf("%v", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		}
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

}
