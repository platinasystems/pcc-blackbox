package main

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	m "github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/pcc-models/security"
	"github.com/platinasystems/test"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/exec"
	"testing"
	"time"
)

var (
	ldapReady    bool
	oktaReady    bool
	roles        map[string]*pcc.SecurityRole
	users        map[string]*pcc.User
	userRequests map[string]*pcc.UserRequest
	roleNames    []string
	userNames    []string
	tenant       *security.Tenant
)

func testAuthentication(t *testing.T) {
	t.Run("checkSecurityConfigFile", checkSecurityConfigFile)
	t.Run("addRolesAndTenants", addRolesAndTenants)
	t.Run("checkOktaGroupMapping", checkOktaGroupMapping)
	t.Run("checkLDAPGroupMapping", checkLDAPGroupMapping)
	t.Run("addPlatinaUsers", addPlatinaUsers)
	t.Run("checkTenantsScope", checkTenantsScope)
	t.Run("checkRolePermissions", checkRolePermissions)
	t.Run("deleteUsers", deleteUsers)
	t.Run("deleteRolesAndTenants", deleteRolesAndTenants)

}
func checkSecurityConfigFile(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var config pcc.SecurityConfig

	ldapReady = true
	oktaReady = true

	cmd := exec.Command("docker", "cp", "security:/home/conf/application.yml", ".")
	err := cmd.Run()
	checkError(t, res, err)

	appYaml, _ := ioutil.ReadFile("application.yml")

	cmd = exec.Command("rm", "application.yml")
	err = cmd.Run()
	checkError(t, res, err)

	err = yaml.Unmarshal(appYaml, &config)
	checkError(t, res, err)

	if config.Auth.Service.Okta.Token == "" || config.Auth.Service.Okta.Domain == "" {
		oktaReady = false
		log.AuctaLogger.Warn("The security service is not configured for okta")
	}

	if config.Auth.Service.LDAP.Url == "" {
		oktaReady = false
		log.AuctaLogger.Warn("The security service is not configured for LDAP")
	}
}

func addRolesAndTenants(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var err error
	roles = make(map[string]*pcc.SecurityRole)
	roleNames = []string{"role-parent-ro", "role-parent-rw", "role-child-ro", "role-child-rw"}

	name := "test-tenant-bb"
	tenant = &(security.Tenant{Name: name, Description: "blackbox test tenant"})

	tenant, err = Pcc.AddTenant(*tenant)
	checkError(t, res, err)

	roUserManagementGroupOperation := security.GroupOperation{Id: 7}
	rwUserManagementGroupOperation := security.GroupOperation{Id: 8}

	reqRoleParentRO := pcc.UserRole{
		GenericModel: pcc.GenericModel{Id: 0,
			Name:  "test-role-ro-parent-bb",
			Owner: 1},
		GroupOperations: []security.GroupOperation{roUserManagementGroupOperation},
	}

	reqRoleParentRW := pcc.UserRole{
		GenericModel: pcc.GenericModel{Id: 0,
			Name:  "test-role-rw-parent-bb",
			Owner: 1},
		GroupOperations: []security.GroupOperation{rwUserManagementGroupOperation},
	}

	roles["role-parent-ro"], err = Pcc.RegisterRole(reqRoleParentRO)
	checkError(t, res, err)
	roles["role-parent-rw"], err = Pcc.RegisterRole(reqRoleParentRW)
	checkError(t, res, err)

	reqRoleChildRO := pcc.UserRole{
		GenericModel: pcc.GenericModel{Id: 0,
			Name:  "test-role-ro-child-bb",
			Owner: tenant.ID},
		GroupOperations: []security.GroupOperation{roUserManagementGroupOperation},
	}

	reqRoleChildRW := pcc.UserRole{
		GenericModel: pcc.GenericModel{Id: 0,
			Name:  "test-role-rw-child-bb",
			Owner: tenant.ID},
		GroupOperations: []security.GroupOperation{rwUserManagementGroupOperation},
	}

	roles["role-child-ro"], err = Pcc.RegisterRole(reqRoleChildRO)
	checkError(t, res, err)
	roles["role-child-rw"], err = Pcc.RegisterRole(reqRoleChildRW)
	checkError(t, res, err)
}

func checkOktaGroupMapping(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res, Env.CheckOktaAuthConfiguration)

	if !oktaReady {
		t.SkipNow()
	}

	var err error
	group := &pcc.ThirdPartyGroup{
		Group:    Env.AuthConfiguration.OktaGroup,
		RoleID:   roles["role-child-rw"].Id,
		TenantID: tenant.ID,
		Provider: "okta",
		Owner:    tenant.ID,
	}

	group, err = Pcc.AddThirdPartyGroup(group)
	checkError(t, res, err)

	log.AuctaLogger.Infof("Successfully added third party group association %v", *group)

	err = Pcc.ChangeUser(pcc.Credential{UserName: Env.AuthConfiguration.OktaUsername, Password: Env.AuthConfiguration.OktaPassword, Provider: "okta"})
	checkError(t, res, err)

	var token *jwt.Token
	token, err = jwt.ParseWithClaims(Pcc.GetToken(), &pcc.TokenClaims{}, nil)
	claims := token.Claims.(*pcc.TokenClaims)

	if claims.Tenant != tenant.ID {
		err = errors.New("Error assigning tenant id to user")
	}

	if claims.Role != roles["role-child-rw"].Id {
		err = errors.New("Error assigning role id to user")
	}
}

func checkLDAPGroupMapping(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res, Env.CheckLDAPAuthConfiguration)

	if !ldapReady {
		t.SkipNow()
	}

	err := Pcc.ChangeUser(adminCredential)
	checkError(t, res, err)

	group := &pcc.ThirdPartyGroup{
		Group:    Env.AuthConfiguration.LDAPGroup,
		RoleID:   roles["role-child-rw"].Id,
		TenantID: tenant.ID,
		Provider: "ldap",
		Owner:    tenant.ID,
	}

	group, err = Pcc.AddThirdPartyGroup(group)
	checkError(t, res, err)

	log.AuctaLogger.Infof("Successfully added third party group association %v", *group)

	err = Pcc.ChangeUser(pcc.Credential{UserName: Env.AuthConfiguration.LDAPUsername, Password: Env.AuthConfiguration.LDAPPassword, Provider: "ldap"})
	checkError(t, res, err)

	var token *jwt.Token
	token, err = jwt.ParseWithClaims(Pcc.GetToken(), &pcc.TokenClaims{}, nil)
	claims := token.Claims.(*pcc.TokenClaims)

	if claims.Tenant != tenant.ID {
		err = errors.New("Error assigning tenant id to user")
	}

	if claims.Role != roles["role-child-rw"].Id {
		err = errors.New("Error assigning role id to user")
	}
}

func addPlatinaUsers(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	err := Pcc.ChangeUser(adminCredential)
	checkError(t, res, err)

	users = make(map[string]*pcc.User)
	userRequests = make(map[string]*pcc.UserRequest)
	userNames = []string{"user-parent-ro", "user-parent-rw", "user-child-ro", "user-child-rw"}

	userRequests["user-parent-ro"] = &pcc.UserRequest{
		UserName:  "user.parent.ro@platinasystems.com",
		FirstName: "a",
		LastName:  "a",
		Password:  "password-bb",
		TenantId:  1,
		RoleId:    roles["role-parent-ro"].Id,
	}
	userRequests["user-parent-rw"] = &pcc.UserRequest{
		UserName:  "user.parent.rw@platinasystems.com",
		FirstName: "b",
		LastName:  "b",
		Password:  "password-bb",
		TenantId:  1,
		RoleId:    roles["role-parent-rw"].Id,
	}

	userRequests["user-child-ro"] = &pcc.UserRequest{
		UserName:  "user.child.ro@platinasystems.com",
		FirstName: "c",
		LastName:  "c",
		Password:  "password-bb",
		TenantId:  tenant.ID,
		RoleId:    roles["role-child-ro"].Id,
	}
	userRequests["user-child-rw"] = &pcc.UserRequest{
		UserName:  "user.child.rw@platinasystems.com",
		FirstName: "d",
		LastName:  "d",
		Password:  "password-bb",
		TenantId:  tenant.ID,
		RoleId:    roles["role-child-rw"].Id,
	}

	for _, username := range userNames {
		users[username], err = Pcc.AddUserReq(*userRequests[username])
		checkError(t, res, err)
		log.AuctaLogger.Infof("Added user %v", users[username])
	}
}

func checkTenantsScope(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var err error
	err = Pcc.ChangeUser(pcc.Credential{UserName: users["user-parent-rw"].UserName, Password: "password-bb"})
	checkError(t, res, err)

	visibleRoles, err := Pcc.ListRoles()

	log.AuctaLogger.Info(visibleRoles)
	if !isVisible(visibleRoles, "test-role-ro-child-bb") ||
		!isVisible(visibleRoles, "test-role-rw-child-bb") ||
		!isVisible(visibleRoles, "test-role-ro-parent-bb") ||
		!isVisible(visibleRoles, "test-role-rw-parent-bb") {
		err = errors.New("Parent tenant could not read all the expected tenants")
		checkError(t, res, err)
	}

	err = Pcc.ChangeUser(pcc.Credential{UserName: users["user-child-rw"].UserName, Password: "password-bb"})
	checkError(t, res, err)

	visibleRoles, err = Pcc.ListRoles()

	if isVisible(visibleRoles, "test-role-ro-parent-bb") ||
		isVisible(visibleRoles, "test-role-rw-parent-bb") {
		err = errors.New("Child tenant should not be able to see parent resources")
		checkError(t, res, err)
	}
}

func checkRolePermissions(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	err := Pcc.ChangeUser(pcc.Credential{UserName: users["user-child-ro"].UserName, Password: "password-bb"})
	checkError(t, res, err)

	roUserManagementGroupOperation := security.GroupOperation{Id: 7}
	rwUserManagementGroupOperation := security.GroupOperation{Id: 8}

	reqRoleChildRW := pcc.UserRole{
		GenericModel: pcc.GenericModel{
			Id:    roles["role-child-rw"].Id,
			Name:  "test-role-rw-child-bb",
			Owner: tenant.ID},
		GroupOperations: []security.GroupOperation{roUserManagementGroupOperation},
	}

	_, err = Pcc.UpdateRole(reqRoleChildRW)
	if err == nil {
		err = errors.New("You should not be able to edit this resource!")
		checkError(t, res, err)
	} else {
		log.AuctaLogger.Info("Expected outcome: %s", err.Error())
	}

	err = Pcc.ChangeUser(pcc.Credential{UserName: users["user-child-rw"].UserName, Password: "password-bb"})
	checkError(t, res, err)

	reqRoleChildRO := pcc.UserRole{
		GenericModel: pcc.GenericModel{
			Id:    roles["role-child-ro"].Id,
			Name:  "test-role-ro-child-bb",
			Owner: tenant.ID},
		GroupOperations: []security.GroupOperation{rwUserManagementGroupOperation},
	}

	_, err = Pcc.UpdateRole(reqRoleChildRO)
	checkError(t, res, err)
}
func deleteUsers(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	err := Pcc.ChangeUser(adminCredential)
	checkError(t, res, err)

	for _, username := range userNames {
		err = Pcc.DelUser(users[username].UserName)
		checkError(t, res, err)
	}
}

func deleteRolesAndTenants(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	err := Pcc.ChangeUser(adminCredential)
	checkError(t, res, err)

	for _, roleName := range roleNames {
		err = Pcc.DeleteRole(roles[roleName].Id)
		checkError(t, res, err)
	}

	err = Pcc.DelTenant(tenant.ID)
	checkError(t, res, err)
}

func isVisible(visibleRoles []*pcc.SecurityRole, roleName string) bool {
	for _, role := range visibleRoles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}
