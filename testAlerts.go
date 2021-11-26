package main

import (
	"errors"
	"fmt"
	"github.com/minio/minio-go/v7"
	log "github.com/platinasystems/go-common/logs"
	m "github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/pcc-models/s3"
	"github.com/platinasystems/test"
	"testing"
	"time"
)

func testPoolUsageAlert(t *testing.T) {
	/* t.Run("createPoolRGW", createPoolRGW)
	t.Run("verifyPool", verifyPool)
	if t.Failed() {
		return
	}
	t.Run("installRGW", installRGW)
	t.Run("verifyRGW", verifyRGWDeploy)
	if t.Failed() {
		return
	}
	t.Run("createFullCephProfile", createFullCephProfile)
	t.Run("addCephProfile", addCephCredential)
	*/
	t.Run("testCreateBucket2", testCreateBucket2)
	t.Run("testEmptyPoolPrometheus", testEmptyPoolPrometheus)
	t.Run("addFileOver80", addFileOver80)
	t.Run("testEmptyPoolPrometheus", testPoolOver80Prometheus)

}

func testCreateBucket2(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	node, err := Pcc.GetNode(3)

	minioClient, err = initS3Client(node.Host, "QKZBMZCB3250D3PODPRW", "BI0hFg9cUyLhui0X7guV4QWIoS4rnBJQC2WRFcxY")
	checkError(t, res, err)

	bucketName = fmt.Sprintf("bucket-%s", "blackbox-full-ceph")

	createBucket(t, res, ctx, bucketName)
	buckets, err := minioClient.ListBuckets(ctx)
	log.AuctaLogger.Info(buckets)

}

func createFullCephProfile(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	rwdPermission := Permissions{ReadPermission: true, WritePermission: true, DeletePermission: true}

	profileRGW = s3.S3Profile{
		Username:         "blackbox-full",
		ReadPermission:   &rwdPermission.ReadPermission,
		WritePermission:  &rwdPermission.WritePermission,
		DeletePermission: &rwdPermission.DeletePermission,
	}
}

func testEmptyPoolPrometheus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	usage := checkPoolUsagePrometheus(t, res)
	log.AuctaLogger.Infof("usage: %v", usage)
	var err error
	if usage != 0 {
		err = errors.New("Pool should be empty")
	}
	checkError(t, res, err)
}

func testPoolOver80Prometheus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	time.Sleep(1 * time.Minute)

	usage := checkPoolUsagePrometheus(t, res)
	log.AuctaLogger.Infof("usage: %v", usage)
	var err error

	if usage < 80 {
		err = errors.New("Pool usage should be over 80%")
	}
	checkError(t, res, err)
}

func checkPoolUsagePrometheus(t *testing.T, res *m.TestResult) int {
	/*
		client, err := api.NewClient(api.Config{

			Address: fmt.Sprintf("http://%s:9090", Env.PrometheusIp),
		})
		checkError(t, res, err)

		v1api := v1.NewAPI(client)
		ctxProm, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	*/
	query := fmt.Sprintf("pools:%s:quotaUsage", Env.RGWConfiguration.PoolName)

	result, err := Pcc.InstantQuery(query)
	checkError(t, res, err)

	log.AuctaLogger.Infof("result: %v", result.Value)
	return int(result.Value)
}

func addFileOver80(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	_, err := createFile("85MB_file", 85*1e6)
	checkError(t, res, err)

	_, err = minioClient.FPutObject(ctx, bucketName, "85MB_file", "85MB_file", minio.PutObjectOptions{})
	checkError(t, res, err)
}
