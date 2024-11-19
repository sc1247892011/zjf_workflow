package components

import "log"

type ExclusiveGateway struct {
	ExecutionId string   `xml:"executionId,attr"` // 绑定 id 属性
	Outgoing    []string `xml:"Outgoing"`         // 绑定 <Outgoing> 子元素
	Incoming    []string `xml:"Incoming"`         // 绑定 <Incoming> 子元素
	X           string   `XML:"x,attr"`
	Y           string   `XML:"y,attr"`
	H           string   `XML:"h,attr"`
	W           string   `XML:"w,attr"`
	Listener    string   `xml:"Listener"` // 任务监听 执行完毕之后的后续逻辑 方法名称可以用逗号隔开 传递多段逻辑
}

// 方法接收器是 *StartEvent，允许修改 StartEvent 的字段
// 逻辑根并行网关是一样的 除了不需要等待所有的Incoming 到齐，触发一次就全部执行一次outgoing，因为互斥网关后所有的序列流是有条件表达式的，可以自行控制是否继续走
func (exclusiveGateway ExclusiveGateway) Execute(ctx *WorkflowContext) {
	//序列流进入该方法 记录入库
	nodeService := GetServiceFactory().GetNodeService()
	tx := ctx.Tx
	if tx == nil {
		log.Println("Failed to get transaction from ctx ")
		return
	}
	//StartNodeInstance(processInstanceId int, nodeName string, executionId string, previousExecutionId string, assignee string) (int, error)
	//进入互斥网关的序列流只有一条 直接根据表达式条件判断 走下一步 流程不会停止
	nodeId, initerr := nodeService.InitNodeInstance(tx, ctx.ProcessInstanceId, ctx.ProcessDefinitionName, EXCLUSIVE_GATEWAY, exclusiveGateway.ExecutionId, ctx.CurrentExecutionId, SYSTEM_USER_NOBODY)
	if initerr != nil {
		log.Println("Failed to insert exclusiveGateway to database: ", initerr)
		tx.Rollback()
		return
	}

	//网关数据 只要插入一条 就往历史表里同步一条
	historyService := GetServiceFactory().GetHistoryService()
	_, copyerr := historyService.CopyNodeInstance(tx, nodeId, ctx.ProcessInstanceId, ctx.ProcessDefinitionName, EXCLUSIVE_GATEWAY, exclusiveGateway.ExecutionId, ctx.CurrentExecutionId, SYSTEM_USER_NOBODY)
	if copyerr != nil {
		log.Println("Failed to copy history: ", copyerr)
		tx.Rollback()
		return
	}

	//执行监听
	RunListener(exclusiveGateway.Listener, ctx)

	ctx.CurrentExecutionId = exclusiveGateway.ExecutionId
	for _, value := range exclusiveGateway.Outgoing {
		model := *(ctx.Model)
		model.SequenceFlows[value].Execute(ctx)
	}
}
