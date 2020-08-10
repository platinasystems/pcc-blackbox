package pcc

import (
	"fmt"
	"github.com/platinasystems/pcc-models/policy"
	"github.com/platinasystems/pcc-models/scope"
)

////
// Scope Rest Client
////

// Get all Scopes
func (pcc *PccClient) GetScopes() (result []scope.Scope, err error) {
	err = pcc.Get("pccserver/scopes", &result)
	return
}

// Get a Scope
func (pcc *PccClient) GetScope(id uint64) (result scope.Scope, err error) {
	err = pcc.Get(fmt.Sprintf("pccserver/scopes/%d", id), &result)
	return
}

// Add a Scope
func (pcc *PccClient) AddScope(item *scope.Scope) (result scope.Scope, err error) {
	err = pcc.Post("pccserver/scopes", item, &result)
	return
}

// Update a Scope
func (pcc *PccClient) UpdateScope(item *scope.Scope) (result scope.Scope, err error) {
	err = pcc.Put(fmt.Sprintf("pccserver/scopes/%d", item.ID), item, &result)
	return
}

// Delete a Scope
func (pcc *PccClient) DeleteScope(id uint64) (result scope.Scope, err error) {
	err = pcc.Delete(fmt.Sprintf("pccserver/scopes/%d", id), nil, &result)
	return
}

// Get all Scope's policies
func (pcc *PccClient) GetScopePolicies(id uint64) (result []policy.Policy, err error) {
	err = pcc.Get(fmt.Sprintf("pccserver/scopes/%d/policies", id), &result)
	return
}
