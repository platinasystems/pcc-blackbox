package main

import (
	"fmt"
	"testing"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

func addTenant(t *testing.T) {
	t.Run("addTenant", addTenantA)
}

func addTenantA(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		tenants []pcc.Tenant
		tenant  pcc.Tenant
		tenant2 pcc.Tenant
		addReq  pcc.Tenant
		addReq2 pcc.Tenant
		err     error
	)

	fmt.Println("assign all nodes to ROOT")
	var nodes []uint64
	for _, i := range Env.Servers {
		nodes = append(nodes, NodebyHostIP[i.HostIp])
	}
	err = Pcc.AssignTenantNodes(1, nodes)
	if err != nil {
		assert.Fatalf("%v\n", err)
	}

	fmt.Println("Delete existing tenants")
	tenants, err = Pcc.GetTenants()
	if err != nil {
		assert.Fatalf("%v\n", err)
	}
	for _, t := range tenants {
		fmt.Printf("delete tenant %v\n", t.Name)
		Pcc.DelTenant(t.ID)
	}

	addReq.Name = "cust-a"
	addReq.Description = "a tenant of ROOT"
	addReq.Parent = 1

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

	addReq2.Name = "cust-b"
	addReq2.Description = "a tenant of cust-b"
	addReq2.Parent = tenant.ID

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
		TenantId:  tenant.ID,
		Source:    source,
	}
	err = Pcc.AddUser(addUser)
	if err != nil {
		assert.Fatalf("%v\n", err)
	}

	fmt.Printf("try change password\n")
	newEmail := "homer@lovesdonuts.io"
	addUser.Email = newEmail

	err = Pcc.UpdateUser(addUser)
	if err != nil {
		assert.Fatalf("%v\n", err)
	}

	users, err := Pcc.GetUsers()
	if err != nil {
		assert.Fatalf("%v\n", err)
	}

	found := false
	for _, u := range users {
		if u.UserName == "BadBart" {
			fmt.Printf("Found updated user %v\n", u)
			if u.Email == newEmail {
				fmt.Printf("user update worked\n")
				found = true
			} else {
				assert.Fatalf("user update failed\n")
			}
		}
	}
	if !found {
		assert.Fatalf("user update failed and not found\n")
	}

	err = Pcc.DelUser(addUser.UserName)
	if err != nil {
		assert.Fatalf("%v\n", err)
	}
}

func delAllTenants(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		tenants []pcc.Tenant
		err     error
	)

	tenants, err = Pcc.GetTenants()
	if err != nil {
		assert.Fatalf("Failed to GetTenants: %v\n", err)
		return
	}
	for _, t := range tenants {
		id := t.ID
		if t.Name == "ROOT" {
			continue
		}
		err = Pcc.DelTenant(id)
		if err != nil {
			assert.Fatalf("Failed to DelTenant %v: %v\n", id, err)
			return
		}
	}
}

func delAllUsers(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		users []pcc.User
		err   error
	)

	users, err = Pcc.GetUsers()
	if err != nil {
		assert.Fatalf("Failed to GetUsers: %v\n", err)
		return
	}
	for _, u := range users {
		user := u.UserName
		if user == "admin" {
			continue
		}
		err = Pcc.DelUser(user)
		if err != nil {
			assert.Fatalf("Failed to Deluser %v: %v\n", user, err)
			return
		}
	}
}
