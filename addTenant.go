package main

import (
	"fmt"
	"testing"
	"time"

	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/pcc-models/security"
	"github.com/platinasystems/test"
)

func addTenant(t *testing.T) {
	t.Run("addTenant", addTenantA)
}

func addTenantA(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res, Env.CheckServers, Env.CheckPccIp)

	assert := test.Assert{t}
	var (
		tenants []security.Tenant
		tenant  security.Tenant
		tenant2 security.Tenant
		addReq  security.Tenant
		addReq2 security.Tenant
		err     error
	)

	log.AuctaLogger.Infof("assign all nodes to ROOT")
	var nodes []uint64
	for _, i := range Env.Servers {
		nodes = append(nodes, NodebyHostIP[i.HostIp])
	}
	err = Pcc.AssignTenantNodes(1, nodes)
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	log.AuctaLogger.Infof("Delete existing tenants")
	tenants, err = Pcc.GetTenants()
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
	for _, t := range tenants {
		log.AuctaLogger.Infof("delete tenant %v", t.Name)
		Pcc.DelTenant(t.ID)
	}

	addReq = security.Tenant{}
	addReq.Name = "cust-a"
	addReq.Description = "a tenant of ROOT"

	log.AuctaLogger.Infof("add tenant %v", addReq.Name)
	_, err = Pcc.AddTenant(addReq)
	if err != nil {
		msg := fmt.Sprintf("Failed to add tenant %v: %v", addReq.Name, err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	log.AuctaLogger.Infof("find tenant %v", addReq.Name)
	tenant, err = Pcc.FindTenant(addReq.Name)
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
	log.AuctaLogger.Infof("tenant %v, id %v", tenant.Name, tenant.ID)

	addReq2.Name = "cust-b"
	addReq2.Description = "a tenant of cust-b"
	addReq2.Parent = tenant.ID

	log.AuctaLogger.Infof("add tenant %v", addReq.Name)
	_, err = Pcc.AddTenant(addReq2)
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	log.AuctaLogger.Infof("find tenant %v", addReq2.Name)
	tenant2, err = Pcc.FindTenant(addReq2.Name)
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
	log.AuctaLogger.Infof("tenant %v, id %v", tenant2.Name, tenant2.ID)

	log.AuctaLogger.Infof("deleting tenant %v", tenant2.Name)
	err = Pcc.DelTenant(tenant2.ID)
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	} else {
	}
	_, err = Pcc.FindTenant(addReq2.Name)
	if err != nil {
		log.AuctaLogger.Infof("FindTenant failed as expected on deleted tenant")
	} else {
		msg := fmt.Sprintf("%v", "Expecting failure, but didn't")
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	log.AuctaLogger.Infof("assign servers to tenant %v", addReq.Name)
	for _, i := range Env.Servers {
		nodes = append(nodes, NodebyHostIP[i.HostIp])
	}
	err = Pcc.AssignTenantNodes(tenant.ID, nodes)
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	source := fmt.Sprintf("https://%v:7654/setPass", Env.PccIp)

	// remove existing users
	if users, err := Pcc.GetUsers(); err != nil {
		msg := fmt.Sprintf("Failed to get users: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	} else {
		for _, u := range users {
			if u.UserName == "admin" {
				continue
			}
			err = Pcc.DelUser(u.UserName)
			if err != nil {
				msg := fmt.Sprintf("Failed to delete user %v: %v",
					u.UserName, err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
				return
			}
		}
	}

	addUser := pcc.User{
		UserName:  "bsimpson@platinasystems.com",
		FirstName: "Bart",
		LastName:  "Simpson",
		Email:     "bsimpson@platinasystems.com",
		Password:  "lisasux",
		Active:    true,
		Protect:   false,
		RoleId:    1,
		TenantId:  tenant.ID,
		Source:    source,
	}
	_, err = Pcc.AddUser(addUser)
	if err != nil {
		msg := fmt.Sprintf("Failed to add user %v: %v", addUser.UserName,
			err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}

	addUser2 := pcc.User{
		UserName:  "lsimpson@platinasystems.com",
		FirstName: "Lisa",
		LastName:  "Simpson",
		Email:     "lsimpson@platinasystems.com",
		Password:  "bartsux",
		Active:    true,
		Protect:   false,
		RoleId:    1,
		TenantId:  tenant.ID,
		Source:    source,
	}
	if _, err = Pcc.AddUser(addUser2); err != nil {
		msg := fmt.Sprintf("Failed to add user %v: %v", addUser2.UserName,
			err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}

	log.AuctaLogger.Infof("Try change firstname of user %v ", addUser.UserName)
	newName := "Mr Bart"
	addUser.FirstName = newName

	if err = Pcc.UpdateUser(addUser); err != nil {
		msg := fmt.Sprintf("Failed to update user %v: %v", newName, err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}

	if users, err := Pcc.GetUsers(); err == nil {
		found := false
		for _, u := range users {
			if u.UserName == addUser.UserName {
				log.AuctaLogger.Infof("Found updated user %v", u)
				if u.Profile.FirstName == newName {
					log.AuctaLogger.Infof("user update worked")
					found = true
				} else {
					msg := fmt.Sprintf("user update failed")
					res.SetTestFailure(msg)
					log.AuctaLogger.Error(msg)
					assert.FailNow()
					return
				}
			}
		}
		if !found {
			msg := fmt.Sprintf("user update failed and not found")
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
			return
		}
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	err = Pcc.DelUser(addUser.UserName)
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}

func delAllTenants(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	assert := test.Assert{t}
	var (
		tenants []security.Tenant
		err     error
	)

	if tenants, err = Pcc.GetTenants(); err == nil {
		for _, t := range tenants {
			if t.Protect {
				continue
			}
			id := t.ID
			if err = Pcc.DelTenant(id); err != nil {
				msg := fmt.Sprintf("Failed to DelTenant %v: %v", id, err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
				return
			}
		}
	} else {
		msg := fmt.Sprintf("Failed to GetTenants: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}

func delAllUsers(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	assert := test.Assert{t}
	var (
		users []pcc.User
		err   error
	)

	if users, err = Pcc.GetUsers(); err == nil {
		for _, u := range users {
			if u.Protect {
				continue
			}
			username := u.UserName
			if err = Pcc.DelUser(username); err != nil {
				msg := fmt.Sprintf("failed to Delete user %v: %v", username, err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
				return
			}
		}
	} else {
		msg := fmt.Sprintf("Failed to GetUsers: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}
}
