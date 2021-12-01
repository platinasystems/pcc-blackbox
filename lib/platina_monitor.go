package pcc

import (
	"github.com/platinasystems/pcc-models/monitor"
	"github.com/platinasystems/platina-monitor/models"
)

type QueryResponse struct {
	Data []monitor.MonitorResponse `json:"Data"`
}

type RulesResponse struct {
	Data []models.Rule `json:"Data"`
}

func (pcc *PccClient) InstantQuery(query string) (data *monitor.MonitorResponse, err error) {
	request := map[string]interface{}{"query": query}
	var response QueryResponse
	err = pcc.Post("platina-monitor/monitor/query", &request, &response)
	data = &(response.Data[0])
	return
}

func (pcc *PccClient) GetRules() (rules *[]models.Rule, err error) {
	var response RulesResponse
	err = pcc.Get("platina-monitor/alerts/rules", &response)
	rules = &(response.Data)
	return
}
