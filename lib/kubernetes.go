// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
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
	ID          uint64
	Name        string     `json:"name" validate:"required"`
	K8sVersion  string     `json:"k8sVersion" validate:"required"`
	CniPlugin   string     `json:"cniPlugin" validate:"required"`
	Nodes       []K8sNodes `json:"nodes"`
	Pools       []*int     `json:"-" gorm:"-"`
	ControlCIDR string     `json:"controlCIDR" validate:"required,cidrv4"`
	IgwPolicy   string     `json:"igwPolicy" validate:"required"`
}

type K8sNodes struct {
	ID uint64 `json:"id"`
}

type KClusterUpdateRequest struct {
	RolePolicy string     `json:"rolePolicy"`
	ToAdd      []K8sNodes `json:"toAdd"`
	ToRemove   []K8sNodes `json:"toRemove"`
}

type K8sCluster struct {
	models.KCluster
}

func (p *PccClient) CreateKubernetes(k8sReq K8sClusterRequest) (err error) {
	err = p.Post("pccserver/kubernetes", &k8sReq, nil)
	return
}

func (p *PccClient) UpdateKubernetes(id uint64, k8sUpdate KClusterUpdateRequest) (err error) {
	fmt.Printf("debug UpdateKubernetes: id %v update %v\n", id, k8sUpdate)
	endpoint := fmt.Sprintf("pccserver/kubernetes/%v", id)
	err = p.Put(endpoint, &k8sUpdate, nil)
	return
}

func (p *PccClient) GetKubernetes() (clusters []K8sCluster, err error) {
	err = p.Get("pccserver/kubernetes", &clusters)
	return
}

func (p *PccClient) GetKubernetesId(id uint64) (cluster K8sCluster, err error) {
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
	percent = cluster.AnsibleJob.ProgressPercentage
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
	err = p.Delete(endpoint, &req)
	return
}
