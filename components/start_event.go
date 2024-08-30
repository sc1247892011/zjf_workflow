package components

import (
	"log"
)

type StartEvent struct {
	ExecutionId string `xml:"executionId,attr"` // 绑定 id 属性
	Name        string `xml:"name,attr"`
	Outgoing    string `xml:"Outgoing"` // 绑定 <Outgoing> 子元素
	FormData    string `xml:"FormData"` // 绑定 <FormData> 子元素 用来给流程启动做前端页面展示
}

func (startEvent StartEvent) Execute(ctx *WorkflowContext) {
	//持久化 新建工作流的输入数据
	nodeService := GetServiceFactory().GetNodeService()
	nodeId, initerr := nodeService.InitNodeInstance(ctx.ProcessInstanceId, startEvent.Name, startEvent.ExecutionId, "", ctx.CurrentUserId)
	//迁徙数据到历史库
	historyService := GetServiceFactory().GetHistoryService()
	historyService.CopyNodeInstanceById(nodeId)
	if initerr != nil {
		log.Println("Failed to insert endEvent to database: ", initerr)
	}

	//流程运转
	ctx.CurrentExecutionId = startEvent.ExecutionId
	sequenceFlow := ctx.Model.SequenceFlows[startEvent.Outgoing]
	sequenceFlow.Execute(ctx)
}
