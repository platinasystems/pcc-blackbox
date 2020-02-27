package main

import (
	"fmt"
	"github.com/platinasystems/test"
	"testing"
	"github.com/lib/pq"
	"time"
	"github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/tiles/pccserver/models"
)

var startTime time.Time

func testCeph(t *testing.T) {
	if t.Run("parseCephConfig", parseCephConfig) {
		if run, ok := pcc.GetCephConfig().Tests[pcc.TestCreateCephCluster]; ok && run {
			if t.Run("createCephCluster", testCreateCephCluster) {
				t.Run("verifyCephInstallation", testVerifyCephInstallation)
			}
		} else {
			fmt.Println("Ceph Cluster creation test is skipped")
		}
		if run, ok := pcc.GetCephConfig().Tests[pcc.TestCreateCephPools]; ok && run {
			if t.Run("createCephPool", testCreateCephPool) {
				t.Run("verifyCephPoolCreation", testVerifyCephPoolCreation)
			}
		} else {
			fmt.Println("Ceph Pools creation test is skipped")
		}
		if run, ok := pcc.GetCephConfig().Tests[pcc.TestCreateCephFS]; ok && run {
			if t.Run("createCephFS", testCreateCephFS) {
				t.Run("verifyCephFSCreation", testVerifyCephFSCreation)
			}
		} else {
			fmt.Println("Ceph FS creation test is skipped")
		}
		if run, ok := pcc.GetCephConfig().Tests[pcc.TestDeleteCephFS]; ok && run {
			if t.Run("deleteCephFS", testDeleteCephFS) {
				t.Run("verifyCephFSDeletion", testVerifyCephFSDeletion)
			}
		} else {
			fmt.Println("Ceph FS deletion test is skipped")
		}
		if run, ok := pcc.GetCephConfig().Tests[pcc.TestDeleteCephPools]; ok && run {
			if t.Run("deleteCephPool", testDeleteCephPool) {
				t.Run("verifyCephPoolDeletion", testVerifyCephPoolDeletion)
			}
		} else {
			fmt.Println("Ceph Pools deletion test is skipped")
		}
		if run, ok := pcc.GetCephConfig().Tests[pcc.TestDeleteCephCluster]; ok && run {
			if t.Run("deleteCephCluster", testDeleteCephCluster) {
				t.Run("verifyCephUninstallation", testVerifyCephUninstallation)
			}
		} else {
			fmt.Println("Ceph Cluster deletion test is skipped")
		}
	}
}

func parseCephConfig(t *testing.T){
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	var(
		identifier string
		err error
	)
	if len(Env.Invaders) > 0 {
		identifier = Env.Invaders[0].HostIp
	} else if len(Env.Servers) > 0 {
		identifier = Env.Servers[0].HostIp
	}
	if identifier != "" {
		if err = pcc.GetPccClient().SetCephConfig(Env.CephConfiguration, identifier); err != nil {
			err = fmt.Errorf("Failed to set Ceph Test config..ERROR:%v", err)
		}
	} else {
		err = fmt.Errorf("No unique identifier found")
	}
	if err != nil {
		assert.Fatalf("%v", err)
	}
}

func testCreateCephCluster(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	startTime = time.Now()
	err := createCephCluster()
	if err != nil {
		errMsg := fmt.Sprintf("%v", err)
		fmt.Println(errMsg)
		assert.Fatalf(errMsg)
		return
	}
}

func createCephCluster() (err error) {
	var (
		createRequest     pcc.CreateCephClusterRequest
		clusterId	  uint64
	)
	fmt.Println("Ceph cluster installation is starting")
	if createRequest, err = getCephCreateClusterRequest(); err == nil {
		//TODO: Delete existing ceph cluster with same name if any
		clusterId, err = pcc.GetPccClient().CreateCephCluster(createRequest)
		if err != nil {
			errMsg := fmt.Sprintf("Ceph cluster deployment failed..ERROR:%v", err)
			fmt.Println(errMsg)
			err = fmt.Errorf(errMsg)
			if clusterId != 0 {
				pcc.GetPccClient().SetCephClusterId(clusterId)
			}
		}else {
			pcc.GetPccClient().SetCephClusterId(clusterId)
			fmt.Println("Ceph cluster installation has started. Cluster id:", clusterId)
		}
	}
	return
}

func getCephCreateClusterRequest() (createRequest pcc.CreateCephClusterRequest, err error){
	var sNodes []pcc.CephNodes
	if sNodes, err = getNodesList(pcc.GetCephConfig().NumberOfNodes); err != nil {
		return
	}
	createRequest = pcc.CreateCephClusterRequest{
		Name:       pcc.GetPccClient().GetCephClusterName(),
		Nodes:      sNodes,
		Tags: 	    pq.StringArray{"ROTATIONAL","SATA","PCIe"},
	}
	createRequest.PublicNetwork = pcc.GetCephConfig().PublicNetwork
	createRequest.ClusterNetwork = pcc.GetCephConfig().ClusterNetwork
	createRequest.ControlCIDR = pcc.GetCephConfig().ControlCIDR
	createRequest.IgwPolicy = pcc.GetCephConfig().IgwPolicy
	if createRequest.ClusterNetwork == "" || createRequest.PublicNetwork == "" {
		err = fmt.Errorf("Invalid Public or Cluster Network")
	}
	return
}

func testCreateCephPool(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	startTime = time.Now()
	err := createCephPool()
	if err != nil {
		errMsg := fmt.Sprintf("%v", err)
		fmt.Println(errMsg)
		assert.Fatalf(errMsg)
		return
	}
}

func createCephPool() (errAggr error) {
	var (
		poolId	uint64
	)
	if id := pcc.GetPccClient().GetCephClusterId(); id != 0 {
		for t, pools := range pcc.CephPools {
			for pool, _ := range pools {
				if createRequest, err := getCephPoolCreateRequest(pool, id); err == nil {
					fmt.Printf("Ceph pool [%v] creation is starting\n", pool)
					poolId, err = pcc.GetPccClient().CreateCephPool(createRequest)
					if err != nil {
						errMsg := fmt.Sprintf("Ceph pool [%v] creation failed..ERROR:%v", pool, err)
						fmt.Println(errMsg)
						err = fmt.Errorf(errMsg)
						errAggr = fmt.Errorf(fmt.Sprint(errAggr) + fmt.Sprintf("%v\n", err))
						if poolId != 0 {
							pcc.CephPools[t][pool] = poolId
						}
					} else {
						pcc.CephPools[t][pool] = poolId
						fmt.Printf("Ceph pool [%v] creation has started. poolId: %v\n", pool, poolId)
					}
				}
			}
		}
	}else {
		errAggr = fmt.Errorf("No Ceph Cluster found")
	}
	return
}

func getCephPoolCreateRequest(name string, clusterId uint64) (createRequest pcc.CreateCephPoolRequest, err error) {
	createRequest = pcc.CreateCephPoolRequest{
		Name: name,
		CephClusterId: clusterId,
		Size: 2,
		Quota: 10000,
		QuotaUnit: "MiB",
		PoolType: models.CEPH_POOL_PROFILE_TYPE_REPLICATED.String(),
	}
	return
}

func testCreateCephFS(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	startTime = time.Now()
	err := createCephFS()
	if err != nil {
		errMsg := fmt.Sprintf("%v", err)
		fmt.Println(errMsg)
		assert.Fatalf(errMsg)
		return
	}
}

func createCephFS() (err error) {
	var(
		fsId uint64
	)
	fmt.Printf("Ceph FS [%v] creation is starting\n", pcc.CEPH_FS_NAME)
	if id := pcc.GetPccClient().GetCephClusterId(); id != 0 {
		if createRequest, err := getCephFSCreateRequest(pcc.CEPH_FS_NAME, id); err == nil {
			fsId, err = pcc.GetPccClient().CreateCephFS(createRequest)
			if err != nil {
				errMsg := fmt.Sprintf("Ceph FS [%v] creation failed..ERROR:%v", pcc.CEPH_FS_NAME, err)
				fmt.Println(errMsg)
				err = fmt.Errorf(errMsg)
			} else {
				fmt.Printf("Ceph FS [%v] creation has started. fsId: %v\n", pcc.CEPH_FS_NAME, fsId)
			}
		}
	}else {
		err = fmt.Errorf("No Ceph Cluster found")
	}
	return
}

func getCephFSCreateRequest(name string, clusterId uint64) (createRequest pcc.CreateCephFSRequest, err error) {
	var metadataPool, defaultPool, dataPool1, dataPool2 *models.CephPool
	metadataPool, err = pcc.GetPccClient().GetCephPool(pcc.CEPH_POOL_METADATA)
	if err != nil {
		return
	}
	defaultPool, err = pcc.GetPccClient().GetCephPool(pcc.CEPH_POOL_DEFAULT)
	if err != nil {
		return
	}
	dataPool1, err = pcc.GetPccClient().GetCephPool(pcc.CEPH_POOL_DATA_1)
	if err != nil {
		return
	}
	dataPool2, err = pcc.GetPccClient().GetCephPool(pcc.CEPH_POOL_DATA_2)
	if err != nil {
		return
	}
	createRequest = pcc.CreateCephFSRequest{
		Name: name,
		CephClusterId: clusterId,
		MetadataPool: metadataPool,
		DefaultPool: defaultPool,
		DataPools: []*models.CephPool{dataPool1, dataPool2},
	}
	return
}

func testDeleteCephFS(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	startTime = time.Now()
	err := deleteCephFS()
	if err != nil {
		errMsg := fmt.Sprintf("%v", err)
		fmt.Println(errMsg)
		assert.Fatalf(errMsg)
		return
	}
}

func deleteCephFS() (err error) {
	fmt.Printf("Ceph FS [%v] deletion is starting\n", pcc.CEPH_FS_NAME)
	time.Sleep(time.Second * 5)

	if id := pcc.GetPccClient().GetCephClusterId(); id != 0 {
		cephFS, errGet := pcc.GetPccClient().GetCephFS(pcc.CEPH_FS_NAME)
		if errGet != nil {
			err = fmt.Errorf("%v", errGet)
		}else {
			fsId := cephFS.Id
			if fsId != 0 {
				err = pcc.GetPccClient().DeleteCephFS(cephFS.Id)
				if err != nil {
					err = fmt.Errorf("Ceph FS [%v] deletion failed..ERROR: %v", pcc.CEPH_FS_NAME, err)
				} else {
					fmt.Printf("Ceph FS [%v] deletion has started. fsId: %v\n", pcc.CEPH_FS_NAME, fsId)
				}
			} else {
				err = fmt.Errorf("Ceph FS [%v] not found to delete", fsId)
			}
		}
	}else {
		err = fmt.Errorf("No Ceph Cluster found")
	}
	return
}

func testDeleteCephPool(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	startTime = time.Now()
	err := deleteCephPool()
	if err != nil {
		errMsg := fmt.Sprintf("%v", err)
		fmt.Println(errMsg)
		assert.Fatalf(errMsg)
		return
	}
}

func deleteCephPool() (errAggr error) {
	fmt.Println("Ceph pools deletion is starting")
	time.Sleep(time.Second * 5)

	if id := pcc.GetPccClient().GetCephClusterId(); id != 0 {
		for _, pools := range pcc.CephPools {
			for pool, id := range pools {
				fmt.Printf("Ceph pool [%v] deletion is starting\n", pool)
				err := pcc.GetPccClient().DeleteCephPool(id)
				if err != nil {
					err = fmt.Errorf("Ceph pool [%v] deletion failed..ERROR: %v", pool, err)
					errAggr = fmt.Errorf(fmt.Sprint(errAggr) + fmt.Sprintf("%v\n", err))
				} else {
					fmt.Printf("Ceph pool [%v] deletion has started. poolId: %v\n", pool, id)
				}
			}
		}
	}else {
		errAggr = fmt.Errorf("No Ceph Cluster found")
	}
	return
}

func testDeleteCephCluster(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	startTime = time.Now()
	err := deleteCephCluster()
	if err != nil {
		errMsg := fmt.Sprintf("%v", err)
		fmt.Println(errMsg)
		assert.Fatalf(errMsg)
		return
	}
}

func deleteCephCluster() (err error){
	fmt.Printf("Ceph cluster [%v] uninstallation is starting\n", pcc.GetPccClient().GetCephClusterName())
	time.Sleep(time.Second * 5)

	if id := pcc.GetPccClient().GetCephClusterId(); id != 0 {
		err = pcc.GetPccClient().DeleteCephCluster(id)
		if err != nil {
			err = fmt.Errorf("Ceph cluster deletion failed..ERROR: %v", err)
		} else {
			fmt.Println("Ceph cluster uninstallation has started. Cluster id:", id)
		}
	}else {
		err = fmt.Errorf("No Ceph Cluster found")
	}
	return
}

func testVerifyCephFSCreation(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	fmt.Printf("Verifying Ceph FS [%v] creation...Timeout:[%v sec]\n", pcc.CEPH_FS_NAME, pcc.CEPH_FS_CREATION_TIMEOUT)

	s, err := pcc.GetPccClient().VerifyCeph(startTime, pcc.CEPH_FS_CREATE_EVENT, pcc.CEPH_FS_NAME)
	if err != nil {
		errMsg := fmt.Sprintf("Ceph FS [%v] creation verification failed...ERROR: %v", pcc.CEPH_FS_NAME, err)
		fmt.Println(errMsg)
		assert.Fatalf(errMsg)
		return
	}else {
		fmt.Printf("Ceph FS [%v] created properly..[%v]\n", pcc.CEPH_FS_NAME, s.Msg)
	}
}

func testVerifyCephFSDeletion(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	fmt.Printf("Verifying Ceph FS [%v] deletion...Timeout:[%v sec]\n", pcc.CEPH_FS_NAME, pcc.CEPH_FS_DELETION_TIMEOUT)

	s, err := pcc.GetPccClient().VerifyCeph(startTime, pcc.CEPH_FS_DELETE_EVENT, pcc.CEPH_FS_NAME)
	if err != nil {
		errMsg := fmt.Sprintf("Ceph FS [%v] deletion verification failed...ERROR: %v", pcc.CEPH_FS_NAME, err)
		fmt.Println(errMsg)
		assert.Fatalf(errMsg)
		return
	}else {
		fmt.Printf("Ceph FS [%v] deleted properly..[%v]\n", pcc.CEPH_FS_NAME, s.Msg)
	}
}
func testVerifyCephPoolCreation(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var err []error
	for _, pools := range pcc.CephPools {
		for pool, _ := range pools {
			fmt.Printf("Verifying ceph pool [%v] creation...Timeout:[%v sec]\n", pool, pcc.CEPH_POOL_CREATION_TIMEOUT)
			_, errP := pcc.GetPccClient().GetCephPool(pool)
			if errP != nil {
				errMsg := fmt.Sprintf("Ceph pool [%v] creation failed..ERROR: %v", pool, errP)
				fmt.Println(errMsg)
				err = append(err, errP)
			}else {
				s, errT := pcc.GetPccClient().VerifyCeph(startTime, pcc.CEPH_POOL_CREATE_EVENT, pool)
				if errT != nil {
					errMsg := fmt.Sprintf("Ceph pool [%v] creation verification failed...ERROR: %v", pool, errT)
					fmt.Println(errMsg)
					err = append(err, errT)
				}else {
					fmt.Printf("Ceph pool [%v] created successfully..[%v]\n", pool, s.Msg)
				}
			}
		}
	}
	if len(err) > 0 {
		errMsg := fmt.Sprintf("Ceph pool creation failed: %v", err)
		fmt.Println(errMsg)
		assert.Fatalf(errMsg)
		return
	}
}

func testVerifyCephPoolDeletion(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var err []error
	for _, pools := range pcc.CephPools {
		for pool, _ := range pools {
			fmt.Printf("Verifying ceph pool [%v] deletion...Timeout:[%v sec]\n", pool, pcc.CEPH_POOL_CREATION_TIMEOUT)
			s, errT := pcc.GetPccClient().VerifyCeph(startTime, pcc.CEPH_POOL_DELETE_EVENT, pool)
			if errT != nil {
				errMsg := fmt.Sprintf("Ceph pool [%v] deletion verification failed...ERROR: %v", pool, errT)
				fmt.Println(errMsg)
				err = append(err, errT)
			}else {
				fmt.Printf("Ceph pool [%v] deleted successfully..[%v]\n", pool, s.Msg)
			}

		}
	}
	if len(err) > 0 {
		errMsg := fmt.Sprintf("Ceph pool deletion failed: %v", err)
		fmt.Println(errMsg)
		assert.Fatalf(errMsg)
		return
	}
}

func testVerifyCephInstallation(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	fmt.Printf("Verifying ceph installation...Timeout:[%v sec]\n", pcc.CEPH_3_NODE_INSTALLATION_TIMEOUT)
	_, err := pcc.GetPccClient().GetCephCluster()
	if err != nil {
		errMsg := fmt.Sprintf("Ceph cluster installation verification failed...ERROR: %v", err)
		fmt.Println(errMsg)
		assert.Fatalf(errMsg)
		return
	}else {
		s, err := pcc.GetPccClient().VerifyCeph(startTime, pcc.CEPH_CLUSTER_INSTALL_EVENT, pcc.GetPccClient().GetCephClusterName())
		if err != nil {
			errMsg := fmt.Sprintf("Ceph cluster installation verification failed...ERROR: %v", err)
			fmt.Println(errMsg)
			assert.Fatalf(errMsg)
			return
		}else {
			fmt.Printf("Ceph cluster [%v] deployed properly..[%v]\n", pcc.GetPccClient().GetCephClusterName(), s.Msg)
		}
	}
}

func testVerifyCephUninstallation(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	fmt.Printf("Verifying ceph uninstallation...Timeout:[%v sec]\n", pcc.CEPH_3_NODE_UNINSTALLATION_TIMEOUT)
	_, err := pcc.GetPccClient().GetCephCluster()
	if err != nil {
		errMsg := fmt.Sprintf("Ceph cluster uninstallation verification failed...ERROR: %v", err)
		fmt.Println(errMsg)
		assert.Fatalf(errMsg)
		return
	}else {
		s, err := pcc.GetPccClient().VerifyCeph(startTime, pcc.CEPH_CLUSTER_UNINSTALL_EVENT, pcc.GetPccClient().GetCephClusterName())
		if err != nil {
			errMsg := fmt.Sprintf("Ceph cluster uninstallation failed...ERROR: %v\n", err)
			fmt.Println(errMsg)
			assert.Fatalf(errMsg)
			return
		} else {
			fmt.Printf("Ceph cluster [%v] undeployed properly..[%v]\n", pcc.GetPccClient().GetCephClusterName(), s.Msg)
		}
	}
}

func getNodesList(num int) ([]pcc.CephNodes, error){
	var(
		DIM = num
		sNodes = make([]pcc.CephNodes, DIM)
		nodesSetCompleted bool = false
	)
	var j = 0
	for _, i := range Env.Invaders {
		sNodes[j] = pcc.CephNodes{ID: NodebyHostIP[i.HostIp]}
		j++
		if j == DIM {
			nodesSetCompleted = true
		}
	}
	if pcc.GetCephConfig().HighAvailability && len(sNodes) < 2 {
		return sNodes, fmt.Errorf("Atleast 2 invaders are required for High Availability configuration..Found:%v",len(sNodes))
	}
	for _, i := range Env.Servers {
		if nodesSetCompleted {
			continue
		}
		sNodes[j] = pcc.CephNodes{ID: NodebyHostIP[i.HostIp]}
		j++
		if j == DIM {
			nodesSetCompleted = true
			continue
		}
	}
	return sNodes, nil
}
