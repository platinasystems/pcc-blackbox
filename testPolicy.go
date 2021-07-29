package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/pcc-models/app"
	"github.com/platinasystems/pcc-models/policy"
	scope2 "github.com/platinasystems/pcc-models/scope"
)

const POLICY_BB_TEST = "blackbox-test-policy"
const SCOPE_BB_TEST = "blackbox-scope-test-policy"
const ROLE_BB_TEST = "blackbox-role-test-policy"

////
// TEST Policies
////


func testPreparePolicies(t *testing.T) {
	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var (
		defaultRack *scope2.Scope
		roles       []pcc.Role
		scopes      []scope2.Scope
		err         error
	)

	// Set the default rack scope
	scopes, err = Pcc.GetScopes()
	checkError(t, res, err)
	nodes, err := Pcc.GetNodeIds()
	checkError(t, res, err)
	for i := range scopes {
		s := scopes[i]
		if s.Default && s.IsRack() {
			defaultRack = &s
			break
		}
	}

	for i := range nodes {
		nodeId := nodes[i]
		node, err := Pcc.GetNode(nodeId)
		checkError(t, res, err)
		if *node.ScopeId != defaultRack.ID {
			log.AuctaLogger.Infof("Assign the default rack scope to node %d:%s", nodeId, node.Name)
			node.ScopeId = &defaultRack.ID
			err = Pcc.UpdateNode(node)
			checkError(t, res, err)
		}
	}

	// Clean roles
	roles, err = Pcc.GetNodeRoles()
	checkError(t, res, err)
	for i := range roles {
		role := roles[i]
		if role.Name == ROLE_BB_TEST {
			log.AuctaLogger.Infof("Deleting the role", role.Description)
			_, err = Pcc.DeleteNodeRole(role.ID)
			checkError(t, res, err)
			break
		}
	}

	// Clean scopes
	scopes, err = Pcc.GetScopes()
	checkError(t, res, err)
	for i := range scopes {
		scope := scopes[i]
		if scope.Name == SCOPE_BB_TEST {
			log.AuctaLogger.Infof("Deleting the scope", scope.Description)
			_, err = Pcc.DeleteScope(scope.ID)
			checkError(t, res, err)
			break
		}
	}

	// Clean policies
	policies, err := Pcc.GetPolicies() // GET the policies
	checkError(t, res, err)
	for i := range policies { // Delete an old policy
		p := policies[i]
		if p.Description == POLICY_BB_TEST {
			log.AuctaLogger.Infof("Deleting the policy", p.Description)
			_, err = Pcc.DeletePolicy(p.Id)
			checkError(t, res, err)
			break
		}
	}
}

//
// Build a policy with inputs and Add it
//
func buildPolicy(t *testing.T) (*policy.Policy, *app.AppConfiguration) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var (
		err         error
		p           *policy.Policy
		apps        []app.AppConfiguration
		inputs      []policy.PolicyInput
		application *app.AppConfiguration
	)

	log.AuctaLogger.Infof("Getting apps for building policy")
	apps, err = Pcc.GetApps() // GET the apps
	checkError(t, res, err)

l1:
	for i := range apps { // Build the inputs. Iterate the applications
		app2 := apps[i]
		for j := range app2.Actions {
			action := app2.Actions[j]
			if action.Name == "install" && len(action.Inputs) > 0 {
				for k := range action.Inputs { // Iterate the inputs and copy values from the default
					input := action.Inputs[k]
					if input.Default != nil {
						if def := fmt.Sprintf("%v", input.Default); def != "" {
							inputs = append(inputs, policy.PolicyInput{AppInput: app.AppInput{Name: input.Name, Value: def}})
						}
					}
				}
			}

			if len(inputs) > 0 {
				application = &app2
				break l1
			}
		}
	}

	if application == nil {
		msg := "Unable to find a valid application"
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
	p = addPolicy(t, application, nil, inputs)
	return p, application
}

// Add the policy to PCC
func addPolicy(t *testing.T, application *app.AppConfiguration, scopes *[]uint64, inputs []policy.PolicyInput) *policy.Policy {
	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var err error
	log.AuctaLogger.Infof("--- ADD POLICY")
	p := policy.Policy{Description: POLICY_BB_TEST, AppId: application.ID, Inputs: inputs}
	log.AuctaLogger.Infof("Adding the policy %s with inputs: %v", p.Description, p.Inputs)
	if scopes != nil {
		log.AuctaLogger.Infof("Associate the policy %s to the scopes: %v", p.Description, p.ScopeIDs)
		p.ScopeIDs = *scopes
	}
	p, err = Pcc.AddPolicy(&p) // ADD a policy
	checkError(t, res, err)
	log.AuctaLogger.Infof("Added the policy %d %s", p.Id, p.Description)
	return &p
}

//
// add, get, update and delete a policy
//
func testPolicies(t *testing.T) {
	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	defer testPreparePolicies(t)

	var (
		err    error
		p      *policy.Policy
		p1, p2 policy.Policy
	)

	////
	// Build and add the policy
	////
	p, _ = buildPolicy(t)
	p1, err = Pcc.GetPolicy(p.Id) // GET the policy
	checkError(t, res, err)

	////
	// Update the policy
	////
	p.Inputs = p1.Inputs[:len(p1.Inputs)-1] // Remove the last input
	p.Id = p1.Id

	log.AuctaLogger.Infof("--- UPDATE POLICY")
	log.AuctaLogger.Infof("Updating the policy %v", p)
	p2, err = Pcc.UpdatePolicy(p) // Update the policy
	checkError(t, res, err)

	////
	// Validate the policy
	////
	log.AuctaLogger.Infof("--- GET POLICY")
	p2, err = Pcc.GetPolicy(p.Id) // GET and check if update worked
	checkError(t, res, err)
	if len(p1.Inputs) == len(p2.Inputs) {
		msg := "There was an error updating the policy"
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
	log.AuctaLogger.Infof("Policy %d %s was correctly updated", p.Id, p.Description)
}

//
// Test policy-scope association
//
func testPolicyScope(t *testing.T) {
	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	defer testPreparePolicies(t)
	var (
		s           scope2.Scope
		err         error
		application app.AppConfiguration
		nodes       []uint64
		inputs      []policy.PolicyInput
	)
	s.Name = SCOPE_BB_TEST
	s.Description = SCOPE_BB_TEST
	s.Type = scope2.LabelTypeName

	////
	// Add the scope
	////
	log.AuctaLogger.Info("--- ADD SCOPE")
	log.AuctaLogger.Infof("Creating the scope %s", s.Description)
	s, err = Pcc.AddScope(&s)
	checkError(t, res, err)
	log.AuctaLogger.Infof("Added the scope %d %s", s.ID, s.Description)

	////
	// Add the policy for LLDPD
	////
	applications, err := Pcc.GetApp(strings.ToLower(pcc.ROLE_LLDPD))
	checkError(t, res, err)
	application = applications[0]

	addInput := func(k string, v string) {
		pi := policy.PolicyInput{}
		pi.Name = k
		pi.Value = v
		inputs = append(inputs, pi)
	}
	addInput("fake1", "value1")
	addInput("fake2", "value2")
	addInput("fake3", "value3")
	_ = addPolicy(t, &application, &([]uint64{s.ID}), inputs)

	////
	// Assign the scope to the node
	////
	log.AuctaLogger.Info("--- ASSIGN SCOPE")
	nodes, err = Pcc.GetNodeIds()
	checkError(t, res, err)
	nodeId := nodes[0]
	node, err := Pcc.GetNode(nodeId)
	checkError(t, res, err)
	node.ScopeId = &s.ID
	log.AuctaLogger.Infof("Assigning the scope %s to the node %s", s.Description, node.Name)
	err = Pcc.UpdateNode(node)
	checkError(t, res, err)

	// Remove the scope
	defer func() {
		node.ScopeId = nil
		err = Pcc.UpdateNode(node)
	}()

	////
	// Assign the default role to the node
	////
	log.AuctaLogger.Info("--- ASSIGN THE ROLE")
	log.AuctaLogger.Infof("Removing all roles from the node %d", nodeId)
	node.RoleIds = []uint64{}
	err = Pcc.UpdateNode(node)
	checkError(t, res, err)
	// TODO check for inputs parameters. Look at the default/ansible.log file
}
