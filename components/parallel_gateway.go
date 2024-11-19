package components

import "log"

type ParallelGateway struct {
	ExecutionId string   `xml:"executionId,attr"` // 绑定 id 属性
	Outgoing    []string `xml:"Outgoing"`         // 绑定 <Outgoing> 子元素
	Incoming    []string `xml:"Incoming"`         // 绑定 <Incoming> 子元素
	X           string   `xml:"x,attr"`
	Y           string   `xml:"y,attr"`
	H           string   `xml:"h,attr"`
	W           string   `xml:"w,attr"`
	Listener    string   `xml:"Listener"` // 任务监听 执行完毕之后的后续逻辑 方法名称可以用逗号隔开 传递多段逻辑
}

// 方法接收器是 *StartEvent，允许修改 StartEvent 的字段
func (parallelGateway ParallelGateway) Execute(ctx *WorkflowContext) {
	//序列流进入该方法 记录入库
	nodeService := GetServiceFactory().GetNodeService()
	tx := ctx.Tx
	if tx == nil {
		log.Println("Failed to get transaction from ctx ")
		return
	}
	//StartNodeInstance(processInstanceId int, nodeName string, executionId string, previousExecutionId string, assignee string) (int, error)
	nodeId, initerr := nodeService.InitNodeInstance(tx, ctx.ProcessInstanceId, ctx.ProcessDefinitionName, PARALLEL_GATEWAY, parallelGateway.ExecutionId, ctx.CurrentExecutionId, SYSTEM_USER_NOBODY)
	if initerr != nil {
		log.Println("Failed to insert ParallelGateway to database: ", initerr)
		tx.Rollback()
		return
	}

	//网关数据 只要插入一条 就往历史表里同步一条
	historyService := GetServiceFactory().GetHistoryService()
	_, copyerr := historyService.CopyNodeInstance(tx, nodeId, ctx.ProcessInstanceId, ctx.ProcessDefinitionName, PARALLEL_GATEWAY, parallelGateway.ExecutionId, ctx.CurrentExecutionId, SYSTEM_USER_NOBODY)
	if copyerr != nil {
		log.Println("Failed to CopyNodeInstanceById to database: ", copyerr)
		tx.Rollback()
		return
	}

	amountIncomingNum := len(parallelGateway.Incoming)
	//去数据库查询 当前已经有几条记录完成
	finishTaskNum, err := nodeService.CountParallelGatewayIncoming(tx, ctx.ProcessInstanceId, parallelGateway.ExecutionId)
	if err != nil {
		log.Println("Failed to count ParallelGateway IncomingNum to database: ", err)
		tx.Rollback()
		return
	}

	//执行监听
	RunListener(parallelGateway.Listener, ctx)

	//如果只差当前一个 就全部完成,那么就执行完成的逻辑
	//在事务里 即使是没有提交的数据 也可以查询到 所以不需要+1
	if finishTaskNum == amountIncomingNum {
		parallelGateway.Complete(ctx)
	} else if finishTaskNum > amountIncomingNum {
		//大于说明工作流里有循环，存在历史数据，不能整除 说明第n轮并没有执行完毕
		//该判断主要是为了考虑打回 如果是打回到并行网关前的子分支 打回是需要做取消动作的 最后肯定是能维持数量相等
		if amountIncomingNum != 0 && ((finishTaskNum)%amountIncomingNum == 0) {
			parallelGateway.Complete(ctx)
		}
	} else if finishTaskNum < amountIncomingNum {
		//网关的序列流任务没有全部接收
		//当前序列流子任务已经结束 可以提交事务
		tx.Commit()
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

	//任务节点里做了判断的 如果不是并行网关才会自行提交事务
	//到达任务有3种方式 网关 任务 开始节点 ，但是并行网关 因为需要初始化多个任务节点 所以必须等其他的都执行完毕才能够提交
	ctx.Tx.Commit()
}
