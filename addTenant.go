package main

import (
	"fmt"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/tiles/pccserver/security/model"
	"testing"
	"github.com/platinasystems/test"
)

func addTenant(t *testing.T) {
	t.Run("addTenant", addTenantA)
}

func addTenantA(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		tenants []model.Tenant
		tenant  model.Tenant
		tenant2 model.Tenant
		err     error
	)

	fmt.Println("Delete existing tenants")
	tenants, err = Pcc.GetTenants()
	if err != nil {
		assert.Fatalf("%v\n", err)
	}
	for _, t := range tenants {
		fmt.Printf("delete tenant %v\n", t.Name)
		Pcc.DelTenant(t.ID)
	}

	addReq := model.Tenant{
		Name:        "cust-a",
		Description: "a tenant of ROOT",
		Parent:      1,
	}

	fmt.Printf("add tenant %v\n", addReq.Name)
	err = Pcc.AddTenant(addReq)
	if err != nil {
		assert.Fatalf("%v\n", err)
	}

	fmt.Printf("find tenant %v\n", addReq.Name)
	tenant, err = Pcc.FindTenant(addReq.Name)
	if err != nil {
		assert.Fatalf("%v\n", err)
	}
	fmt.Printf("tenant %v, id %v\n", tenant.Name, tenant.ID)

	addReq2 := model.Tenant{
		Name:        "cust-b",
		Description: "a tenant of cust-b",
		Parent:      tenant.ID,
	}
	fmt.Printf("add tenant %v\n", addReq.Name)
	err = Pcc.AddTenant(addReq2)
	if err != nil {
		assert.Fatalf("%v\n", err)
	}

	fmt.Printf("find tenant %v\n", addReq2.Name)
	tenant2, err = Pcc.FindTenant(addReq2.Name)
	if err != nil {
		assert.Fatalf("%v\n", err)
	}
	fmt.Printf("tenant %v, id %v\n", tenant2.Name, tenant2.ID)

	fmt.Printf("deleting tenant %v\n", tenant2.Name)
	err = Pcc.DelTenant(tenant2.ID)
	if err != nil {
		assert.Fatalf("%v\n", err)
	} else {
	}
	_, err = Pcc.FindTenant(addReq2.Name)
	if err != nil {
		fmt.Printf("FindTenant failed as exepted on deleted tenant\n")
	} else {
		assert.Fatalf("%v\n", "Expecting failure, but didn't")
	}

	fmt.Printf("assign servers to tenant %v\n", addReq.Name)
	var nodes []uint64
	for _, i := range Env.Servers {
		nodes = append(nodes, NodebyHostIP[i.HostIp])
	}
	err = Pcc.AssignTenantNodes(tenant.ID, nodes)
	if err != nil {
		assert.Fatalf("%v\n", err)
	}

	source := fmt.Sprintf("https://%v:7654/setPass", Env.PccIp)
	addUser := pcc.AddUser{
		UserName:  "BadBart",
		FirstName: "Bart",
		LastName:  "Simpson",
		Email:     "stig@platinasystems.com",
		Password:  "lisasux",
		Active:    true,
		Protect:   false,
		RoleId:    1,
		TenantId:  1,
		Source:    source,
	}
	err = Pcc.AddUser(addUser)
	if err != nil {
		assert.Fatalf("%v\n", err)
	}

	fmt.Printf("try change password\n")
	req := pcc.AddUser{
		UserName: "BadBart",
		Active:   false,
	}

	err = Pcc.UpdateUser(req)
	if err != nil {
		assert.Fatalf("%v\n", err)
	}

	users, err := Pcc.GetUsers()
	if err != nil {
		assert.Fatalf("%v\n", err)
	}
	for _, u := range users {
		if u.Username == "BadBart" {
			fmt.Printf("Found added user %v\n", u)
			if u.Enabled == false {
				fmt.Printf("user update worked\n")
			} else {
				assert.Fatalf("user update failed\n")
			}
		} else {
			fmt.Printf("user %v\n", u)
		}
	}

	err = Pcc.DelUser(addUser.UserName)
	if err != nil {
		assert.Fatalf("%v\n", err)
	}
}
