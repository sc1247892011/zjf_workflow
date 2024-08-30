package components

import (
	"time"
)

// ProcessInstance 定义了流程实例的数据结构
type ProcessInstance struct {
	Id                    int    //数据库自增主键
	ProcessDefinitionName string //流程定义名称
	Version               int    //流程定义版本
	Business_key          string //业务键
	Status                string
	CreatedBy             string
	StartTime             time.Time
	EndTime               *time.Time
}

// RuntimeService 提供了操作流程实例的接口
type RuntimeService interface {
	StartProcessInstance(ProcessDefinitionName string, Business_key string, createdBy string, formParams string) (int, error)
}
