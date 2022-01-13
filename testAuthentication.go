package main

import (
	"errors"
	"github.com/golang-jwt/jwt"
	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	m "github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/pcc-models/security"
	"github.com/platinasystems/test"
	"testing"
	"time"
)

var (
	roles                                map[string]*pcc.SecurityRole
	users                                map[string]*pcc.User
	userRequests                         map[string]*pcc.UserRequest
	roleNames, userNames                 []string
	tenant, tenantOkta, tenantLDAP       *security.Tenant
	authProfileOktaID, authProfileLDAPID uint64
)

func testAuthentication(t *testing.T) {
	t.Run("addRolesAndTenantsPlatina", addRolesAndTenantsPlatina)
	t.Run("addPlatinaUsers", addPlatinaUsers)
	t.Run("checkTenantsScope", checkTenantsScope)
	t.Run("checkRolePermissions", checkRolePermissions)
	t.Run("deleteUsers", deleteUsers)
	t.Run("deleteRolesAndTenants", deleteRolesAndTenants)

	if Env.OktaConfiguration.ApiKey != "" {
		t.Run("addRolesAndTenantsOkta", addRolesAndTenantsOkta)
		t.Run("addOktaBootstrapUser", addOktaBootstrapUser)
		t.Run("addOktaAuthProfiles", addOktaAuthProfiles)
		t.Run("addOktaGroupMapping", addOktaGroupMapping)
		t.Run("checkOktaGroupMapping", checkOktaGroupMapping)
		t.Run("deleteOktaAuthProfile", deleteOktaAuthProfile)
		t.Run("deleteOktaUser", deleteOktaUser)
		t.Run("deleteRolesAndTenantsOkta", deleteRolesAndTenantsOkta)
	}

	if Env.LDAPConfiguration.URL != "" {
		t.Run("addRolesAndTenantsLDAP", addRolesAndTenantsLDAP)
		t.Run("addLDAPBootstrapUser", addLDAPBootstrapUser)
		t.Run("addLDAPAuthProfiles", addLDAPAuthProfiles)
		t.Run("addLDAPGroupMapping", addLDAPGroupMapping)
		t.Run("checkLDAPGroupMapping", checkLDAPGroupMapping)
		t.Run("deleteLDAPAuthProfile", deleteLDAPAuthProfile)
		t.Run("deleteLDAPUser", deleteLDAPUser)
		t.Run("deleteRolesAndTenantsLDAP", deleteRolesAndTenantsLDAP)
	}
}

func addRolesAndTenantsPlatina(t *testing.T) {
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

func addRolesAndTenantsOkta(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var err error

	tenantOkta = &(security.Tenant{Name: "test-tenant-bb-okta", Description: "blackbox test tenant for okta"})
	tenantOkta, err = Pcc.AddTenant(*tenantOkta)
	checkError(t, res, err)

	rwUserManagementGroupOperation := security.GroupOperation{Id: 8}

	reqRoleChildRWOkta := pcc.UserRole{
		GenericModel: pcc.GenericModel{Id: 0,
			Name:  "test-role-okta",
			Owner: tenantOkta.ID},
		GroupOperations: []security.GroupOperation{rwUserManagementGroupOperation},
	}

	roles["role-okta"], err = Pcc.RegisterRole(reqRoleChildRWOkta)
	checkError(t, res, err)
}

func addRolesAndTenantsLDAP(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var err error

	tenantLDAP = &(security.Tenant{Name: "test-tenant-bb-LDAP", Description: "blackbox test tenant for LDAP"})

	tenantLDAP, err = Pcc.AddTenant(*tenantLDAP)
	checkError(t, res, err)

	rwUserManagementGroupOperation := security.GroupOperation{Id: 8}

	reqRoleChildRWLDAP := pcc.UserRole{
		GenericModel: pcc.GenericModel{Id: 0,
			Name:  "test-role-LDAP",
			Owner: tenantLDAP.ID},
		GroupOperations: []security.GroupOperation{rwUserManagementGroupOperation},
	}

	roles["role-LDAP"], err = Pcc.RegisterRole(reqRoleChildRWLDAP)
	checkError(t, res, err)
}

func addOktaBootstrapUser(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var err error

	userRequests["user-bootstrap-okta"] = &pcc.UserRequest{
		UserName:  "user-bootstrap-okta@platinasystems.com",
		FirstName: "Okta",
		LastName:  "Okta",
		Password:  "password-bb",
		TenantId:  tenantOkta.ID,
		RoleId:    roles["role-okta"].Id,
	}

	users["user-bootstrap-okta"], err = Pcc.AddUserReq(*userRequests["user-bootstrap-okta"])
	checkError(t, res, err)
	log.AuctaLogger.Infof("Added user %v", users["user-bootstrap-okta"])
}

func addLDAPBootstrapUser(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var err error
	users = make(map[string]*pcc.User)

	userRequests["user-bootstrap-LDAP"] = &pcc.UserRequest{
		UserName:  "user-bootstrap-LDAP@platinasystems.com",
		FirstName: "LDAP",
		LastName:  "LDAP",
		Password:  "password-bb",
		TenantId:  tenantLDAP.ID,
		RoleId:    roles["role-LDAP"].Id,
	}

	users["user-bootstrap-LDAP"], err = Pcc.AddUserReq(*userRequests["user-bootstrap-LDAP"])
	checkError(t, res, err)
	log.AuctaLogger.Infof("Added user %v", users["user-bootstrap-LDAP"])
}

func addLDAPAuthProfiles(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var err error

	authProfileReqLDAP := &pcc.AuthProfile{
		Provider: "LDAP",
		Parameters: pcc.LDAPConfig{
			URL:         Env.LDAPConfiguration.URL,
			GroupBaseDN: Env.LDAPConfiguration.GroupBaseDN,
		},
	}

	log.AuctaLogger.Infof("LDAP auth profile request: %v", authProfileReqLDAP)

	err = Pcc.ChangeUser(pcc.Credential{UserName: users["user-bootstrap-LDAP"].UserName,
		Password: "password-bb"})
	err = Pcc.AddAuthenticationProfile(authProfileReqLDAP)
	checkError(t, res, err)
	profiles, _ := Pcc.GetAuthenticationProfiles()
	authProfileLDAPID = profiles[0].ID
	log.AuctaLogger.Infof("LDAP auth profile ID: %v", authProfileLDAPID)
}

func addOktaAuthProfiles(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var err error

	authProfileReqOkta := &pcc.AuthProfile{
		Provider: "okta",
		Parameters: pcc.OktaConfig{
			Domain: Env.OktaConfiguration.Domain,
			ApiKey: Env.OktaConfiguration.ApiKey,
		},
	}

	log.AuctaLogger.Infof("Okta auth profile request: %v", authProfileReqOkta)

	err = Pcc.ChangeUser(pcc.Credential{UserName: users["user-bootstrap-okta"].UserName,
		Password: "password-bb"})
	checkError(t, res, err)
	err = Pcc.AddAuthenticationProfile(authProfileReqOkta)
	checkError(t, res, err)
	profiles, _ := Pcc.GetAuthenticationProfiles()
	authProfileOktaID = profiles[0].ID
	log.AuctaLogger.Infof("Okta auth profile ID: %v", authProfileOktaID)
}
func addOktaGroupMapping(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var err error

	group := &pcc.ThirdPartyGroup{
		Group:         Env.OktaConfiguration.Group,
		RoleID:        roles["role-okta"].Id,
		TenantID:      tenantOkta.ID,
		AuthProfileID: authProfileOktaID,
	}

	log.AuctaLogger.Infof("Okta group mapping: %v", group)
	err = Pcc.ChangeUser(pcc.Credential{UserName: users["user-bootstrap-okta"].UserName,
		Password: "password-bb"})
	checkError(t, res, err)

	group, err = Pcc.AddThirdPartyGroup(group)
	checkError(t, res, err)
}

func addLDAPGroupMapping(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var err error

	group := &pcc.ThirdPartyGroup{
		Group:         Env.LDAPConfiguration.Group,
		RoleID:        roles["role-LDAP"].Id,
		TenantID:      tenantLDAP.ID,
		AuthProfileID: authProfileLDAPID,
	}

	log.AuctaLogger.Infof("LDAP group mapping: %v", group)

	err = Pcc.ChangeUser(pcc.Credential{UserName: users["user-bootstrap-LDAP"].UserName,
		Password: "password-bb"})

	group, err = Pcc.AddThirdPartyGroup(group)
	checkError(t, res, err)
}

func checkOktaGroupMapping(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var err error

	err = Pcc.ChangeUser(pcc.Credential{UserName: Env.OktaConfiguration.Username, Password: Env.OktaConfiguration.Password, Tenant: "test-tenant-bb-okta"})
	checkError(t, res, err)

	var token *jwt.Token
	token, err = jwt.ParseWithClaims(Pcc.GetToken(), &pcc.TokenClaims{}, nil)
	claims := token.Claims.(*pcc.TokenClaims)

	if claims.Tenant != tenantOkta.ID {
		err = errors.New("Error assigning tenant id to user")
	}

	if claims.Role != roles["role-okta"].Id {
		err = errors.New("Error assigning role id to user")
	}
}

func checkLDAPGroupMapping(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var err error

	err = Pcc.ChangeUser(pcc.Credential{UserName: Env.LDAPConfiguration.Username, Password: Env.LDAPConfiguration.Password, Tenant: "test-tenant-bb-LDAP"})
	checkError(t, res, err)

	var token *jwt.Token
	token, err = jwt.ParseWithClaims(Pcc.GetToken(), &pcc.TokenClaims{}, nil)
	claims := token.Claims.(*pcc.TokenClaims)

	if claims.Tenant != tenantLDAP.ID {
		err = errors.New("Error assigning tenant id to user")
	}

	if claims.Role != roles["role-LDAP"].Id {
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

	userNames = []string{"user-parent-ro", "user-parent-rw", "user-child-ro", "user-child-rw"}

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

func deleteOktaAuthProfile(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	err := Pcc.ChangeUser(pcc.Credential{UserName: users["user-bootstrap-okta"].UserName,
		Password: "password-bb"})
	checkError(t, res, err)
	err = Pcc.DeleteAuthenticationProfile(authProfileOktaID)
	checkError(t, res, err)

}

func deleteLDAPAuthProfile(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	err := Pcc.ChangeUser(pcc.Credential{UserName: users["user-bootstrap-LDAP"].UserName,
		Password: "password-bb"})
	checkError(t, res, err)
	err = Pcc.DeleteAuthenticationProfile(authProfileLDAPID)
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

func deleteOktaUser(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	err := Pcc.ChangeUser(adminCredential)
	checkError(t, res, err)

	err = Pcc.DelUser(users["user-bootstrap-okta"].UserName)
	checkError(t, res, err)
}

func deleteLDAPUser(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	err := Pcc.ChangeUser(adminCredential)
	checkError(t, res, err)

	err = Pcc.DelUser(users["user-bootstrap-LDAP"].UserName)
	checkError(t, res, err)
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

func deleteRolesAndTenantsOkta(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	err := Pcc.ChangeUser(adminCredential)
	checkError(t, res, err)

	err = Pcc.DeleteRole(roles["role-okta"].Id)
	checkError(t, res, err)
	err = Pcc.DelTenant(tenantOkta.ID)
	checkError(t, res, err)
}

func deleteRolesAndTenantsLDAP(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	err := Pcc.ChangeUser(adminCredential)
	checkError(t, res, err)

	err = Pcc.DeleteRole(roles["role-LDAP"].Id)
	checkError(t, res, err)
	err = Pcc.DelTenant(tenantLDAP.ID)
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
