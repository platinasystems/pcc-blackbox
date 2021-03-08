package main

import (
	"fmt"
	"testing"
	"time"

	log "github.com/platinasystems/go-common/logs"
	model "github.com/platinasystems/pcc-blackbox/models"
	ceph3 "github.com/platinasystems/pcc-models/ceph"
	"github.com/platinasystems/tiles/pccserver/controllers/ceph"
	"github.com/platinasystems/tiles/pccserver/models"
)

////
// TEST
////
// Test the cache addition. Delete the cache at the end
func testCephCacheSetup(t *testing.T) {
	if clusters, err := Pcc.GetAllCephClusters(); err == nil {
		if len(clusters) == 0 {
			testCeph(t)
		}
	}
}

func addCephCache(t *testing.T, cacheMode string, name string) *ceph3.CephCacheTier {

	res := model.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "addCephCache")
	CheckDependencies(t, res, Env.CheckCephConfiguration, CheckCephClusterExists)

	var (
		clusters []*models.CephCluster
		pools    []*models.CephPool
		err      error
	)

	if clusters, err = Pcc.GetAllCephClusters(); err == nil {
		if len(clusters) == 0 {
			msg := "you should add at least one ceph cluster"
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		}

		cluster := clusters[0]

		if pools, err = Pcc.GetAllCephPools(cluster.Id); err == nil {
			if len(pools) == 0 {
				msg := "you should add at least one ceph pool"
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
			}

			pool := pools[0]

			log.AuctaLogger.Infof("\nCEPH CACHE: Selected the pool [%s]", pool.Name)
			request := ceph.CacheRequest{}
			request.StoragePoolID = pool.Id
			request.Size = pool.Size
			request.Quota = pool.Quota
			request.QuotaUnit = pool.QuotaUnit
			if cacheMode != "" {
				request.Mode = cacheMode
			}

			poolName := fmt.Sprintf("%s-%s", name, pool.Name)
			request.Name = poolName
			if _, err := Pcc.CreateCephCache(&request); err != nil {
				msg := fmt.Sprintf("%v\n", err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
			}

			for limit := 1; limit <= 15; limit++ { // TODO sync on notification
				log.AuctaLogger.Infof("\nWait for %s creation...", poolName)
				time.Sleep(time.Duration(10) * time.Second)
				if caches, err := Pcc.GetCephCaches(); err == nil {
					for i := range caches {
						if pool, err := Pcc.GetCephPool2(caches[i].CachePoolID); err == nil {
							if pool.Name == poolName {
								return caches[i]
							}
						} else {
							msg := fmt.Sprintf("%v\n", err)
							res.SetTestFailure(msg)
							log.AuctaLogger.Error(msg)
							t.FailNow()
						}
					}
				} else {
					msg := fmt.Sprintf("%v\n", err)
					res.SetTestFailure(msg)
					log.AuctaLogger.Error(msg)
					t.FailNow()
				}
			}
			msg := "some error happens in creation"
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		} else {
			msg := fmt.Sprintf("%v\n", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		}
	} else {
		msg := fmt.Sprintf("%v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	return nil
}

// Test the cache addition. Delete the cache at the end
func testCephCacheAdd(t *testing.T) {

	res := model.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "addIpam")
	CheckDependencies(t, res, Env.CheckCephConfiguration, CheckCephClusterExists)

	log.AuctaLogger.Info("\nCEPH CACHE: adding the cache")
	var (
		err        error
		cephCache  *ceph3.CephCacheTier
		cephCache2 *ceph3.CephCacheTier
	)

	cephCache = addCephCache(t, "", "blackbox")

	log.AuctaLogger.Infof("\nCEPH CACHE: added the cache %d", cephCache.ID)
	defer func() {
		Pcc.DeleteCephCache(cephCache.ID)
		log.AuctaLogger.Infof("\nCEPH CACHE: removed the cache %d\n", cephCache.ID)
	}()

	// Check if the cache exist
	if cephCache2, err = Pcc.GetCephCache(cephCache.ID); err != nil {
		msg := fmt.Sprintf("%v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	if !cephCache.IsWriteback() {
		msg := "The cache should be writeback mode by default"
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	if cephCache.StoragePoolID != cephCache2.StoragePoolID {
		msg := "the storage pool is different"
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
}

// Test the cache deletion.
func testCephCacheDelete(t *testing.T) {

	res := model.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "testCephCacheDelete")
	CheckDependencies(t, res, Env.CheckCephConfiguration, CheckCephClusterExists)

	log.AuctaLogger.Info("\nCEPH CACHE: adding the cache")
	var (
		err        error
		cephCache  *ceph3.CephCacheTier
		cephCaches []*ceph3.CephCacheTier
	)

	cephCache = addCephCache(t, "", "blackbox2")
	Pcc.DeleteCephCache(cephCache.ID)

	// Check if the cache exist
	if _, err = Pcc.GetCephCaches(); err == nil {
		for i := range cephCaches {
			cc := cephCaches[i]

			if cc.ID == cephCache.ID {
				msg := "the cache still exist"
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
			}
		}
	} else {
		msg := fmt.Sprintf("%v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
}
