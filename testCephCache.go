package main

import (
	"fmt"
	ceph3 "github.com/platinasystems/pcc-models/ceph"
	"github.com/platinasystems/tiles/pccserver/controllers/ceph"
	"github.com/platinasystems/tiles/pccserver/models"
	"testing"
	"time"
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
	var (
		clusters []*models.CephCluster
		pools    []*models.CephPool
		err      error
	)

	if clusters, err = Pcc.GetAllCephClusters(); err == nil {
		if len(clusters) == 0 {
			t.Fatal("you should add at least one ceph cluster")
		}

		cluster := clusters[0]

		if pools, err = Pcc.GetAllCephPools(cluster.Id); err == nil {
			if len(pools) == 0 {
				t.Fatal("you should add at least one ceph pool")
			}

			pool := pools[0]

			fmt.Printf("\nCEPH CACHE: Selected the pool [%s]", pool.Name)
			request := ceph.CacheRequest{}
			request.StoragePoolID = pool.Id
			request.Size = pool.Size
			request.Quota = pool.Quota
			request.QuotaUnit = pool.QuotaUnit
			if cacheMode != "" {
				request.Mode = &cacheMode
			}

			poolName := fmt.Sprintf("%s-%s", name, pool.Name)
			request.Name = poolName
			if _, err := Pcc.CreateCephCache(&request); err != nil {
				t.Fatal(err)
			}

			for limit := 1; limit <= 15; limit++ { // TODO sync on notification
				fmt.Printf("\nWait for %s creation...", poolName)
				time.Sleep(time.Duration(10) * time.Second)
				if caches, err := Pcc.GetCephCaches(); err == nil {
					for i := range caches {
						if pool, err := Pcc.GetCephPool2(caches[i].CachePoolID); err == nil {
							if pool.Name == poolName {
								return caches[i]
							}
						} else {
							t.Fatal(err)
						}
					}
				} else {
					t.Fatal(err)
				}
			}
			t.Fatal("some error happens in creation")
		} else {
			t.Fatal(err)
		}
	} else {
		t.Fatal(err)
	}

	return nil
}

// Test the cache addition. Delete the cache at the end
func testCephCacheAdd(t *testing.T) {
	fmt.Println("\nCEPH CACHE: adding the cache")
	var (
		err        error
		cephCache  *ceph3.CephCacheTier
		cephCache2 *ceph3.CephCacheTier
	)

	cephCache = addCephCache(t, "", "blackbox")

	fmt.Printf("\nCEPH CACHE: added the cache %d", cephCache.ID)
	defer func() {
		Pcc.DeleteCephCache(cephCache.ID)
		fmt.Printf("\nCEPH CACHE: removed the cache %d\n", cephCache.ID)
	}()

	// Check if the cache exist
	if cephCache2, err = Pcc.GetCephCache(cephCache.ID); err != nil {
		t.Fatal(err)
	}

	if !cephCache.IsWriteback() {
		t.Fatal("The cache should be writeback mode by default")
	}

	if cephCache.StoragePoolID != cephCache2.StoragePoolID {
		t.Fatal("the storage pool is different")
	}
}

// Test the cache deletion.
func testCephCacheDelete(t *testing.T) {
	fmt.Println("\nCEPH CACHE: adding the cache")
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
				t.Fatal("the cache still exist")
			}
		}
	} else {
		t.Fatal(err)
	}
}
