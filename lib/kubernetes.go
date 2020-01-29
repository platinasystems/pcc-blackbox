package pcc

import (
	"encoding/json"
	"fmt"

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
	ID         uint64
	Name       string     `json:"name" validate:"required"`
	K8sVersion string     `json:"k8sVersion" validate:"required"`
	CniPlugin  string     `json:"cniPlugin" validate:"required"`
	Nodes      []K8sNodes `json:"nodes"`
}

type K8sNodes struct {
	ID uint64
}

type K8sCluster struct {
	models.KCluster
}

func (p PccClient) CreateKubernetes(k8sReq K8sClusterRequest) (err error) {
	var (
		body []byte
		data []byte
		resp HttpResp
	)
	endpoint := fmt.Sprintf("pccserver/kubernetes")
	if data, err = json.Marshal(k8sReq); err != nil {
		err = fmt.Errorf("invalid struct for K8s creation")
		return
	}
	if resp, body, err = p.pccGateway("POST", endpoint, data); err != nil {
		err = fmt.Errorf("%v\n%v\n", string(body), err)
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("K8s cretion failed:\n%v\n", string(body))
		return
	}
	return
}

func (p PccClient) GetKubernetes() (clusters []K8sCluster, err error) {
	var (
		body []byte
		resp HttpResp
	)
	endpoint := fmt.Sprintf("pccserver/kubernetes")
	if resp, body, err = p.pccGateway("GET", endpoint, nil); err != nil {
		err = fmt.Errorf("%v\n%v\n", string(body), err)
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("K8s get failed:\n%v\n", string(body))
		return
	}
	err = json.Unmarshal(resp.Data, &clusters)
	return
}

func (p PccClient) GetKubernetesId(id uint64) (cluster K8sCluster, err error) {
	var (
		body []byte
		resp HttpResp
	)
	endpoint := fmt.Sprintf("pccserver/kubernetes/%v", id)
	if resp, body, err = p.pccGateway("GET", endpoint, nil); err != nil {
		err = fmt.Errorf("%v\n%v\n", string(body), err)
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("K8s get failed:\n%v\n", string(body))
		return
	}
	err = json.Unmarshal(resp.Data, &cluster)
	return
}

func (p PccClient) FindKubernetesId(name string) (id uint64, err error) {

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

func (p PccClient) GetKubernetesDeployStatus(id uint64) (status string,
	percent int8, err error) {

	var cluster K8sCluster

	if cluster, err = p.GetKubernetesId(id); err != nil {
		return
	}
	status = cluster.DeployStatus
	percent = cluster.AnsibleJob.ProgressPercentage
	return
}

func (p PccClient) GetKubernetesHealth(id uint64) (health string, err error) {

	var cluster K8sCluster

	if cluster, err = p.GetKubernetesId(id); err != nil {
		return
	}
	health = cluster.HealthStatus
	return
}

func (p PccClient) DeleteKubernetes(id uint64, force bool) (err error) {

	var (
		data []byte
		body []byte
		resp HttpResp
	)
	type delK8Req struct {
		forceRemove bool
	}

	req := delK8Req{forceRemove: force}
	if data, err = json.Marshal(req); err != nil {
		return
	}

	endpoint := fmt.Sprintf("pccserver/kubernetes/%v", id)
	resp, body, err = p.pccGateway("DELETE", endpoint, data)
	if err != nil {
		err = fmt.Errorf("%v\n%v\n", string(body), err)
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("K8s delete failed:\n%v\n", string(body))
		return
	}
	return
}