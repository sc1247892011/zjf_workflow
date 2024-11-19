package components

import (
	"database/sql"
	"time"
)

type WorkflowContext struct {
	Model                 *Model // 工作流模型对象
	ProcessInstanceId     int    // 流程实例id 唯一标识每个流程实例
	ProcessDefinitionName string // 流程定义名称
	// Version               int       // 流程定义版本
	// BusinessKey           string    // 业务标识符
	CurrentUserId      string    // 当前操作的用户Id
	CurrentExecutionId string    // 当前执行的任务（节点，网关，序列流）结构Id
	Data               string    // 流程节点数据json
	StartTime          time.Time // 工作流启动时间
	Tx                 *sql.Tx   // 当前事务
}
