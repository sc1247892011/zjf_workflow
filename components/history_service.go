package components

import (
	"database/sql"
)

type HistoryService interface {
	//迁徙节点数据到历史表
	CopyNodeInstanceById(tx *sql.Tx, nodeId int) error

	CopyNodeInstance(tx *sql.Tx, nodeId int, processInstanceId int, processDefinitionName string, nodeName string, executionId string,
		previousExecutionId string, assignee string) (int, error)

	//流程进度查询接口
	GetProcessCompleteTask(ProcessInstanceId int) ([]map[string]interface{}, error)
	GetTransaction() (*sql.Tx, error)
}
