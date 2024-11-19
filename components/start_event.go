package components

import (
	"log"
)

type StartEvent struct {
	ExecutionId string `xml:"executionId,attr"` // 绑定 id 属性
	Name        string `xml:"name,attr"`
	Outgoing    string `xml:"Outgoing"` // 绑定 <Outgoing> 子元素
	FormData    string `xml:"FormData"` // 绑定 <FormData> 子元素 用来给流程启动做前端页面展示
	X           string `xml:"x,attr"`
	Y           string `xml:"y,attr"`
	H           string `xml:"h,attr"`
	W           string `xml:"w,attr"`
	Listener    string `xml:"Listener"` // 任务监听 执行完毕之后的后续逻辑 方法名称可以用逗号隔开 传递多段逻辑
}

func (startEvent StartEvent) Execute(ctx *WorkflowContext) {
	//持久化 新建工作流的输入数据
	nodeService := GetServiceFactory().GetNodeService()
	tx := ctx.Tx
	if tx == nil {
		log.Println("Failed to get transaction from ctx ")
		return
	}
	nodeId, initerr := nodeService.InitNodeInstance(tx, ctx.ProcessInstanceId, ctx.ProcessDefinitionName, startEvent.Name, startEvent.ExecutionId, "", ctx.CurrentUserId)
	//迁徙数据到历史库
	if initerr != nil {
		log.Println("Failed to insert startEvent to database: ", initerr)
		tx.Rollback()
		return
	}
	historyService := GetServiceFactory().GetHistoryService()
	_, he := historyService.CopyNodeInstance(tx, nodeId, ctx.ProcessInstanceId, ctx.ProcessDefinitionName, startEvent.Name, startEvent.ExecutionId, "", ctx.CurrentUserId)
	if he != nil {
		log.Println("Failed to insert startEvent to database: ", he)
		tx.Rollback()
		return
	}

	//流程运转
	ctx.Tx = tx
	ctx.CurrentExecutionId = startEvent.ExecutionId

	RunListener(startEvent.Listener, ctx)

	sequenceFlow := ctx.Model.SequenceFlows[startEvent.Outgoing]
	sequenceFlow.Execute(ctx)
}
