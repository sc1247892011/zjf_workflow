package components

import "log"

type ExclusiveGateway struct {
	ExecutionId string   `xml:"executionId,attr"` // 绑定 id 属性
	Outgoing    []string `xml:"Outgoing"`         // 绑定 <Outgoing> 子元素
	Incoming    []string `xml:"Incoming"`         // 绑定 <Incoming> 子元素
}

// 方法接收器是 *StartEvent，允许修改 StartEvent 的字段
// 逻辑根并行网关是一样的 除了不需要等待所有的Incoming 到齐，触发一次就全部执行一次outgoing，因为互斥网关后所有的序列流是有条件表达式的，可以自行控制是否继续走
func (exclusiveGateway ExclusiveGateway) Execute(ctx *WorkflowContext) {
	//序列流进入该方法 记录入库
	nodeService := GetServiceFactory().GetNodeService()
	//StartNodeInstance(processInstanceId int, nodeName string, executionId string, previousExecutionId string, assignee string) (int, error)
	nodeId, initerr := nodeService.InitNodeInstance(ctx.ProcessInstanceId, "exclusiveGateway", exclusiveGateway.ExecutionId, ctx.CurrentExecutionId, "nobody")
	if initerr != nil {
		log.Println("Failed to insert exclusiveGateway to database: ", initerr)
	}

	//网关数据 只要插入一条 就往历史表里同步一条
	historyService := GetServiceFactory().GetHistoryService()
	historyService.CopyNodeInstanceById(nodeId)

	ctx.CurrentExecutionId = exclusiveGateway.ExecutionId
	for _, value := range exclusiveGateway.Outgoing {
		model := *(ctx.Model)
		model.SequenceFlows[value].Execute(ctx)
	}

}
