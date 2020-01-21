package pcc

import (
	"github.com/lib/pq"
	"github.com/platinasystems/tiles/pccserver/models"
	"encoding/json"
	"fmt"
	"time"
)
const (
	CEPH_CLUSTER_NAME = "cephtest"
	CEPH_3_NODE_INSTALLATION_TIMEOUT = 500
	CEPH_3_NODE_UNINSTALLATION_TIMEOUT = 100
	CEPH_INSTALLATION_SUCCESS_NOTIFICATION = "Ceph cluster has been deployed"
	CEPH_INSTALLATION_FAILED_NOTIFICATION = "Ceph cluster deployment failed"
	CEPH_UNINSTALLATION_SUCCESS_NOTIFICATION = "ceph cluster has been removed from DB"
	CEPH_UNINSTALLATION_FAILED_NOTIFICATION = "Ceph cluster uninstallation failed"
)

type CreateCephClusterRequest struct {
	Name              string         `gorm:"name" json:"name"`
	Nodes             []CephNodes    `json:"nodes" gorm:"foreignkey:ceph_cluster_id"`
	Version           string         `gorm:"version" json:"version"`
	Tags              pq.StringArray `gorm:"tags" json:"tags"`

	models.CephClusterConfig `gorm:"" json:"config"`
}

type CephNodes struct {
	ID uint64
}

func (p PccClient) CreateCephCluster(request CreateCephClusterRequest) (id int, err error) {
	var (
		body              []byte
		data              []byte
		resp              HttpResp
	)
	id = -1
	endpoint := fmt.Sprintf("pccserver/storage/ceph/cluster")
	if data, err = json.Marshal(request); err != nil {
		err = fmt.Errorf("Invalid struct for ceph creation..ERROR: %v", err)
	}else {
		if resp, body, err = p.pccGateway("POST", endpoint, data); err != nil {
			err = fmt.Errorf("%v\n%v\n", string(body), err)
		} else {
			if resp.Status != 200 {
				fmt.Printf("Ceph creation failed:\n%v\n", string(body))
				err = fmt.Errorf("%v\n", string(body))
			}
			time.Sleep(time.Second * 5)
			cluster, errGet := p.GetCephCluster(request.Name)
			if errGet == nil {
 				id = int(cluster.Id)
 			}
		}
	}
	return
}

func (p PccClient) GetCephCluster(name string) (*models.CephCluster, error){
	if clusters, err := p.GetAllCephClusters(); err != nil {
		return nil, err
	}else {
		for _, cluster := range clusters {
			if cluster.Name != CEPH_CLUSTER_NAME {
				continue
			} else {
				return cluster, nil
			}
		}
	}
	return nil, fmt.Errorf("Ceph cluster %v not found", name)
}

func (p PccClient) GetAllCephClusters() (clusterList []*models.CephCluster, err error){
	var (
		body              []byte
		data              []byte
		resp              HttpResp
	)
	endpoint := fmt.Sprintf("pccserver/storage/ceph/cluster")
	if resp, body, err = p.pccGateway("GET", endpoint, data); err != nil {
		return nil, fmt.Errorf("%v\n%v\n", string(body), err)
	}
	if resp.Status != 200 {
		fmt.Printf("Ceph status check failed:\n%v\n", string(body))
		return nil, fmt.Errorf("%v\n", string(body))
	}
	err = json.Unmarshal(resp.Data, &clusterList)
	if err != nil {
		return nil, fmt.Errorf("JSON unmarshal failed for status check..ERROR: %v", err)
	}
	return clusterList, nil
}

func (p PccClient) DeleteCephCluster(id uint64) (err error){
	var (
		body              []byte
		data              []byte
		resp              HttpResp
	)
	/*cluster, err := p.GetCephCluster(name)
	if err != nil {
		return err
	}*/
	endpoint := fmt.Sprintf("pccserver/storage/ceph/cluster/%v", id)
	if resp, body, err = p.pccGateway("DELETE", endpoint, data); err != nil {
		return fmt.Errorf("%v\n%v\n", string(body), err)
	}
	if resp.Status != 200 {
		fmt.Printf("Ceph status check failed:\n%v\n", string(body))
		return fmt.Errorf("%v\n", string(body))
	}
	return nil
}