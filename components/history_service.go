package components

type HistoryService interface {
	//迁徙节点数据到历史表
	CopyNodeInstanceById(nodeId int) error
}
