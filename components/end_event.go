package components

import "log"

type EndEvent struct {
	ExecutionId string `xml:"executionId,attr"` // 绑定 id 属性
	Name        string `xml:"name,attr"`
	Incoming    string `xml:"Incoming"` // 绑定 <Incoming> 子元素
}

// 方法接收器是 *StartEvent，允许修改 StartEvent 的字段
func (endEvent EndEvent) Execute(ctx *WorkflowContext) {
	//更新数据库流程实例状态 迁移数据到历史表
	nodeService := GetServiceFactory().GetNodeService()
	nodeId, initerr := nodeService.InitNodeInstance(ctx.ProcessInstanceId, endEvent.Name, endEvent.ExecutionId, ctx.CurrentExecutionId, "")
	if initerr != nil {
		log.Println("Failed to insert endEvent to database: ", initerr)
	}
	//迁徙数据到历史库
	historyService := GetServiceFactory().GetHistoryService()
	historyService.CopyNodeInstanceById(nodeId)
}
