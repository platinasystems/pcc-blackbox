// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"encoding/json"
	"fmt"

	"github.com/platinasystems/tiles/pccserver/security/model"
)

type Tenant struct {
	model.Tenant
}

type ChangeTenant struct {
	TenantId uint64   `json:"tenant"`
	NodeIds  []uint64 `json:"ids"`
}

type Profile struct {
	Id        uint64 `json:"id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
	Source    string `json:"source"`
}

type Role struct {
	model.Role
}

// tried re-using model.User, but not sure about
// unmarshal in to type interface
type User struct {
	Id       uint64 `json:"id"`
	UserName string `json:"username"`
	Active   bool   `json:"active"`
	Owner    uint64 `json:"owner"`
	Tenant   Tenant `json:"tenant"`
	Protect  bool   `json:"protect"`
	Role     Role   `json:"role"`
	Profile  `json:"profile"`
}

type AddUser struct {
	UserName  string `json:"username"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
	Password  string `json:"password,omitempty"`
	TenantId  uint64 `json:"tenant"`
	RoleId    uint64 `json:"roleID"`
	Source    string `json:"source"`
	Active    bool   `json:"active"`
	Protect   bool   `json:"protect"`
}

type DelUser struct {
	UserName string `json:"username"`
}

func (p PccClient) AddTenant(addReq Tenant) (err error) {
	var data []byte

	endpoint := fmt.Sprintf("user-management/tenant/register")
	if data, err = json.Marshal(addReq); err != nil {
		return
	}
	_, err = p.pccUserManagement("POST", endpoint, data)
	return
}

func (p PccClient) DelTenant(tenantId uint64) (err error) {
	var (
		data   []byte
		delReq Tenant
	)

	endpoint := fmt.Sprintf("user-management/tenant/delete")
	delReq.ID = tenantId
	if data, err = json.Marshal(delReq); err != nil {
		return
	}
	_, err = p.pccUserManagement("POST", endpoint, data)
	return
}

func (p PccClient) GetTenants() (tenants []Tenant, err error) {
	var body []byte

	endpoint := fmt.Sprintf("user-management/tenant/list")
	body, err = p.pccUserManagement("GET", endpoint, nil)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &tenants)
	return
}

func (p PccClient) AssignTenantNodes(tenantId uint64, nodes []uint64) (
	err error) {

	var data []byte

	endpoint := fmt.Sprintf("user-management/tenant/nodes/update")
	Req := ChangeTenant{
		TenantId: tenantId,
		NodeIds:  nodes,
	}
	if data, err = json.Marshal(Req); err != nil {
		return
	}
	_, err = p.pccUserManagement("POST", endpoint, data)
	return
}

func (p PccClient) FindTenant(tenantName string) (tenant Tenant, err error) {
	var tenants []Tenant

	tenants, err = p.GetTenants()
	if err != nil {
		return
	}

	for _, t := range tenants {
		if t.Name == tenantName {
			tenant = t
			return
		}
	}
	err = fmt.Errorf("Couldn't find tenant %v", tenantName)
	return
}

func (p PccClient) GetUsers() (users []User, err error) {
	var body []byte

	endpoint := fmt.Sprintf("user-management/user/list")
	body, err = p.pccUserManagement("GET", endpoint, nil)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &users)
	return
}

func (p PccClient) AddUser(addUser AddUser) (err error) {
	var data []byte

	endpoint := fmt.Sprintf("user-management/user/register")
	if data, err = json.Marshal(addUser); err != nil {
		return
	}
	_, err = p.pccUserManagement("POST", endpoint, data)
	return
}

func (p PccClient) UpdateUser(addUser AddUser) (err error) {
	var data []byte

	endpoint := fmt.Sprintf("user-management/user/update")
	if data, err = json.Marshal(addUser); err != nil {
		return
	}
	_, err = p.pccUserManagement("POST", endpoint, data)
	return
}

func (p PccClient) DelUser(user string) (err error) {
	var data []byte

	endpoint := fmt.Sprintf("user-management/user/delete")
	delUser := DelUser{UserName: user}
	if data, err = json.Marshal(delUser); err != nil {
		return
	}
	_, err = p.pccUserManagement("POST", endpoint, data)
	return
}
