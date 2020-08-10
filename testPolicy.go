package main

import (
	"fmt"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/pcc-models/app"
	"github.com/platinasystems/pcc-models/policy"
	scope2 "github.com/platinasystems/pcc-models/scope"
	"strings"
	"testing"
)

const POLICY_BB_TEST = "blackbox-test-policy"
const SCOPE_BB_TEST = "blackbox-scope-test-policy"
const ROLE_BB_TEST = "blackbox-role-test-policy"

////
// TEST Policies
////

func checkError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func testPreparePolicies(t *testing.T) {
	// Clean policies
	policies, err := Pcc.GetPolicies() // GET the policies
	checkError(t, err)
	for i := range policies { // Delete an old policy
		p := policies[i]
		if p.Description == POLICY_BB_TEST {
			fmt.Println("Deleting the policy", p.Description)
			_, err = Pcc.DeletePolicy(p.Id)
			checkError(t, err)
			break
		}
	}

	// Clean scopes
	scopes, err := Pcc.GetScopes()
	checkError(t, err)
	for i := range scopes {
		scope := scopes[i]
		if scope.Name == SCOPE_BB_TEST {
			fmt.Println("Deleting the scope", scope.Description)
			_, err = Pcc.DeleteScope(scope.ID)
			checkError(t, err)
			break
		}
	}

	// Clean roles
	roles, err := Pcc.GetNodeRoles()
	checkError(t, err)
	for i := range roles {
		role := roles[i]
		if role.Name == ROLE_BB_TEST {
			fmt.Println("Deleting the role", role.Description)
			_, err = Pcc.DeleteNodeRole(role.ID)
			checkError(t, err)
			break
		}
	}
}

//
// Build a policy with inputs and Add it
//
func buildPolicy(t *testing.T) (*policy.Policy, *app.AppConfiguration) {
	var (
		err         error
		p           *policy.Policy
		apps        []app.AppConfiguration
		inputs      []policy.PolicyInput
		application *app.AppConfiguration
	)

	fmt.Println("Getting apps for building policy")
	apps, err = Pcc.GetApps() // GET the apps
	checkError(t, err)

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
		t.Fatal("Unable to find a valid application")
	}
	p = addPolicy(t, application, nil, inputs)
	return p, application
}

// Add the policy to PCC
func addPolicy(t *testing.T, application *app.AppConfiguration, scopes *[]uint64, inputs []policy.PolicyInput) *policy.Policy {
	var err error
	fmt.Println("\n--- ADD POLICY")
	p := policy.Policy{Description: POLICY_BB_TEST, AppId: application.ID, Inputs: inputs}
	fmt.Printf("Adding the policy %s with inputs: %v\n", p.Description, p.Inputs)
	if scopes != nil {
		fmt.Printf("Associate the policy %s to the scopes: %v\n", p.Description, p.ScopeIDs)
		p.ScopeIDs = *scopes
	}
	p, err = Pcc.AddPolicy(&p) // ADD a policy
	checkError(t, err)
	fmt.Printf("Added the policy %d %s\n\n", p.Id, p.Description)
	return &p
}

//
// add, get, update and delete a policy
//
func testPolicies(t *testing.T) {
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
	checkError(t, err)

	////
	// Update the policy
	////
	p.Inputs = p1.Inputs[:len(p1.Inputs)-1] // Remove the last input
	p.Id = p1.Id

	fmt.Println("--- UPDATE POLICY")
	fmt.Printf("Updating the policy %v\n\n", p)
	p2, err = Pcc.UpdatePolicy(p) // Update the policy
	checkError(t, err)

	////
	// Validate the policy
	////
	fmt.Println("--- GET POLICY")
	p2, err = Pcc.GetPolicy(p.Id) // GET and check if update worked
	checkError(t, err)
	if len(p1.Inputs) == len(p2.Inputs) {
		t.Fatal("There was an error updating the policy")
	}
	fmt.Printf("Policy %d %s was correctly updated\n", p.Id, p.Description)
}

//
// Test policy-scope association
//
func testPolicyScope(t *testing.T) {
	defer testPreparePolicies(t)
	var (
		s           scope2.Scope
		err         error
		application app.AppConfiguration
		roleId      uint64
		nodes       []uint64
		inputs      []policy.PolicyInput
	)
	s.Name = SCOPE_BB_TEST
	s.Description = SCOPE_BB_TEST
	s.Type = scope2.LabelTypeName

	////
	// Add the scope
	////
	fmt.Println("\n--- ADD SCOPE")
	fmt.Printf("Creating the scope %s\n", s.Description)
	s, err = Pcc.AddScope(&s)
	checkError(t, err)
	fmt.Printf("Added the scope %d %s\n\n", s.ID, s.Description)

	////
	// Add the policy for LLDPD
	////
	applications, err := Pcc.GetApp(strings.ToLower(pcc.ROLE_LLDPD))
	checkError(t, err)
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
	fmt.Println("\n--- ASSIGN SCOPE")
	nodes, err = Pcc.GetNodeIds()
	checkError(t, err)
	nodeId := nodes[0]
	node, err := Pcc.GetNode(nodeId)
	checkError(t, err)
	node.ScopeId = &s.ID
	fmt.Printf("Assigning the scope %s to the node %s\n", s.Description, node.Name)
	err = Pcc.UpdateNode(node)
	checkError(t, err)

	// Remove the scope
	defer func() {
		node.ScopeId = nil
		err = Pcc.UpdateNode(node)
	}()

	////
	// Assign the default role to the node
	////
	fmt.Println("\n--- ASSIGN THE ROLE")
	fmt.Printf("Removing all roles from the node %d\n", nodeId)
	node.RoleIds = []uint64{}
	err = Pcc.UpdateNode(node)
	checkError(t, err)
	roleId, err = Pcc.FindRoleId(pcc.ROLE_DEFAULT)
	checkError(t, err)
	fmt.Printf("Assigning the role %s with application %s to %s\n", s.Description, application.Name, node.Name)
	err = setRolesToNodesAndCheck([]uint64{roleId}, pcc.ROLE_LLDPD, []uint64{nodeId}, 0)
	checkError(t, err)

	// TODO check for inputs parameters. Look at the default/ansible.log file
}
