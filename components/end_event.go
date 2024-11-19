package components

import "log"

type EndEvent struct {
	ExecutionId string `xml:"executionId,attr"` // 绑定 id 属性
	Name        string `xml:"name,attr"`
	Incoming    string `xml:"Incoming"` // 绑定 <Incoming> 子元素
	X           string `xml:"x,attr"`
	Y           string `xml:"y,attr"`
	H           string `xml:"h,attr"`
	W           string `xml:"w,attr"`
	Listener    string `xml:"Listener"` // 任务监听 执行完毕之后的后续逻辑 方法名称可以用逗号隔开 传递多段逻辑
}

// 方法接收器是 *StartEvent，允许修改 StartEvent 的字段
func (endEvent EndEvent) Execute(ctx *WorkflowContext) {
	//更新数据库流程实例状态 迁移数据到历史表
	nodeService := GetServiceFactory().GetNodeService()
	tx := ctx.Tx
	if tx == nil {
		log.Println("Failed to get transaction from ctx ")
		return
	}
	nodeId, initerr := nodeService.InitNodeInstance(tx, ctx.ProcessInstanceId, ctx.ProcessDefinitionName, endEvent.Name, endEvent.ExecutionId, ctx.CurrentExecutionId, "")
	if initerr != nil {
		log.Println("Failed to insert endEvent to database: ", initerr)
		tx.Rollback()
		return
	}
	//迁徙数据到历史库
	historyService := GetServiceFactory().GetHistoryService()
	_, he := historyService.CopyNodeInstance(tx, nodeId, ctx.ProcessInstanceId, ctx.ProcessDefinitionName, endEvent.Name, endEvent.ExecutionId, ctx.CurrentExecutionId, "")
	if he != nil {
		log.Println("Failed to insert endEvent to history: ", he)
		tx.Rollback()
		return
	}
	//更新数据库任务状态
	runtimeService := GetServiceFactory().GetRuntimeService()
	completeerr := runtimeService.CompleteProcessInstance(tx, ctx.ProcessInstanceId)
	if completeerr != nil {
		log.Println("Failed to complete: ", completeerr)
		tx.Rollback()
		return
	}
	//删除当前流程实例的数据
	clearerr := nodeService.ClearProcessData(tx, ctx.ProcessInstanceId)
	if clearerr != nil {
		log.Println("Failed to ClearProcessData from database: ", clearerr)
		tx.Rollback()
		return
	}

	RunListener(endEvent.Listener, ctx)
	tx.Commit()
}
