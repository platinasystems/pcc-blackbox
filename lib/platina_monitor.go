package pcc

import (
	"errors"
	"fmt"
	"github.com/platinasystems/pcc-models/monitor"
	"github.com/platinasystems/platina-monitor/models"
)

type QueryResponse struct {
	Data []monitor.MonitorResponse `json:"Data"`
}

type RulesResponse struct {
	Data []models.Rule `json:"Data"`
}

type RuleResponse struct {
	Data models.Rule `json:"Data"`
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

func (pcc *PccClient) GetRuleByName(ruleName string) (rule *models.Rule, err error) {
	var response RulesResponse
	err = pcc.Get("platina-monitor/alerts/rules", &response)
	if err != nil {
		return
	}
	for _, r := range response.Data {
		if r.Name == ruleName {
			rule = &r
			return
		}
	}
	rule = nil
	err = errors.New("Could not find rule")
	return
}

func (pcc *PccClient) UpdateRule(rule *models.Rule, ruleID uint64) (updated *models.Rule, err error) {
	var response RuleResponse
	err = pcc.Put(fmt.Sprintf("platina-monitor/alerts/rules/%d", ruleID), rule, &response)
	updated = &(response.Data)
	return
}
