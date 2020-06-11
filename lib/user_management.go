// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"fmt"
	"github.com/platinasystems/pcc-models/security"
)

// FIXME move in a common module
type GenericModel struct {
	Id          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Owner       uint64 `json:"owner"`
}

func (gm *GenericModel) GetName() string {
	return gm.Name
}

func (gm *GenericModel) GetDescription() string {
	return gm.Description
}

func (gm *GenericModel) GetId() uint64 {
	return gm.Id
}

func (gm *GenericModel) GetOwner() uint64 {
	return gm.Owner
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

type SecurityRole struct {
	GenericModel
	Protect bool `json:"protect"`
}

type Operation struct {
	GenericModel
	GroupId bool `json:"groupID"`
}

// tried re-using model.User, but not sure about
// unmarshal in to type interface
// FIXME import from pcc-models
type User struct {
	Id        uint64          `json:"id"`
	UserName  string          `json:"username"`
	FirstName string          `json:"firstname"`
	LastName  string          `json:"lastname"`
	Email     string          `json:"email"`
	Password  string          `json:"password,omitempty"`
	TenantId  uint64          `json:"tenant"`
	RoleId    uint64          `json:"roleID"`
	Source    string          `json:"source"`
	Active    bool            `json:"active"`
	Protect   bool            `json:"protect"`
	Role      *SecurityRole   `json:"role"`
	Tenant    security.Tenant `json:"tenant"`
	Profile   Profile         `json:"profile"`
}

func (pcc *PccClient) AddTenant(tenant security.Tenant) (t *security.Tenant, err error) {
	endpoint := fmt.Sprintf("user-management/tenant/register")
	err = pcc.Post(endpoint, &tenant, &tenant)
	t = &tenant
	return
}

func (pcc *PccClient) DelTenant(tenantId uint64) (err error) {
	endpoint := fmt.Sprintf("user-management/tenant/%d", tenantId)
	err = pcc.Delete(endpoint, nil, nil)
	return
}

func (pcc *PccClient) GetTenants() (tenants []security.Tenant, err error) {
	endpoint := fmt.Sprintf("user-management/tenant/list")
	err = pcc.Get(endpoint, &tenants)
	return
}

func (pcc *PccClient) GetTenant(id uint64) (tenant *security.Tenant, err error) {
	endpoint := fmt.Sprintf("user-management/tenant/%d", id)
	var t security.Tenant
	err = pcc.Get(endpoint, &t)
	tenant = &t
	return
}

func (pcc *PccClient) GetOperation(id uint64) (result *Operation, err error) {
	var r Operation
	endpoint := fmt.Sprintf("user-management/operation/%d", id)
	err = pcc.Get(endpoint, &r)
	result = &r
	return
}

func (pcc *PccClient) GetGroupOperation(id uint64) (result *Operation, err error) {
	var r Operation
	endpoint := fmt.Sprintf("user-management/groupoperation/%d", id)
	err = pcc.Get(endpoint, &r)
	result = &r
	return
}

func (pcc *PccClient) GetOperations() (operations []Operation, err error) {
	endpoint := fmt.Sprintf("user-management/operation/describe")
	err = pcc.Get(endpoint, &operations)
	return
}

func (pcc *PccClient) GetGroupOperations() (operations []Operation, err error) {
	endpoint := fmt.Sprintf("user-management/groupoperation/describe")
	err = pcc.Get(endpoint, &operations)
	return
}

func (pcc *PccClient) DeleteOperation(id uint64) (err error) {
	endpoint := fmt.Sprintf("user-management/operation/%d", id)
	err = pcc.Delete(endpoint, nil, nil)
	return
}

func (pcc *PccClient) DeleteGroupOperation(id uint64) (err error) {
	endpoint := fmt.Sprintf("user-management/groupoperation/%d", id)
	err = pcc.Delete(endpoint, nil, nil)
	return
}

func (pcc *PccClient) AddRole(name string, description string) (role *GenericModel, err error) {
	endpoint := fmt.Sprintf("user-management/role")
	role = &GenericModel{Name: name, Description: description}
	err = pcc.Post(endpoint, role, role)
	return
}

func (pcc *PccClient) GetSecurityRole(id uint64) (role *GenericModel, err error) {
	var r GenericModel
	endpoint := fmt.Sprintf("user-management/role/%d", id)
	err = pcc.Get(endpoint, &r)
	role = &r
	return
}

func (pcc *PccClient) DeleteRole(id uint64) (err error) {
	endpoint := fmt.Sprintf("user-management/role/%d", id)
	err = pcc.Delete(endpoint, nil, nil)
	return
}

func (pcc *PccClient) GetSecurityRoles() (roles []GenericModel, err error) {
	endpoint := fmt.Sprintf("user-management/role")
	err = pcc.Get(endpoint, &roles)
	return
}

func (pcc *PccClient) AssignTenantNodes(tenantId uint64, nodes []uint64) (err error) {
	endpoint := fmt.Sprintf("user-management/tenant/nodes/update")
	req := ChangeTenant{
		TenantId: tenantId,
		NodeIds:  nodes,
	}
	err = pcc.Post(endpoint, &req, nil)
	return
}

func (pcc *PccClient) FindTenant(tenantName string) (tenant security.Tenant, err error) {
	var tenants []security.Tenant

	if tenants, err = pcc.GetTenants(); err == nil {
		for _, t := range tenants {
			if t.Name == tenantName {
				tenant = t
				return
			}
		}
		err = fmt.Errorf("Couldn't find tenant %v", tenantName)
	}
	return
}

func (pcc *PccClient) GetUsers() (users []User, err error) {
	endpoint := fmt.Sprintf("user-management/user/list")
	err = pcc.Get(endpoint, &users)
	return
}

func (pcc *PccClient) AddUser(user User) (added *User, err error) {
	endpoint := fmt.Sprintf("user-management/user/register")
	err = pcc.Post(endpoint, &user, &user)
	added = &user
	return
}

func (pcc *PccClient) UpdateUser(user User) (err error) {
	endpoint := fmt.Sprintf("user-management/user/update")
	err = pcc.Post(endpoint, &user, &user)
	return
}

func (pcc *PccClient) DelUser(user string) (err error) {
	endpoint := fmt.Sprintf("user-management/user/delete")
	delUser := User{UserName: user}
	err = pcc.Post(endpoint, &delUser, nil)
	return
}

func (pcc *PccClient) SetUserSpace(scope string, content string) (err error) {
	endpoint := fmt.Sprintf("user-management/userspace/%s", scope)
	err = pcc.Post(endpoint, &content, nil)
	return
}

func (pcc *PccClient) GetUserSpace(scope string) (content string, err error) {
	endpoint := fmt.Sprintf("user-management/userspace/%s", scope)
	err = pcc.Get(endpoint, &content)
	return
}

func (pcc *PccClient) DeleteUserSpace(scope string) (err error) {
	endpoint := fmt.Sprintf("user-management/userspace/%s", scope)
	err = pcc.Delete(endpoint, nil, nil)
	return
}

func (pcc *PccClient) AddEntity(name string, description string) (entity *GenericModel, err error) {
	endpoint := fmt.Sprintf("user-management/entity/register")
	gm := GenericModel{Name: name, Description: description}
	err = pcc.Post(endpoint, &gm, &gm)
	entity = &gm
	return
}

func (pcc *PccClient) GetEntities() (entities []GenericModel, err error) {
	endpoint := fmt.Sprintf("user-management/entity/list")
	err = pcc.Get(endpoint, &entities)
	return
}

func (pcc *PccClient) GetEntity(id uint64) (role *GenericModel, err error) {
	var r GenericModel
	endpoint := fmt.Sprintf("user-management/entity/%d", id)
	err = pcc.Get(endpoint, &r)
	role = &r
	return
}
