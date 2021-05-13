package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	m "github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/pcc-models/authentication"
	"github.com/platinasystems/pcc-models/ceph"
	"github.com/platinasystems/pcc-models/s3"
	"github.com/platinasystems/test"
	"github.com/platinasystems/tiles/pccserver/models"
	"strings"
	"testing"
	"time"
)

var (
	id          uint64
	poolID      uint64
	cluster     *models.CephCluster
	targetNode  *models.CephNode
	profileRGW  s3.S3Profile
	minioClient *minio.Client
	bucketName  string
	objectName  string
	profiles    []s3.S3Profile
	ctx         = context.Background()
)

func testRGW(t *testing.T) {
	t.Run("createPoolRGW", createPoolRGW)
	t.Run("verifyPool", verifyPool)
	if t.Failed() {
		return
	}
	t.Run("installRGW", installRGW)
	t.Run("verifyRGW", verifyRGWDeploy)
	t.Run("createCephProfilesWithPermission", createCephProfilesWithPermission)
	if t.Failed() {
		return
	}
	t.Run("testAllProfilesPermission", testAllProfilesPermission)
	if t.Failed() {
		return
	}
	t.Run("testRemoveRGW", testRemoveRGW)
}

func testAllProfilesPermission(t *testing.T) {
	for _, p := range profiles {
		profileRGW = p
		t.Run("addCephProfile", addCephCredential)
		t.Run("testProfilePermission", testProfilePermission)
	}
}
func testProfilePermission(t *testing.T) {
	t.Run("testCreateBucket", testCreateBucket)
	t.Run("testPutObject", testPutObject)
	t.Run("testRetrieveObject", testRetrieveObject)
	t.Run("testRemoveObject", testRemoveObject)
	t.Run("testRemoveBucket", testRemoveBucket)
}

func createPoolRGW(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var (
		err error
	)

	cluster, err = Pcc.GetCephCluster(Env.RGWConfiguration.ClusterName)
	checkError(t, res, err)

	poolRequest := pcc.CreateCephPoolRequest{
		CephClusterId: cluster.Id,
		Name:          Env.RGWConfiguration.PoolName,
		QuotaUnit:     "GiB",
		Quota:         "1",
		PoolType:      models.CEPH_POOL_PROFILE_TYPE_REPLICATED.String(),
		Size:          3,
	}

	var poolRes *models.CephPool
	poolRes, err = Pcc.GetCephPool(Env.RGWConfiguration.PoolName, cluster.Id)

	if poolRes != nil {
		log.AuctaLogger.Warn("Pool already exists, skipping creation")
		poolID = poolRes.Id
		return
	}

	poolID, err = Pcc.CreateCephPool(poolRequest)
	checkError(t, res, err)
}

func verifyPool(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	timeout := time.After(15 * time.Minute)
	tick := time.Tick(30 * time.Second)
	for true {
		select {
		case <-timeout:
			msg := "Timed out waiting for pool response"
			checkError(t, res, errors.New(msg))
		case <-tick:
			pool, err := Pcc.GetCephPool("bb-rgw-pool", cluster.Id)
			if err != nil {
				msg := fmt.Sprintf("Failed to get deploy status "+
					"%v", err)
				checkError(t, res, errors.New(msg))
			}
			switch pool.DeployStatus {
			case pcc.RGW_DEPLOY_STATUS_PROGRESS:
				log.AuctaLogger.Info("RGW installation in progress...")
			case pcc.RGW_DEPLOY_STATUS_COMPLETED:
				log.AuctaLogger.Info("RGW installation completed")
				return
			case pcc.RGW_DEPLOY_STATUS_FAILED:
				msg := "RGW installation failed"
				checkError(t, res, errors.New(msg))
			default:
				msg := fmt.Sprintf("Unexpected status - %v",
					pool.DeployStatus)
				checkError(t, res, errors.New(msg))
			}
		}
	}
}

func installRGW(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var certId uint64
	certName := Env.RGWConfiguration.CertName

	if exists, cert, err := Pcc.FindCertificate(certName); err == nil {
		if exists {
			certId = cert.Id
		} else {
			msg := fmt.Sprintf("No certificate found with name %s", certName)
			checkError(t, res, errors.New(msg))
		}
	} else {
		checkError(t, res, err)
	}

	targetNode = cluster.Nodes[randomGenerator.Intn(len(cluster.Nodes))]

	RGWRequest := ceph.RadosGateway{
		CephPoolID:    poolID,
		CertificateID: certId,
		Name:          Env.RGWConfiguration.RGWName,
		Port:          443,
		TargetNodes:   []uint64{targetNode.NodeId},
	}

	addedRGW, err := Pcc.AddRadosGW(&RGWRequest)
	checkError(t, res, err)

	id = addedRGW.ID
}

func verifyRGWDeploy(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	timeout := time.After(45 * time.Minute)
	tick := time.Tick(1 * time.Minute)
	for true {
		select {
		case <-timeout:
			msg := "Timed out waiting for RGW"
			checkError(t, res, errors.New(msg))
		case <-tick:
			gw, err := Pcc.GetRadosGW(id)
			if err != nil {
				msg := fmt.Sprintf("Failed to get deploy status "+
					"%v", err)
				checkError(t, res, errors.New(msg))
			}

			switch gw.DeployStatus {
			case pcc.RGW_DEPLOY_STATUS_PROGRESS:
				log.AuctaLogger.Info("RGW installation in progress...")
			case pcc.RGW_DEPLOY_STATUS_COMPLETED:
				log.AuctaLogger.Info("RGW installation completed")
				return
			case pcc.RGW_DEPLOY_STATUS_FAILED:
				msg := "RGW installation failed"
				checkError(t, res, errors.New(msg))
			default:
				msg := fmt.Sprintf("Unexpected status - %v",
					gw.DeployStatus)
				checkError(t, res, errors.New(msg))
			}
		}
	}
}

func createCephProfilesWithPermission(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	profileRWD := s3.S3Profile{
		Username:         "blackbox-full",
		ReadPermission:   true,
		WritePermission:  true,
		DeletePermission: true,
	}

	profileRW := s3.S3Profile{
		Username:         "blackbox-rw",
		ReadPermission:   true,
		WritePermission:  true,
		DeletePermission: false,
	}

	profileR := s3.S3Profile{
		Username:         "blackbox-r",
		ReadPermission:   true,
		WritePermission:  false,
		DeletePermission: false,
	}

	profiles = append(profiles, profileRWD, profileRW, profileR)

}

func addCephCredential(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	serviceType := "ceph"

	appCredential := authentication.AuthProfile{
		Name:          fmt.Sprintf("%s-%s", profileRGW.Username, serviceType),
		Type:          serviceType,
		ApplicationId: id,
		Profile:       profileRGW,
		Active:        true}

	log.AuctaLogger.Infof("creating the ceph profile %v", appCredential)

	var err error
	_, err = Pcc.CreateAppCredentialProfileCeph(&appCredential)
	checkError(t, res, err)

	timeout := time.After(5 * time.Minute)
	tick := time.Tick(15 * time.Second)
	found := false

	for !found {
		select {
		case <-timeout:
			msg := "Timed out waiting for RGW"
			checkError(t, res, errors.New(msg))
		case <-tick:
			acs, err := Pcc.GetAppCredentials("ceph")
			if err != nil {
				msg := fmt.Sprintf("Failed to get deploy status "+
					"%v", err)
				checkError(t, res, errors.New(msg))
			}

			for _, ac := range acs {
				if ac.Name == fmt.Sprintf("%s-%s", profileRGW.Username, "ceph") {
					jsonString, _ := json.Marshal(ac.Profile)
					json.Unmarshal(jsonString, &profileRGW)
					if profileRGW.AccessKey != "" {
						found = true
					}
				}
			}
		}
	}

	log.AuctaLogger.Infof("created the ceph profile", profileRGW)
}

func initS3Client(host string, accessKey string, secretKey string) (*minio.Client, error) {
	endpoint := fmt.Sprintf("%s:443", host)
	accessKeyID := accessKey
	secretAccessKey := secretKey
	useSSL := true

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

func testCreateBucket(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	node, err := Pcc.GetNode(targetNode.NodeId)

	minioClient, err = initS3Client(node.Host, profileRGW.AccessKey, profileRGW.SecretKey)
	checkError(t, res, err)

	bucketName = "test-bucket-bb"

	if exists, errBucketExists := minioClient.BucketExists(ctx, bucketName); errBucketExists == nil {
		if !exists {
			err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
			checkResultByPermission(t, res, err, profileRGW.WritePermission, "create a bucket")
		} else {
			log.AuctaLogger.Warnf("Bucket %s already exists", bucketName)
		}
	} else {
		checkError(t, res, err)
	}

	buckets, err := minioClient.ListBuckets(ctx)
	log.AuctaLogger.Info(buckets)
}

func testPutObject(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var err error
	objectName = "testEnv.json"
	filePath := "testEnv.json"

	_, err = minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{})
	checkResultByPermission(t, res, err, profileRGW.WritePermission, "put an object")
}

func testRetrieveObject(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	_, err := minioClient.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	checkResultByPermission(t, res, err, profileRGW.ReadPermission, "retrieve an object")
}

func testRemoveObject(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Infof("Removing object: %s", objectName)
	err := minioClient.RemoveObject(context.Background(), bucketName, objectName, minio.RemoveObjectOptions{})
	checkResultByPermission(t, res, err, profileRGW.DeletePermission, "remove an object")
}

func testRemoveBucket(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Infof("Removing bucket: %s", bucketName)
	err := minioClient.RemoveBucket(context.Background(), bucketName)
	checkResultByPermission(t, res, err, profileRGW.DeletePermission, "remove a bucket")
}

func testRemoveRGW(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	//Delete is implemented as a synchronous operation, a timeout doesn't necessarily
	//imply a failure

	if _, err := Pcc.DeleteRadosGW(id); err != nil && !strings.Contains(err.Error(), "Timeout") {
		checkError(t, res, err)
	}

	timeout := time.After(45 * time.Minute)
	tick := time.Tick(1 * time.Minute)
	for true {
		select {
		case <-timeout:
			msg := "Timed out waiting for RGW"
			checkError(t, res, errors.New(msg))
		case <-tick:
			gws, err := Pcc.GetRadosGWs()

			if err != nil {
				msg := fmt.Sprintf("Failed to get deploy status "+
					"%v", err)
				checkError(t, res, errors.New(msg))
			}

			found := false
			for _, gw := range gws {
				if gw.ID == id {
					found = true
				}
			}

			if !found {
				log.AuctaLogger.Info("RGW uninstalled successfully")
				return
			}
		}
	}
}

func checkResultByPermission(t *testing.T, res *m.TestResult, err error, permission bool, operation string) {
	if err == nil {
		if permission {
			log.AuctaLogger.Infof("Success: %s", operation)
		} else {
			msg := fmt.Sprintf("The user is not supposed to be able to %s", operation)
			checkError(t, res, errors.New(msg))
		}
	} else {
		if permission {
			msg := fmt.Sprintf("%v", err)
			checkError(t, res, errors.New(msg))
		} else if strings.Contains(err.Error(), "Access Denied") {
			log.AuctaLogger.Infof("Success: %s, denied", operation)
		} else if !strings.Contains(err.Error(), "Access Denied") {
			msg := fmt.Sprintf("%v", err)
			checkError(t, res, errors.New(msg))
		}
	}
}
