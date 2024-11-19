package components

import (
	"database/sql"
	"time"
)

// NodeInstance 定义了节点实例的数据结构
type NodeInstance struct {
	Id                    int
	ProcessInstanceId     int
	ProcessDefinitionName string
	NodeName              string
	ExecutionId           string // 修改为 string，与数据库中的 VARCHAR(50) 对应
	OutputData            string // 可以使用 string 存储 JSON 数据
	PreviousExecutionId   string // 修改为 string，与数据库中的 VARCHAR(50) 对应
	Assignee              string // 节点的负责人 (网关 和 序列流 负责人为空)
	StartTime             time.Time
	EndTime               time.Time
}

// TaskRuntimeService 提供了操作节点实例的接口
type NodeService interface {
	//获取事务
	GetTransaction() (*sql.Tx, error)
	//初始化工作流节点 插入数据库 返回自增id
	InitNodeInstance(tx *sql.Tx, processInstanceId int, ProcessDefinitionName string, nodeName string, executionId string, previousExecutionId string, assignee string) (int, error)
	GetAttributeByExpression(tx *sql.Tx, expression string, processInstanceId int) (map[string]interface{}, error)
	//需要加锁防止并发情况下的
	CountParallelGatewayIncoming(tx *sql.Tx, processInstanceId int, executionId string) (int, error)
	UpdateNodeInstanceOutput(tx *sql.Tx, id int, outputData string) error
	GetAssigneeUndoneTask(assignee string) ([]map[string]interface{}, error)
	GetTaskDetailByTaskId(taskId int) (map[string]interface{}, error)
	GetTaskForm(processDefinitionName string, executionId string) (string, error)
	ClearProcessData(tx *sql.Tx, processInstanceId int) error
}
