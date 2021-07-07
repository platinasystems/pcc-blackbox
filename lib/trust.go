package pcc

import (
	"encoding/json"
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

func (pcc *PccClient) GetTrusts() (trusts []security.Trust, err error) {
	err = pcc.Get("pccserver/trusts", &trusts)
	return
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

func (pcc *PccClient) SecondaryEndedRemoteTrustCreation(appType string, slaveParams SlaveParametersRGW, filePath string) (trust security.Trust, err error) {
	endPoint := "pccserver/trusts"
	strParams, _ := json.Marshal(slaveParams)

	m := map[string]string{
		"side":        "slave",
		"appType":     appType,
		"slaveParams": string(strParams),
	}
	fileMap := map[string]string{"trustFile": filePath}
	// this is actually a POST
	err = pcc.PutFiles("POST", endPoint, fileMap, m, &trust)
	return
}

func (pcc *PccClient) DeleteTrust(id uint64) (result security.Trust, err error) {
	err = pcc.Delete(fmt.Sprintf("pccserver/trusts/%d", id), nil, &result)
	return
}

func (pcc *PccClient) DeleteTrusts() (result security.Trust, err error) {
	err = pcc.Delete("pccserver/trusts/", nil, &result)
	return
}

func (pcc *PccClient) TrustExists(id uint64) (found bool, err error) {
	var trusts []security.Trust
	trusts, err = pcc.GetTrusts()
	if err != nil {
		return
	}
	for _, trust := range trusts {
		if trust.ID == id {
			found = true
			return
		}
	}
	return
}
