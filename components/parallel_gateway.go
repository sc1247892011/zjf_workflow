package components

import "log"

type ParallelGateway struct {
	ExecutionId string   `xml:"executionId,attr"` // 绑定 id 属性
	Outgoing    []string `xml:"Outgoing"`         // 绑定 <Outgoing> 子元素
	Incoming    []string `xml:"Incoming"`         // 绑定 <Incoming> 子元素
}

// 方法接收器是 *StartEvent，允许修改 StartEvent 的字段
func (parallelGateway ParallelGateway) Execute(ctx *WorkflowContext) {
	//序列流进入该方法 记录入库
	nodeService := GetServiceFactory().GetNodeService()
	//StartNodeInstance(processInstanceId int, nodeName string, executionId string, previousExecutionId string, assignee string) (int, error)
	nodeId, initerr := nodeService.InitNodeInstance(ctx.ProcessInstanceId, "parallelGateway", parallelGateway.ExecutionId, ctx.CurrentExecutionId, "nobody")
	if initerr != nil {
		log.Println("Failed to insert ParallelGateway to database: ", initerr)
	}

	//网关数据 只要插入一条 就往历史表里同步一条
	historyService := GetServiceFactory().GetHistoryService()
	historyService.CopyNodeInstanceById(nodeId)

	amountIncomingNum := len(parallelGateway.Incoming)
	//去数据库查询 当前已经有几条记录完成
	finishTaskNum, err := nodeService.CountParallelGatewayIncoming(ctx.ProcessInstanceId, parallelGateway.ExecutionId)
	if err != nil {
		log.Println("Failed to count ParallelGateway IncomingNum to database: ", err)
	}

	if finishTaskNum == amountIncomingNum {
		parallelGateway.Complete(ctx)
	} else if finishTaskNum > amountIncomingNum {
		//大于说明工作流里有循环，存在历史数据，不能整除 说明第n轮并没有执行完毕
		if amountIncomingNum != 0 && (finishTaskNum%amountIncomingNum == 0) {
			parallelGateway.Complete(ctx)
		}
	} else if finishTaskNum < amountIncomingNum {
		//网关的序列流任务没有全部接收
		return
	}
}

func (parallelGateway ParallelGateway) Complete(ctx *WorkflowContext) {
	//网关得入库 所以得更新ctx的id进历史数据的结构
	ctx.CurrentExecutionId = parallelGateway.ExecutionId

	//不更新数据库 因为没有输出
	//遍历执行全部的outgoing序列流逻辑
	for _, value := range parallelGateway.Outgoing {
		model := *(ctx.Model)
		model.SequenceFlows[value].Execute(ctx)
	}

}
