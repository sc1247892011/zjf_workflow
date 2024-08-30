package components

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"
)

// MySQLRuntimeService 是 RuntimeService 接口的一个 MySQL 实现
type MySQLRuntimeService struct {
	DB *sql.DB
}

var mysqlRuntimeServiceInstance *MySQLRuntimeService
var mysqlRuntimeServiceOnce sync.Once

// InitializeMySQLRuntimeService 初始化单例实例 得提供一个初始化的启动函数给goframe 还是其他的什么框架 让他能批量启动
func InitializeMySQLRuntimeService(db *sql.DB) {
	mysqlRuntimeServiceOnce.Do(func() {
		mysqlRuntimeServiceInstance = &MySQLRuntimeService{DB: db}
	})
}

// GetMySQLRuntimeService 获取单例实例
func GetMySQLRuntimeService() *MySQLRuntimeService {
	if mysqlRuntimeServiceInstance == nil {
		panic("MySQLRuntimeService is not initialized. Call InitializeMySQLRuntimeService first.")
	}
	return mysqlRuntimeServiceInstance
}

// StartProcessInstance 创建一个新的流程实例
func (r *MySQLRuntimeService) StartProcessInstance(processDefinitionName string, business_key string, createdBy string, formParams string) (int, error) {
	mySQLRepositoryService := GetMySQLRepositoryService()
	//判断是否有现成的 流程定义缓存
	modelMap := GetModelMap()
	var model *Model
	var parseErr error // 外部声明 parseErr 变量
	if (*modelMap)[processDefinitionName] != nil {
		model = (*modelMap)[processDefinitionName]
	} else {
		ppd, err := mySQLRepositoryService.GetLatestProcessDefinitionByName(processDefinitionName)
		if err != nil {
			return 0, fmt.Errorf("failed to retrieve last insert id: %v", err)
		}

		pd := *ppd
		//解析xml
		model, parseErr = ParseXMLByte(pd.XMLContent)
		//更新版本 流程定义不用更新
		model.Version = pd.Version
		if parseErr != nil {
			log.Println("This is a parseErr:", parseErr)
		}

		//更新缓存
		(*modelMap)[model.ProcessDefinitionName] = model
	}
	//在流程实例表里插入记录
	query := `
        INSERT INTO process_instance ( process_definition_name, version, status, created_by, business_key ,start_time)
        VALUES (?, ?,'running',?, ?, ?)
    `
	startTime := time.Now()
	result, err2 := r.DB.Exec(query, model.ProcessDefinitionName, model.Version, createdBy, business_key, startTime)
	if err2 != nil {
		return 0, fmt.Errorf("failed to start process instance: %v", err2)
	}

	id, err2 := result.LastInsertId()
	if err2 != nil {
		return 0, fmt.Errorf("failed to retrieve last insert id: %v", err2)
	}

	//找到开始节点
	startEvent := model.StartEvents
	var startEventElement StartEvent
	//目前只有一个startEvent 以后不知道会不会扩展为多个
	for key := range startEvent {
		startEventElement = startEvent[key]
	}
	var ctx = &WorkflowContext{
		Model:                 model,
		ProcessInstanceId:     int(id),
		ProcessDefinitionName: processDefinitionName,
		Version:               model.Version,
		BusinessKey:           business_key,
		CurrentUserId:         createdBy,
		Data:                  formParams,
		StartTime:             time.Now(),
	}

	startEventElement.Execute(ctx)
	return int(id), nil
}
