package components

import (
	"time"
)

// NodeInstance 定义了节点实例的数据结构
type NodeInstance struct {
	Id                  int
	ProcessInstanceId   int
	NodeName            string
	ExecutionId         string // 修改为 string，与数据库中的 VARCHAR(50) 对应
	OutputData          string // 可以使用 string 存储 JSON 数据
	PreviousExecutionId string // 修改为 string，与数据库中的 VARCHAR(50) 对应
	Assignee            string // 节点的负责人 (网关 和 序列流 负责人为空)
	StartTime           time.Time
	EndTime             time.Time
}

// TaskRuntimeService 提供了操作节点实例的接口
type NodeService interface {
	//初始化工作流节点 插入数据库 返回自增id
	InitNodeInstance(processInstanceId int, nodeName string, executionId string, previousExecutionId string, assignee string) (int, error)
	GetAttributeByExpression(expression string, processInstanceId int) (map[string]interface{}, error)
	CountParallelGatewayIncoming(processInstanceId int, executionId string) (int, error)
	UpdateNodeInstanceOutput(id int, outputData string) error
}
