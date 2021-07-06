package pcc

import (
	"fmt"
	log "github.com/platinasystems/go-common/logs"
	"github.com/platinasystems/pcc-models/security"
)

type SlaveParametersRGW struct {
	ClusterID           uint64               `json:"clusterID"`
	ClusterName         string               `json:"clusterName"`
	AvailableNodes      []AvailableNode      `json:"availableNodes"`
	TargetNodes         []uint64             `json:"targetNodes"`
	FreeSpace           uint64               `json:"freeSpace"`
	ErasureCodeProfiles []ErasureCodeProfile `json:"erasureCodeProfiles"`
}

type AvailableNode struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type ErasureCodeProfile struct {
	DataChunks   uint64 `json:"dataChunks"`
	CodingChunks uint64 `json:"codingChunks"`
}

func (pcc *PccClient) GetTrust(id uint64) (result security.Trust, err error) {
	err = pcc.Get(fmt.Sprintf("pccserver/trusts/%d", id), &result)
	return
}

func (pcc *PccClient) GetTrustFile(id uint64) (result security.TrustFile, err error) {
	err = pcc.GetRaw(fmt.Sprintf("pccserver/trusts/%d/download", id), &result)
	return
}

func (pcc *PccClient) SelectTargetNodes(trust *security.Trust, id uint64) (result security.Trust, err error) {
	err = pcc.Put(fmt.Sprintf("pccserver/trusts/%d", id), *trust, &result)
	return
}

func (pcc *PccClient) StartRemoteTrustCreation(trust *security.Trust) (result security.Trust, err error) {
	err = pcc.Post("pccserver/trusts", *trust, &result)
	return
}

func (pcc *PccClient) PrimaryEndedRemoteTrustCreation(appType string, masterAppID uint64, filePath string) (trust security.Trust, err error) {
	endPoint := "pccserver/trusts"
	m := map[string]string{
		"side":        "master",
		"appType":     appType,
		"masterAppID": fmt.Sprintf("%d", masterAppID),
	}
	log.AuctaLogger.Info(m)
	fileMap := map[string]string{"trustFile": filePath}
	err = pcc.PutFiles("POST", endPoint, fileMap, m, &trust)
	return
}

func (pcc *PccClient) SecondaryEndedRemoteTrustCreation(appType string, slaveParams string, filePath string) (trust security.Trust, err error) {
	endPoint := "pccserver/trusts"
	m := map[string]string{
		"side":        "slave",
		"appType":     appType,
		"slaveParams": slaveParams,
	}
	fileMap := map[string]string{"trustFile": filePath}
	// this is actually a POST
	err = pcc.PutFiles("POST", endPoint, fileMap, m, &trust)
	return
}
