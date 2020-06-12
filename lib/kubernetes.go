// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/platinasystems/tiles/pccserver/executor"
	"github.com/platinasystems/tiles/pccserver/kubernetes"
	"github.com/platinasystems/tiles/pccserver/models"
)

const (
	K8S_DEPLOY_STATUS_PROGRESS     = kubernetes.DEPLOY_STATUS_PROGRESS
	K8S_DEPLOY_STATUS_COMPLETED    = kubernetes.DEPLOY_STATUS_COMPLETED
	K8S_DEPLOY_STATUS_FAILED       = kubernetes.DEPLOY_STATUS_FAILED
	K8S_DEPLOY_APP_STATUS_PROGRESS = kubernetes.DEPLOY_APP_STATUS_PROGRESS
)

type K8sClusterRequest struct {
	ID               uint64
	Name             string     `json:"name" validate:"required"`
	K8sVersion       string     `json:"k8sVersion" validate:"required"`
	CniPlugin        string     `json:"cniPlugin" validate:"required"`
	Nodes            []K8sNodes `json:"nodes"`
	Pools            []*int     `json:"-" gorm:"-"`
	NetworkClusterId uint64     `json:"networkClusterID" gorm:"network_cluster_id"`
}

type K8sNodes struct {
	ID uint64
}

type K8sCluster struct {
	models.KCluster
}

func (p *PccClient) CreateKubernetes(k8sReq K8sClusterRequest) (err error) {
	err = p.Post("pccserver/kubernetes", &k8sReq, nil)
	return
}

func (p *PccClient) GetKubernetes() (clusters []K8sCluster, err error) {
	err = p.Get("pccserver/kubernetes", &clusters)
	return
}

func (p *PccClient) GetKubernetesId(id uint64) (cluster K8sCluster, err error) {
	cluster = K8sCluster{
		models.KCluster{Task: &executor.Task{}},
	}
	endpoint := fmt.Sprintf("pccserver/kubernetes/%v", id)
	err = p.Get(endpoint, &cluster)
	return
}

func (p *PccClient) FindKubernetesId(name string) (id uint64, err error) {

	var clusters []K8sCluster
	if clusters, err = p.GetKubernetes(); err != nil {
		return
	}
	for _, c := range clusters {
		if c.Name == name {
			id = c.ID
			return
		}
	}
	err = fmt.Errorf("Kubernetes cluster not found")
	return
}

func (p *PccClient) GetKubernetesDeployStatus(id uint64) (status string,
	percent int8, err error) {

	var cluster K8sCluster

	if cluster, err = p.GetKubernetesId(id); err != nil {
		return
	}
	status = cluster.DeployStatus
	if clusterTask, ok := (cluster.Task).(*executor.Task); ok {
		percent = clusterTask.Progress
	}
	return
}

func (p *PccClient) GetKubernetesHealth(id uint64) (health string, err error) {

	var cluster K8sCluster

	if cluster, err = p.GetKubernetesId(id); err != nil {
		return
	}
	health = cluster.HealthStatus
	return
}

func (p *PccClient) DeleteKubernetes(id uint64, force bool) (err error) {
	type delK8Req struct {
		forceRemove bool
	}

	endpoint := fmt.Sprintf("pccserver/kubernetes/%v", id)
	req := delK8Req{forceRemove: force}
	err = p.Delete(endpoint, &req, nil)
	return
}

type K8sStorageClassType string

const (
	K8S_STORAGE_CLASS_TYPE_CEPHFS   K8sStorageClassType = "cephfs"
	K8S_STORAGE_CLASS_TYPE_CEPH_RBD K8sStorageClassType = "rbd"
)

const (
	K8S_CLUSTER_NAME_PREFIX         = "k8sTest"
	K8S_STORAGE_CLASS_NAME_CEPHFS   = "sc-%d-%s-cephfs-provisioner"
	K8S_STORAGE_CLASS_NAME_CEPH_RBD = "sc-%d-%s-rbd-provisioner"
)
const (
	K8S_APP_DEPLOYMENT_TIMEOUT         = 1000
	K8S_APP_UNDEPLOYMENT_TIMEOUT       = 1000
	K8S_STORAGE_CLASS_CREATION_TIMEOUT = 1000
	K8S_STORAGE_CLASS_DELETION_TIMEOUT = 1000
)

const (
	K8S_APP_DEPLOY_EVENT             = "k8sAppDeploy"
	K8S_APP_UNDEPLOY_EVENT           = "k8sAppUnDeploy"
	K8S_STORAGE_CLASS_CREATION_EVENT = "storageClassCreation"
	K8S_STORAGE_CLASS_DELETION_EVENT = "storageClassUndeploy"
)

const (
	K8S_APP_DEPLOYMENT_SUCCESS_NOTIFICATION         = "Deploy for applications [%+v] on cluster [%s] with ID [%d] has been executed correctly"
	K8S_APP_DEPLOYMENT_FAILED_NOTIFICATION          = "Deploy for applications [%+v] on cluster [%s] with ID [%d] has failed"
	K8S_APP_UNDEPLOYMENT_SUCCESS_NOTIFICATION       = "Undeploy for applications with ID [%+v] on cluster [%s] with ID [%d] has been undeployed correctly"
	K8S_APP_UNDEPLOYMENT_FAILED_NOTIFICATION        = "Undeploy for applications with ID [%+v] on cluster [%s] with ID [%d] has failed"
	K8S_STORAGE_CLASS_CREATION_SUCCESS_NOTIFICATION = "Deploy for storageclasses on cluster [%s] with ID [%d] has been executed correctly"
	K8S_STORAGE_CLASS_CREATION_FAILED_NOTIFICATION  = "Deploy for storageclasses on cluster [%s] with ID [%d] has failed"
	K8S_STORAGE_CLASS_DELETION_SUCCESS_NOTIFICATION = "Deploy for storageclasses on cluster [%s] with ID [%d] has been executed correctly"
	K8S_STORAGE_CLASS_DELETION_FAILED_NOTIFICATION  = "Deploy for storageclasses on cluster [%s] with ID [%d] has failed"
)

const (
	TestCreateK8sCluster      = "createK8sCluster"
	TestCreateK8sStorageClass = "createK8sStorageClass"
	TestDeployK8sApp          = "deployApp"
	TestUndeployK8sApp        = "undeployApp"
	TestDeleteK8sStorageClass = "deleteK8sStorageClass"
	TestDeleteK8sCluster      = "deleteK8sCluster"
)

var (
	K8S_CLUSTER_NOT_FOUND_ERROR = "Kubernetes cluster[%v] not found"
)

type DeployAppRequest struct {
	Apps []*models.KApp `json:"-"`
}

func (p *PccClient) DeployK8sApp(request DeployAppRequest, k8sClusterId uint64) (err error) {
	endpoint := fmt.Sprintf("pccserver/kubernetes/%v/app", k8sClusterId)
	err = p.Post(endpoint, request.Apps, nil)

	return
}

type UndeployAppRequest struct {
	AppIds []uint64 `json:"appIds"`
}

func (p *PccClient) UnDeployK8sApp(request UndeployAppRequest, k8sClusterId uint64) (err error) {
	endpoint := fmt.Sprintf("pccserver/kubernetes/%v/app", k8sClusterId)
	err = p.Delete(endpoint, &request, nil)
	return
}

// Storage Class Config For Ceph Cluster
type CephStorageClassConfig struct {
	K8sVersion string   `json:"k8sVersion" validate:"required"`
	CniPlugin  string   `json:"-"`
	PoolIds    []uint64 `json:"pools"`
}

func (p *PccClient) CreateCephStorageClass(request CephStorageClassConfig, k8sClusterId uint64) (err error) {
	endpoint := fmt.Sprintf("pccserver/kubernetes/%v/upgrade", k8sClusterId)
	err = p.Post(endpoint, &request, nil)
	return
}

// TODO: Generic Storage Class Config
type StorageClassConfig struct {
	Name string `json:"-"`
}

// TODO: Generic way to create Storage Class
func (p *PccClient) CreateStorageClass(request StorageClassConfig, k8sClusterId uint64, poolName string) (err error) {
	return
}

type DeleteStorageClassRequest struct {
	StorageclassIds []uint64 `json:"ids"`
}

func (p *PccClient) DeleteStorageClass(request DeleteStorageClassRequest, k8sClusterId uint64) (err error) {
	endpoint := fmt.Sprintf("pccserver/kubernetes/cluster/%v/strgclasses", k8sClusterId)
	err = p.Delete(endpoint, &request, nil)
	return
}

func (p *PccClient) GetStorageClassId(clusterId uint64, name string) (storageClassId uint64, err error) {
	var storageClasses []models.K8sStorageClass
	if storageClasses, err = p.GetAllStorageClasses(clusterId); err != nil {
		return
	}
	for _, storageClass := range storageClasses {
		if strings.Contains(storageClass.ScName, name) {
			storageClassId = storageClass.Id
			break
		}
	}
	if storageClassId == 0 {
		err = fmt.Errorf("No Storage Class found: %v", name)
	}
	return
}

func (p *PccClient) GetAllStorageClasses(id uint64) (storageClasses []models.K8sStorageClass, err error) {
	endpoint := fmt.Sprintf("pccserver/kubernetes/cluster/%v/strgclasses", id)
	err = p.Get(endpoint, &storageClasses)

	return
}

func (p *PccClient) GetK8sAppId(appName string, k8sClusterId uint64) (appId uint64, err error) {
	var cluster K8sCluster
	if cluster, err = p.GetKubernetesId(k8sClusterId); err != nil {
		return
	}
	for _, app := range cluster.Apps {
		if appName == app.Label {
			appId = app.ID
			break
		}
	}
	return
}

func (p *PccClient) GetKubernetesClusterByName(name string) (cluster *K8sCluster, err error) {
	var clusters []K8sCluster
	if clusters, err = p.GetKubernetes(); err == nil {
		for _, c := range clusters {
			if c.Name == name {
				cluster = &c
				return
			}
		}
		if cluster == nil {
			err = fmt.Errorf(K8S_CLUSTER_NOT_FOUND_ERROR, name)
		}
	}
	return
}

type K8sAppVerifier struct {
	timeout   time.Duration
	events    EventsToCheck
	breakLoop chan bool
}

func (v *K8sAppVerifier) GetTimeout() time.Duration {
	return v.timeout
}

func (v *K8sAppVerifier) GetEventsToCheck() EventsToCheck {
	return v.events
}

func (v *K8sAppVerifier) GetBreakLoopChan() chan bool {
	return v.breakLoop
}

type K8sAppConfiguration struct {
	Apps                []ConfigKApp `json:"apps"`
	K8sClusterName      string       `json:"k8sClusterName"`
	K8sCluster          *K8sCluster
	Tests               map[string]bool `json:"tests"`
	PccClient           *PccClient
	AppIds              []uint64
	CephStorageRequired bool `json:"cephStorageRequired"`
	StorageClasses      map[string]uint64
}

type StorageClassHelmConfig struct {
	Type                         string `json:"type"`
	StorageClassHelmVariableName string `json:"storageClassHelmVariableName"`
}

type ConfigKApp struct {
	models.KApp
	StorageClassHelmConfig []StorageClassHelmConfig `json:"storageClassHelmConfig"`
	HelmFilePath           string                   `json:"helmFilePath"`
}

func (config *K8sAppConfiguration) GetK8sClusterName() string {
	return config.K8sClusterName
}

func (config *K8sAppConfiguration) GetK8sClusterId() uint64 {
	return config.K8sCluster.ID
}

func (config *K8sAppConfiguration) GetAppIds() []uint64 {
	return config.AppIds
}

func (p *PccClient) ValidateAppConfig(config *K8sAppConfiguration, identifier string) (err error) {
	var cluster *K8sCluster
	if config.K8sClusterName != "" {
		cluster, err = p.GetKubernetesClusterByName(config.K8sClusterName)
		if err == nil {
			config.K8sCluster = cluster
		} else if err.Error() == fmt.Sprintf(K8S_CLUSTER_NOT_FOUND_ERROR, config.K8sClusterName) {
			if run, ok := config.Tests[TestCreateK8sCluster]; ok && run {
				err = nil
			}
		}
	} else {
		config.K8sClusterName = fmt.Sprintf("%s_%s", K8S_CLUSTER_NAME_PREFIX, identifier)
	}
	if len(config.StorageClasses) == 0 {
		config.StorageClasses = make(map[string]uint64)
	}
	return
}

func (config *K8sAppConfiguration) ParseAndEncode(app ConfigKApp) (string, error) {
	content, err := ioutil.ReadFile(app.HelmFilePath)
	if err != nil {
		return "", err
	}
	if len(app.StorageClassHelmConfig) > 0 {
		output := string(content)
		for _, s := range app.StorageClassHelmConfig {
			var sc_name string
			regex := regexp.MustCompile(s.StorageClassHelmVariableName + `:.*`)
			for name, _ := range config.StorageClasses {
				if strings.Contains(name, s.Type) {
					sc_name = name
					break
				}
			}
			if sc_name == "" {
				return "", fmt.Errorf("Failed to update Helm Values File with proper StorageClass name")
			}
			output = regex.ReplaceAllString(output, s.StorageClassHelmVariableName+": "+sc_name)
		}
		content = []byte(output)
	}
	return base64.StdEncoding.EncodeToString(content), nil
}

func (config *K8sAppConfiguration) GetAllAppNamesString() (appsName string) {
	temp := []string{}
	for _, app := range config.Apps {
		temp = append(temp, app.Label)
	}
	appsName = "[" + strings.Join(temp, " ") + "]"
	return
}

func (config *K8sAppConfiguration) VerifyK8sApp(startTime time.Time, action string, name string) (s EventStatus, err error) {
	s = config.PccClient.Verify(startTime, config.getK8sAppVerifier(action, name))

	failed := !(strings.Contains(s.Msg, fmt.Sprintf(K8S_APP_DEPLOYMENT_SUCCESS_NOTIFICATION, config.GetAllAppNamesString(), config.GetK8sClusterName(), config.K8sCluster.ID)) ||
		strings.Contains(s.Msg, fmt.Sprintf(K8S_APP_UNDEPLOYMENT_SUCCESS_NOTIFICATION, config.AppIds, config.GetK8sClusterName(), config.K8sCluster.ID)) ||
		strings.Contains(s.Msg, fmt.Sprintf(K8S_STORAGE_CLASS_CREATION_SUCCESS_NOTIFICATION, config.GetK8sClusterName(), config.K8sCluster.ID)) ||
		strings.Contains(s.Msg, fmt.Sprintf(K8S_STORAGE_CLASS_DELETION_SUCCESS_NOTIFICATION, config.GetK8sClusterName(), config.K8sCluster.ID)))

	if s.IsError || failed {
		err = fmt.Errorf("%v", s.Msg)
	}
	return
}

func (config *K8sAppConfiguration) getK8sAppVerifier(action string, name string) (v *K8sAppVerifier) {
	switch action {
	case K8S_APP_DEPLOY_EVENT:
		v = &K8sAppVerifier{
			timeout: K8S_APP_DEPLOYMENT_TIMEOUT,
			events: EventsToCheck{
				fmt.Sprintf(K8S_APP_DEPLOYMENT_SUCCESS_NOTIFICATION, config.GetAllAppNamesString(), config.GetK8sClusterName(), config.K8sCluster.ID): true,
				fmt.Sprintf(K8S_APP_DEPLOYMENT_FAILED_NOTIFICATION, config.GetAllAppNamesString(), config.GetK8sClusterName(), config.K8sCluster.ID):  true,
			},
			breakLoop: make(chan bool),
		}
	case K8S_APP_UNDEPLOY_EVENT:
		v = &K8sAppVerifier{
			timeout: K8S_APP_UNDEPLOYMENT_TIMEOUT,
			events: EventsToCheck{
				fmt.Sprintf(K8S_APP_UNDEPLOYMENT_SUCCESS_NOTIFICATION, config.AppIds, config.GetK8sClusterName(), config.K8sCluster.ID): true,
				fmt.Sprintf(K8S_APP_UNDEPLOYMENT_FAILED_NOTIFICATION, config.AppIds, config.GetK8sClusterName(), config.K8sCluster.ID):  true,
			},
			breakLoop: make(chan bool),
		}
	case K8S_STORAGE_CLASS_CREATION_EVENT:
		v = &K8sAppVerifier{
			timeout: K8S_APP_DEPLOYMENT_TIMEOUT,
			events: EventsToCheck{
				fmt.Sprintf(K8S_STORAGE_CLASS_CREATION_SUCCESS_NOTIFICATION, config.GetK8sClusterName(), config.K8sCluster.ID): true,
				fmt.Sprintf(K8S_STORAGE_CLASS_CREATION_FAILED_NOTIFICATION, config.GetK8sClusterName(), config.K8sCluster.ID):  true,
			},
			breakLoop: make(chan bool),
		}
	case K8S_STORAGE_CLASS_DELETION_EVENT:
		v = &K8sAppVerifier{
			timeout: K8S_APP_DEPLOYMENT_TIMEOUT,
			events: EventsToCheck{
				fmt.Sprintf(K8S_STORAGE_CLASS_DELETION_SUCCESS_NOTIFICATION, config.GetK8sClusterName(), config.K8sCluster.ID): true,
				fmt.Sprintf(K8S_STORAGE_CLASS_DELETION_FAILED_NOTIFICATION, config.GetK8sClusterName(), config.K8sCluster.ID):  true,
			},
			breakLoop: make(chan bool),
		}
	}
	return
}
