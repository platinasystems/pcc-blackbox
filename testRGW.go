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

const (
	QuotaExceeded  = "QuotaExceeded"
	TooManyBuckets = "TooManyBuckets"
	AccessDenied   = "Access Denied"
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

type Permissions struct {
	ReadPermission   bool
	WritePermission  bool
	DeletePermission bool
}

type BucketQuotas struct {
	MaxBuckets      uint16
	MaxBucketSize   int64
	MaxBucketObject int64
}

type UserQuotas struct {
	MaxUserObjects int64
	MaxUserSize    int64
}

func testRGW(t *testing.T) {
	t.Run("createPoolRGW", createPoolRGW)
	t.Run("verifyPoolRGW", verifyPoolRGW)
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
	t.Run("testBucketQuotas", testBucketQuotas)
	t.Run("testUserQuotas", testUserQuotas)
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
	CheckDependencies(t, res, CheckCephClusterRGWExists)

	var (
		err error
	)

	cluster, err = Pcc.GetCephCluster(Env.RGWConfiguration.ClusterName)
	checkError(t, res, err)

	poolRequest := pcc.CreateCephPoolRequest{
		CephClusterId: cluster.Id,
		Name:          Env.RGWConfiguration.PoolName,
		QuotaUnit:     "MiB",
		Quota:         "100",
		PoolType:      models.CEPH_POOL_PROFILE_TYPE_REPLICATED.String(),
		Size:          3,
	}

	var poolRes *models.CephPool
	poolRes, err = Pcc.GetCephPool(Env.RGWConfiguration.PoolName, cluster.Id)

	if err == nil {
		log.AuctaLogger.Warn("Pool already exists, skipping creation")
		poolID = poolRes.Id
		return
	}

	poolID, err = Pcc.CreateCephPool(poolRequest)
	checkError(t, res, err)
}

func verifyPoolRGW(t *testing.T) {
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
			pool, err := Pcc.GetCephPool(Env.RGWConfiguration.PoolName, cluster.Id)
			if err != nil {
				msg := fmt.Sprintf("Failed to get deploy status "+
					"%v", err)
				checkError(t, res, errors.New(msg))
			}
			switch pool.DeployStatus {
			case pcc.RGW_DEPLOY_STATUS_PROGRESS:
				log.AuctaLogger.Info("pool installation in progress...")
			case pcc.RGW_DEPLOY_STATUS_COMPLETED:
				log.AuctaLogger.Info("pool installation completed")
				return
			case pcc.RGW_DEPLOY_STATUS_FAILED:
				msg := "pool installation failed"
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
	CheckDependencies(t, res, CheckCephClusterRGWExists)

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

	rwdPermission := Permissions{ReadPermission: true, WritePermission: true, DeletePermission: true}
	rwPermission := Permissions{ReadPermission: true, WritePermission: true, DeletePermission: false}
	rPermission := Permissions{ReadPermission: true, WritePermission: false, DeletePermission: false}

	profileRWD := s3.S3Profile{
		Username:         "blackbox-full",
		ReadPermission:   &rwdPermission.ReadPermission,
		WritePermission:  &rwdPermission.WritePermission,
		DeletePermission: &rwdPermission.DeletePermission,
	}

	//default rw profile
	profileRW := s3.S3Profile{
		Username:         "blackbox-rw",
		ReadPermission:   &rwPermission.ReadPermission,
		WritePermission:  &rwPermission.WritePermission,
		DeletePermission: &rwPermission.DeletePermission,
	}

	profileR := s3.S3Profile{
		Username:         "blackbox-r",
		ReadPermission:   &rPermission.ReadPermission,
		WritePermission:  &rPermission.WritePermission,
		DeletePermission: &rPermission.DeletePermission,
	}

	profiles = append(profiles, profileRWD, profileRW, profileR)

}

func addCephCredential(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	createAppCredentialProfileCeph(t, res, &profileRGW, id)
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

	bucketName = fmt.Sprintf("bucket-%s", profileRGW.Username)

	alreadyExists, err := createBucket(t, res, ctx, bucketName)
	if !alreadyExists {
		checkResultByPermission(t, res, err, *profileRGW.WritePermission, "create a bucket")
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
	checkResultByPermission(t, res, err, *profileRGW.WritePermission, "put an object")
}

func testRetrieveObject(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	_, err := minioClient.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	checkResultByPermission(t, res, err, *profileRGW.ReadPermission, "retrieve an object")
}

func testRemoveObject(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Infof("Removing object: %s", objectName)
	err := minioClient.RemoveObject(context.Background(), bucketName, objectName, minio.RemoveObjectOptions{})
	checkResultByPermission(t, res, err, *profileRGW.DeletePermission, "remove an object")
}

func testRemoveBucket(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Infof("Removing bucket: %s", bucketName)
	err := minioClient.RemoveBucket(context.Background(), bucketName)
	checkResultByPermission(t, res, err, *profileRGW.DeletePermission, "remove a bucket")
}

func testBucketQuotas(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	bucketQuotas := BucketQuotas{MaxBuckets: 1, MaxBucketSize: 8, MaxBucketObject: 2}
	rwdPermission := Permissions{ReadPermission: true, WritePermission: true, DeletePermission: true}

	profileBucketQuotas := s3.S3Profile{
		Username:          "bucket-quota-profile",
		ReadPermission:    &rwdPermission.ReadPermission,
		WritePermission:   &rwdPermission.WritePermission,
		DeletePermission:  &rwdPermission.DeletePermission,
		MaxBuckets:        &bucketQuotas.MaxBuckets,
		MaxBucketSize:     &bucketQuotas.MaxBucketSize,
		MaxBucketSizeUnit: "KiB",
		MaxBucketObjects:  &bucketQuotas.MaxBucketObject}

	createAppCredentialProfileCeph(t, res, &profileBucketQuotas, id)

	n, err := Pcc.GetNode(targetNode.NodeId)
	checkError(t, res, err)

	minioClient, err = initS3Client(n.Host, profileBucketQuotas.AccessKey, profileBucketQuotas.SecretKey)
	checkError(t, res, err)

	log.AuctaLogger.Info("test maximum number of buckets")
	numBucket := int(bucketQuotas.MaxBuckets) + 1
	var bucketNames []string
	for i := 0; i < numBucket; i++ {
		name := fmt.Sprintf("bucket-%d", i)
		if _, err = createBucket(t, res, ctx, name); err == nil {
			log.AuctaLogger.Infof("Bucket: %s correctly created", name)
			bucketNames = append(bucketNames, name)
		} else {
			if strings.Contains(err.Error(), TooManyBuckets) {
				log.AuctaLogger.Infof(fmt.Sprintf("Negative case pass: %s, exceeded maximum number of buckets", TooManyBuckets))
			} else {
				checkError(t, res, err)
			}
		}
	}

	if _, err = createFile("10kb_file", 1e4); err != nil {
		checkError(t, res, err)
	}
	if _, err = createFile("1kb_file", 1e3); err != nil {
		checkError(t, res, err)
	}

	log.AuctaLogger.Info("test maximum bucket size")
	//put file that exceeds maximum bucket size
	if info, err := minioClient.FPutObject(ctx, bucketNames[0], "obj", "10kb_file", minio.PutObjectOptions{}); err == nil {
		err = errors.New(fmt.Sprintf("Error: uploaded a file of size %d that exceeds the maximum bucket size %d %s", info.Size, bucketQuotas.MaxBucketSize, profileBucketQuotas.MaxBucketSizeUnit))
		checkError(t, res, err)
	} else {
		if strings.Contains(err.Error(), QuotaExceeded) {
			log.AuctaLogger.Infof(fmt.Sprintf("Negative case pass: error trying to uploade file of size %d that exceeds maximum bucket size of %d %s", info.Size, bucketQuotas.MaxBucketSize, profileBucketQuotas.MaxBucketSizeUnit))
		} else {
			checkError(t, res, err)
		}
	}

	log.AuctaLogger.Info("test maximum object number on bucket")
	numObjects := int(bucketQuotas.MaxBucketObject) + 1
	var objectNames []string
	for i := 0; i < numObjects; i++ {
		obj := fmt.Sprintf("object-%d", i)
		if _, err = minioClient.FPutObject(ctx, bucketNames[0], obj, "1kb_file", minio.PutObjectOptions{}); err == nil {
			log.AuctaLogger.Infof("Object: %s correctly uploaded", obj)
			objectNames = append(objectNames, obj)
		} else {
			if strings.Contains(err.Error(), QuotaExceeded) {
				log.AuctaLogger.Infof(fmt.Sprintf("Negative case pass: error, maximum number of objects inside a bucket [%d] exceeded", bucketQuotas.MaxBucketObject))
			} else {
				checkError(t, res, err)
			}
		}
	}

	log.AuctaLogger.Info("removing created objects and buckets")
	for _, obj := range objectNames {
		log.AuctaLogger.Infof("Removing object: %s", obj)
		err = minioClient.RemoveObject(context.Background(), bucketNames[0], obj, minio.RemoveObjectOptions{})
		checkError(t, res, err)
	}
	for _, bucket := range bucketNames {
		log.AuctaLogger.Infof("Removing bucket: %s", bucket)
		err = minioClient.RemoveBucket(context.Background(), bucket)
		checkError(t, res, err)
	}
}

func testUserQuotas(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	userQuotas := UserQuotas{MaxUserObjects: 2, MaxUserSize: 8}
	rwdPermission := Permissions{ReadPermission: true, WritePermission: true, DeletePermission: true}

	profileUserQuotas := s3.S3Profile{
		Username:         "user-quota-profile",
		ReadPermission:   &rwdPermission.ReadPermission,
		WritePermission:  &rwdPermission.WritePermission,
		DeletePermission: &rwdPermission.DeletePermission,
		MaxUserSize:      &userQuotas.MaxUserSize,
		MaxUserSizeUnit:  "KiB",
		MaxUserObjects:   &userQuotas.MaxUserObjects}

	createAppCredentialProfileCeph(t, res, &profileUserQuotas, id)

	n, err := Pcc.GetNode(targetNode.NodeId)
	checkError(t, res, err)

	minioClient, err = initS3Client(n.Host, profileUserQuotas.AccessKey, profileUserQuotas.SecretKey)
	checkError(t, res, err)

	bucket := "user-quota-bucket"
	_, err = createBucket(t, res, ctx, bucket)
	checkError(t, res, err)

	log.AuctaLogger.Info("test maximum user size")
	//put file that exceeds maximum user size
	obj := "user-quota-obj"
	if info, err := minioClient.FPutObject(ctx, bucket, obj, "10kb_file", minio.PutObjectOptions{}); err == nil {
		err = errors.New(fmt.Sprintf("Error: uploaded a file of size %d that exceeds the maximum user size %d %s", info.Size, userQuotas.MaxUserSize, profileUserQuotas.MaxUserSizeUnit))
		checkError(t, res, err)
	} else {
		if strings.Contains(err.Error(), QuotaExceeded) {
			log.AuctaLogger.Infof(fmt.Sprintf("Negative case pass: error trying to uploade file of size %d that exceeds maximum user size of %d %s", info.Size, userQuotas.MaxUserSize, profileUserQuotas.MaxUserSizeUnit))
		} else {
			checkError(t, res, err)
		}
	}

	log.AuctaLogger.Info("test maximum user object")
	numObjects := int(userQuotas.MaxUserObjects) + 1
	var objectNames []string
	for i := 0; i < numObjects; i++ {
		obj = fmt.Sprintf("object-%d", i)
		if _, err = minioClient.FPutObject(ctx, bucket, obj, "1kb_file", minio.PutObjectOptions{}); err == nil {
			log.AuctaLogger.Infof("Object: %s correctly uploaded", obj)
			objectNames = append(objectNames, obj)
		} else {
			if strings.Contains(err.Error(), QuotaExceeded) {
				log.AuctaLogger.Infof(fmt.Sprintf("Negative case pass: error, maximum number of user objects [%d] exceeded", userQuotas.MaxUserObjects))
			} else {
				checkError(t, res, err)
			}
		}
	}

	log.AuctaLogger.Info("removing created objects and buckets")
	for _, obj = range objectNames {
		log.AuctaLogger.Infof("Removing object: %s", obj)
		err = minioClient.RemoveObject(context.Background(), bucket, obj, minio.RemoveObjectOptions{})
		checkError(t, res, err)
	}
	log.AuctaLogger.Infof("Removing bucket: %s", bucket)
	err = minioClient.RemoveBucket(context.Background(), bucket)
	checkError(t, res, err)

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
			msg := fmt.Sprintf("False positive: The user is not supposed to be able to %s", operation)
			checkError(t, res, errors.New(msg))
		}
	} else {
		if permission {
			msg := fmt.Sprintf("%v", err)
			checkError(t, res, errors.New(msg))
		} else if strings.Contains(err.Error(), AccessDenied) {
			log.AuctaLogger.Infof("Negative case pass: %s, denied", operation)
		} else if !strings.Contains(err.Error(), AccessDenied) {
			msg := fmt.Sprintf("%v", err)
			checkError(t, res, errors.New(msg))
		}
	}
}

func createAppCredentialProfileCeph(t *testing.T, res *m.TestResult, profile *s3.S3Profile, appId uint64) {
	serviceType := "ceph"

	appCredential := authentication.AuthProfile{
		Name:          fmt.Sprintf("%s-%s", profile.Username, serviceType),
		Type:          serviceType,
		ApplicationId: appId,
		Profile:       profile,
		Active:        true}

	log.AuctaLogger.Infof("Creating the ceph profile %v", appCredential)

	var err error
	_, err = Pcc.CreateAppCredentialProfileCeph(&appCredential)
	checkError(t, res, err)

	timeout := time.After(5 * time.Minute)
	tick := time.Tick(15 * time.Second)

	for {
		select {
		case <-timeout:
			msg := "Timed out waiting for App Credential"
			checkError(t, res, errors.New(msg))
		case <-tick:
			acs, err := Pcc.GetAppCredentials("ceph")
			if err != nil {
				msg := fmt.Sprintf("Failed to get deploy status "+
					"%v", err)
				checkError(t, res, errors.New(msg))
			}

			for _, ac := range acs {
				if ac.Name == fmt.Sprintf("%s-%s", profile.Username, serviceType) && ac.DeployStatus == "completed" {
					jsonString, _ := json.Marshal(ac.Profile)
					json.Unmarshal(jsonString, profile)
					log.AuctaLogger.Infof("Created the ceph profile %v", profile)
					return
				}
			}
		}
	}
}

func createBucket(t *testing.T, res *m.TestResult, ctx context.Context, bucketName string) (exists bool, err error) {

	if exists, err = minioClient.BucketExists(ctx, bucketName); err == nil {
		if !exists {
			err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		} else {
			log.AuctaLogger.Warnf("Bucket %s already exists", bucketName)
		}
	} else {
		checkError(t, res, err)
	}
	return
}
