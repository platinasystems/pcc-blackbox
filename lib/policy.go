package pcc

import (
	"fmt"
	"github.com/platinasystems/pcc-models/policy"
)

////
// Policy Rest Client
////

// Get all policies
func (pcc *PccClient) GetPolicies() (result []policy.Policy, err error) {
	err = pcc.Get("pccserver/policies", &result)
	return
}

// Get a Policy
func (pcc *PccClient) GetPolicy(id uint64) (result policy.Policy, err error) {
	err = pcc.Get(fmt.Sprintf("pccserver/policies/%d", id), &result)
	return
}

// Add a Policy
func (pcc *PccClient) AddPolicy(item *policy.Policy) (result policy.Policy, err error) {
	err = pcc.Post("pccserver/policies", item, &result)
	return
}

// Update a Policy
func (pcc *PccClient) UpdatePolicy(item *policy.Policy) (result policy.Policy, err error) {
	err = pcc.Put(fmt.Sprintf("pccserver/policies/%d", item.Id), item, &result)
	return
}

// Delete a Policy
func (pcc *PccClient) DeletePolicy(id uint64) (result policy.Policy, err error) {
	err = pcc.Delete(fmt.Sprintf("pccserver/policies/%d", id), nil, &result)
	return
}
