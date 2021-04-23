package main

import (
	"context"
	"encoding/json"
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

	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

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
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
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
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		case <-tick:
			pool, err := Pcc.GetCephPool("bb-rgw-pool", cluster.Id)
			if err != nil {
				msg := fmt.Sprintf("Failed to get deploy status "+
					"%v", err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
			}
			switch pool.DeployStatus {
			case pcc.RGW_DEPLOY_STATUS_PROGRESS:
				log.AuctaLogger.Info("RGW installation in progress...")
			case pcc.RGW_DEPLOY_STATUS_COMPLETED:
				log.AuctaLogger.Info("RGW installation completed")
				return
			case pcc.RGW_DEPLOY_STATUS_FAILED:
				msg := "RGW installation failed"
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
			default:
				msg := fmt.Sprintf("Unexpected status - %v",
					pool.DeployStatus)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
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
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		}
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
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

	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

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
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		case <-tick:
			gw, err := Pcc.GetRadosGW(id)
			if err != nil {
				msg := fmt.Sprintf("Failed to get deploy status "+
					"%v", err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
			}

			switch gw.DeployStatus {
			case pcc.RGW_DEPLOY_STATUS_PROGRESS:
				log.AuctaLogger.Info("RGW installation in progress...")
			case pcc.RGW_DEPLOY_STATUS_COMPLETED:
				log.AuctaLogger.Info("RGW installation completed")
				return
			case pcc.RGW_DEPLOY_STATUS_FAILED:
				msg := "RGW installation failed"
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
			default:
				msg := fmt.Sprintf("Unexpected status - %v",
					gw.DeployStatus)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
			}
		}
	}
}

func createCephProfilesWithPermission(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	profileRWD := s3.S3Profile{
		Username:         "blackbox-rwd",
		ReadPermission:   true,
		WritePermission:  true,
		DeletePermission: true,
	}

	profiles = append(profiles, profileRWD)
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
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	timeout := time.After(5 * time.Minute)
	tick := time.Tick(15 * time.Second)
	found := false

	for !found {
		select {
		case <-timeout:
			msg := "Timed out waiting for RGW"
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		case <-tick:
			acs, err := Pcc.GetAppCredentials("ceph")

			if err != nil {
				msg := fmt.Sprintf("Failed to get deploy status "+
					"%v", err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
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

	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	bucketName = "test-bucket-bb"

	if exists, errBucketExists := minioClient.BucketExists(ctx, bucketName); errBucketExists == nil {
		if !exists {
			err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
			if err != nil {
				msg := fmt.Sprintf("%v", err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
			}
		} else {
			log.AuctaLogger.Warnf("Bucket %s already exists", bucketName)
		}
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
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
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	log.AuctaLogger.Infof("Successfully uploaded %s", objectName)
}

func testRetrieveObject(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	_, err := minioClient.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})

	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	log.AuctaLogger.Infof("Successfully retrieved object %s", objectName)
}

func testRemoveObject(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Infof("Removing object: %s", objectName)
	err := minioClient.RemoveObject(context.Background(), bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
}

func testRemoveBucket(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Infof("Removing bucket: %s", bucketName)
	err := minioClient.RemoveBucket(context.Background(), bucketName)
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
}

func testRemoveRGW(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	//Delete is implemented as a synchronous operation, a timeout doesn't necessarily
	//imply a failure

	if _, err := Pcc.DeleteRadosGW(id); err != nil && !strings.Contains(err.Error(), "Timeout") {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	timeout := time.After(45 * time.Minute)
	tick := time.Tick(1 * time.Minute)
	for true {
		select {
		case <-timeout:
			msg := "Timed out waiting for RGW"
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		case <-tick:
			gws, err := Pcc.GetRadosGWs()

			if err != nil {
				msg := fmt.Sprintf("Failed to get deploy status "+
					"%v", err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
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
