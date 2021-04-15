package main

import (
	"testing"
)

func GetNameToTestFunc() (nameToTestFunc map[string]func(*testing.T)) {
	nameToTestFunc = map[string]func(*testing.T){
		"getNodes":                                  getNodes,
		"getSecKeys":                                getSecKeys,
		"updateSecurityKey_MaaS":                    updateSecurityKey_MaaS,
		"addClusterHeads":                           addClusterHeads,
		"addBrownfieldServers":                      addBrownfieldServers,
		"configServerInterfaces":                    configServerInterfaces,
		"updateBmcInfo":                             updateBmcInfo,
		"updateIpam":                                updateIpam,
		"addNetCluster":                             addNetCluster,
		"updateNodes_installMAAS":                   updateNodes_installMAAS,
		"reimageAllBrownNodes":                      reimageAllBrownNodes,
		"addTenant":                                 addTenant,
		"createK8sCluster":                          createK8sCluster,
		"deleteK8sCluster":                          deleteK8sCluster,
		"testCeph":                                  testCeph,
		"testCephCacheSetup":                        testCephCacheSetup,
		"testCephCacheAdd":                          testCephCacheAdd,
		"testCephCacheDelete":                       testCephCacheDelete,
		"testK8sApp":                                testK8sApp,
		"UploadSecurityAuthProfileCert":             UploadSecurityAuthProfileCert,
		"AddAuthenticationProfile":                  AddAuthenticationProfile,
		"UploadSecurityPortusKey":                   UploadSecurityPortusKey,
		"AddPortus":                                 AddPortus,
		"CheckPortusInstallation":                   CheckPortusInstallation,
		"UploadSecurityPortusCert":                  UploadSecurityPortusCert,
		"getAvailableNodes":                         getAvailableNodes,
		"delAllPortus":                              delAllPortus,
		"testHardwareInventory":                     testHardwareInventory,
		"updateNodes_installLLDP":                   updateNodes_installLLDP,
		"delAllNetsCluster":                         delAllNetsCluster,
		"delAllNodes":                               delAllNodes,
		"delAllUsers":                               delAllUsers,
		"delAllTenants":                             delAllTenants,
		"delAllProfiles":                            delAllProfiles,
		"delAllCerts":                               delAllCerts,
		"delAllKeys":                                delAllKeys,
		"delAllIpams":                               delAllIpams,
		"checkInvaderTunnels":                       checkInvaderTunnels,
		"checkServerTunnels":                        checkServerTunnels,
		"checkTunnelConnection":                     checkTunnelConnection,
		"checkTunnelForwardingRules":                checkTunnelForwardingRules,
		"checkTunnelRestore":                        checkTunnelRestore,
		"testPreparePolicies":                       testPreparePolicies,
		"testPolicies":                              testPolicies,
		"testPolicyScope":                           testPolicyScope,
		"addUnreachableNode":                        addUnreachableNode,
		"addInaccessibleNode":                       addInaccessibleNode,
		"checkAgentAndCollectorRestore":             checkAgentAndCollectorRestore,
		"addGreenfieldServers":                      addGreenfieldServers,
		"configNetworkInterfaces":                   configNetworkInterfaces,
		"testGetTopic":                              testGetTopic,
		"testGetTopicSchema":                        testGetTopicSchema,
		"testMonitorSample":                         testMonitorSample,
		"testMonitorHistory":                        testMonitorHistory,
		"testUMRole":                                testUMRole,
		"testUMTenant":                              testUMTenant,
		"testUMUser":                                testUMUser,
		"testUMOperation":                           testUMOperation,
		"testUMEntity":                              testUMEntity,
		"testUMUserSpace":                           testUMUserSpace,
		"testKMKeys":                                testKMKeys,
		"testKMCertificates":                        testKMCertificates,
		"testCreateCredendialMetadataProfile":       testCreateCredendialMetadataProfile,
		"testUpdateCredendialMetadataProfile":       testUpdateCredendialMetadataProfile,
		"testDeleteCredendialMetadataProfile":       testDeleteCredendialMetadataProfile,
		"testDashboardGetAllPCCObjects":             testDashboardGetAllPCCObjects,
		"testDashboardGetPCCObjectByRandomId":       testDashboardGetPCCObjectByRandomId,
		"testDashboardGetPCCObjectById":             testDashboardGetPCCObjectById,
		"testDashboardGetChildrenObjectsByRandomId": testDashboardGetChildrenObjectsByRandomId,
		"testDashboardGetParentObjectsByRandomId":   testDashboardGetParentObjectsByRandomId,
		"testDashboardGetFilteredObjects":           testDashboardGetFilteredObjects,
		"testDashboardGetAdvSearchedObjects":        testDashboardGetAdvSearchedObjects,
		"testDashboardGetAggrHealthCountByType":     testDashboardGetAggrHealthCountByType,
		"testDashboardGetMetadataEnumStrings":       testDashboardGetMetadataEnumStrings,
		"addPrivatePublicCert":                      addPrivatePublicCert,
		"testRGW":                                   testRGW,
	}
	return
}

func GetDefaultTestMap() (defaultTests map[string][]string) {
	defaultTests = map[string][]string{
		"TestNodes":                   {"getNodes", "getSecKeys", "updateSecurityKey_MaaS", "addClusterHeads", "addBrownfieldServers", "configServerInterfaces", "updateBmcInfo"},
		"TestUsers":                   {"addTenant"},
		"TestMaaS":                    {"getNodes", "getSecKeys", "updateSecurityKey_MaaS", "updateNodes_installMAAS", "reimageAllBrownNodes"},
		"TestTenantMaaS":              {"getNodes", "getSecKeys", "updateSecurityKey_MaaS", "updateNodes_installMAAS", "addTenant", "reimageAllBrownNodes"},
		"TestAddK8s":                  {"getNodes", "updateIpam", "addNetCluster", "createK8sCluster"},
		"TestDeleteK8s":               {"deleteK8sCluster"},
		"TestNetCluster":              {"getNodes", "updateIpam", "addNetCluster"},
		"TestAddCeph":                 {"getNodes", "updateIpam", "addNetCluster", "testCeph"},
		"TestDeleteCeph":              {"getNodes", "testDeleteCeph"},
		"TestCephCache":               {"testCephCacheSetup", "testCephCacheAdd", "testCephCacheDelete"},
		"TestRGW":                     {"addPrivatePublicCert", "testRGW"},
		"TestK8sApp":                  {"getNodes", "updateIpam", "addNetCluster", "testK8sApp"},
		"TestPortus":                  {"getNodes", "addBrownfieldServers", "UploadSecurityAuthProfileCert", "AddAuthenticationProfile", "UploadSecurityPortusKey", "UploadSecurityPortusCert", "AddPortus", "CheckPortusInstallation"},
		"TestDelPortus":               {"getAvailableNodes", "delAllPortus"},
		"TestHardwareInventory":       {"getNodes", "testHardwareInventory"},
		"TestFull":                    {"getNodes", "getSecKeys", "updateSecurityKey_MaaS", "addClusterHeads", "addBrownfieldServers", "updateNodes_installLLDP", "updateNodes_installMAAS", "configServerInterfaces", "reimageAllBrownNodes", "addTenant", "reimageAllBrownNodes", "createK8sCluster"},
		"TestClean":                   {"getAvailableNodes", "deleteK8sCluster", "delAllNetsCluster", "delAllPortus", "delAllNodes", "delAllUsers", "delAllTenants", "delAllProfiles", "delAllCerts", "delAllKeys", "delAllIpams"},
		"TestTunnel":                  {"getNodes", "addClusterHeads", "addBrownfieldServers", "checkInvaderTunnels", "checkServerTunnels", "checkTunnelConnection", "checkTunnelForwardingRules", "checkTunnelRestore"},
		"TestPolicy":                  {"getNodes", "addClusterHeads", "testPreparePolicies", "testPolicies", "testPolicyScope"},
		"TestAvailability":            {"getNodes", "addClusterHeads", "addUnreachableNode", "addInaccessibleNode", "checkAgentAndCollectorRestore"},
		"TestGreenfield":              {"getNodes", "updateSecurityKey_MaaS", "installMAAS", "addGreenfieldServers", "getNodes", "configNetworkInterfaces", "reimageAllBrownNodes"},
		"TestConfigNetworkInterfaces": {"getNodes", "configNetworkInterfaces"},
		"TestMonitor":                 {"getNodes", "addClusterHeads", "testGetTopic", "testGetTopicSchema", "testMonitorSample", "testMonitorHistory"},
		"TestUserManagement":          {"testUMRole", "testUMTenant", "testUMUser", "testUMOperation", "testUMEntity", "testUMUserSpace"},
		"TestKeyManager":              {"testKMKeys", "testKMCertificates"},
		"TestAppCredentials":          {"testCreateCredendialMetadataProfile", "testUpdateCredendialMetadataProfile", "testDeleteCredendialMetadataProfile"},
		"TestDashboard":               {"testDashboardGetAllPCCObjects", "testDashboardGetPCCObjectByRandomId", "testDashboardGetPCCObjectById", "testDashboardGetChildrenObjectsByRandomId", "testDashboardGetParentObjectsByRandomId", "testDashboardGetFilteredObjects", "testDashboardGetAdvSearchedObjects", "testDashboardGetAggrHealthCountByType", "testDashboardGetMetadataEnumStrings"},
	}
	return
}
