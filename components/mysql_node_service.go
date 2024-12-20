package components

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// MySQLNodeService 是 NodeService 接口的一个 MySQL 实现
type MySQLNodeService struct {
	DB *sql.DB
}

var mysqlNodeServiceInstance *MySQLNodeService
var mysqlNodeServiceOnce sync.Once

// InitializeMySQLNodeService 初始化单例实例
func InitializeMySQLNodeService(db *sql.DB) {
	mysqlNodeServiceOnce.Do(func() {
		mysqlNodeServiceInstance = &MySQLNodeService{DB: db}
	})
}

// GetMySQLNodeService 获取单例实例
func GetMySQLNodeService() *MySQLNodeService {
	if mysqlNodeServiceInstance == nil {
		panic("MySQLNodeService is not initialized. Call InitializeMySQLNodeService first.")
	}
	return mysqlNodeServiceInstance
}

func (service *MySQLNodeService) GetTransaction() (*sql.Tx, error) {
	return service.DB.Begin()
}

// StartNodeInstance 创建一个新的节点实例
func (service *MySQLNodeService) InitNodeInstance(tx *sql.Tx, processInstanceId int, processDefinitionName string, nodeName string, executionId string, previousExecutionId string, assignee string) (int, error) {
	query := `
        INSERT INTO node_instance (process_instance_id, process_definition_name, node_name, execution_id, previous_execution_id, assignee, start_time)
        VALUES (?, ?, ?, ?, ?, ?, ?)`

	startTime := time.Now()
	result, err := tx.Exec(query, processInstanceId, processDefinitionName, nodeName, executionId, previousExecutionId, assignee, startTime)
	if err != nil {
		return 0, fmt.Errorf("failed to start node instance: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert id: %v", err)
	}

	return int(id), nil
}

// GetAttributeByExpression 根据表达式获取属性值
func (service *MySQLNodeService) GetAttributeByExpression(tx *sql.Tx, expression string, processInstanceId int) (map[string]interface{}, error) {
	// 提取表达式中的属性
	attributes := ExtractAttributes(expression)

	// 用于存储最终的属性和值
	result := make(map[string]interface{})

	for _, attr := range attributes {
		// 假设属性格式为 "A.a"
		parts := strings.Split(attr, ".")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid attribute format: %s", attr)
		}

		executionId := parts[0]
		fieldName := parts[1]

		// 因为打回的关系 还有流程配置的关系 历史表里保留全量数据 可能不止一条，节点表因为流程配置可能也有多条
		// 打回的时候 直接顺着打回目标节点的outgoing全部删除 可以保证至少节点表里最新的数据 就是可用的数据，因为打回的历史数据全部给删除了，留下来的最新的一定是生效的
		// 因为是用来找自己的轮次的 所以根据start_time还是根据 end_time排序都一样
		// 必须用事务 否则查询不到当前批次数据
		query := `SELECT output_data FROM node_instance WHERE execution_id = ? and process_instance_id = ?  ORDER BY start_time DESC LIMIT 1`
		row := tx.QueryRow(query, executionId, processInstanceId)

		var outputData []byte
		if err := row.Scan(&outputData); err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("no data found for execution_id: %s", executionId)
			}
			return nil, fmt.Errorf("failed to query database: %v", err)
		}

		// 解析 JSON 数据
		var data map[string]interface{}
		if err := json.Unmarshal(outputData, &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
		}

		// 获取具体属性的值
		value, exists := data[fieldName]
		if !exists {
			return nil, fmt.Errorf("attribute %s not found in output data", fieldName)
		}

		// 将属性和值存入结果
		result[attr] = value
	}

	return result, nil
}

// 根据流程实例id和 网关的执行结构id 查询当前并行网关是否满足执行下一步的条件
// for update 利用间隙锁 防止并发场景下的幻读
func (service *MySQLNodeService) CountParallelGatewayIncoming(tx *sql.Tx, processInstanceId int, executionId string) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM node_instance
		WHERE process_instance_id = ? AND execution_id = ? for update
	`
	err := tx.QueryRow(query, processInstanceId, executionId).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count parallel gateway incoming: %v", err)
	}
	return count, nil
}

// GetNodeInstanceById 根据Id获取节点实例
func (service *MySQLNodeService) GetNodeInstanceById(id int) (*NodeInstance, error) {
	query := `SELECT id, process_instance_id,process_definition_name, node_name, execution_id, output_data, previous_execution_id, start_time, end_time FROM node_instance WHERE id = ?`
	instance := &NodeInstance{}
	err := service.DB.QueryRow(query, id).Scan(&instance.Id, &instance.ProcessInstanceId, &instance.ProcessDefinitionName, &instance.NodeName, &instance.ExecutionId, &instance.OutputData, &instance.PreviousExecutionId, &instance.StartTime, &instance.EndTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get node instance by Id: %v", err)
	}
	return instance, nil
}

// GetNodeInstancesByProcessInstanceId 根据流程实例Id获取节点实例列表
func (service *MySQLNodeService) GetNodeInstancesByProcessInstanceId(processInstanceId int) ([]*NodeInstance, error) {
	query := `SELECT id, process_instance_id,process_definition_name, node_name, execution_id, output_data, previous_execution_id, start_time, end_time FROM node_instance WHERE process_instance_id = ?`
	rows, err := service.DB.Query(query, processInstanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to get node instances by process instance Id: %v", err)
	}
	defer rows.Close()

	var instances []*NodeInstance
	for rows.Next() {
		instance := &NodeInstance{}
		err := rows.Scan(&instance.Id, &instance.ProcessInstanceId, &instance.ProcessDefinitionName, &instance.NodeName, &instance.ExecutionId, &instance.OutputData, &instance.PreviousExecutionId, &instance.StartTime, &instance.EndTime)
		if err != nil {
			return nil, fmt.Errorf("failed to scan node instance: %v", err)
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

// UpdateNodeInstanceOutput 更新节点实例的输出数据
func (service *MySQLNodeService) UpdateNodeInstanceOutput(tx *sql.Tx, id int, outputData string) error {
	query := `
        UPDATE node_instance
        SET output_data = ?, end_time = NOW()
        WHERE id = ?
    `
	_, err := tx.Exec(query, outputData, id)
	if err != nil {
		return fmt.Errorf("failed to update node instance output: %v", err)
	}
	return nil
}

func (service *MySQLNodeService) GetAssigneeUndoneTask(assignee string) ([]map[string]interface{}, error) {
	// 创建一个空的map数组用于存储结果
	var results []map[string]interface{}

	// 构建查询语句
	query := `
        SELECT id, process_instance_id,process_definition_name, node_name, execution_id, output_data, previous_execution_id, assignee, start_time, end_time
        FROM node_instance
        WHERE assignee = ? AND output_data IS NULL
    `

	// 执行查询
	rows, err := service.DB.Query(query, assignee)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 遍历查询结果
	for rows.Next() {
		var (
			id                    int
			processInstanceID     int
			processDefinitionName string
			nodeName              string
			executionID           string
			outputData            sql.NullString
			previousExecutionID   sql.NullString
			assignee              sql.NullString
			startTime             sql.NullString
			endTime               sql.NullString
		)

		// 扫描每一行数据
		err := rows.Scan(&id, &processInstanceID, &processDefinitionName, &nodeName, &executionID, &outputData, &previousExecutionID, &assignee, &startTime, &endTime)
		if err != nil {
			return nil, err
		}

		// 将数据放入map
		result := map[string]interface{}{
			"id":                      id,
			"process_instance_id":     processInstanceID,
			"process_definition_name": processDefinitionName,
			"node_name":               nodeName,
			"execution_id":            executionID,
			"output_data":             outputData,
			"previous_execution_id":   previousExecutionID,
			"assignee":                assignee,
			"start_time":              startTime,
			"end_time":                endTime,
		}

		// 将map放入结果数组
		results = append(results, result)
	}

	// 检查是否有错误
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (service *MySQLNodeService) GetTaskDetailByTaskId(taskId int) (map[string]interface{}, error) {
	// 构建查询语句
	query := `
        SELECT id, process_instance_id,process_definition_name, node_name, execution_id, output_data, previous_execution_id, assignee, start_time, end_time
        FROM node_instance
        WHERE id = ?
    `

	// 执行查询
	row := service.DB.QueryRow(query, taskId)

	// 定义用于接收查询结果的变量
	var (
		id                    int
		processInstanceID     int
		processDefinitionName string
		nodeName              string
		executionID           string
		outputData            sql.NullString
		previousExecutionID   sql.NullString
		assignee              string
		startTime             sql.NullTime
		endTime               sql.NullTime
	)

	// 扫描查询结果到变量中
	err := row.Scan(&id, &processInstanceID, &nodeName, &executionID, &outputData, &previousExecutionID, &assignee, &startTime, &endTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task with id %d not found", taskId)
		}
		return nil, fmt.Errorf("failed to get task detail: %v", err)
	}

	// 将查询结果赋值到 map 中
	result := map[string]interface{}{
		"id":                      id,
		"process_instance_id":     processInstanceID,
		"process_definition_name": processDefinitionName,
		"node_name":               nodeName,
		"execution_id":            executionID,
		"output_data":             nilIfEmpty(outputData),
		"previous_execution_id":   nilIfEmpty(previousExecutionID),
		"assignee":                assignee,
		"start_time":              nilIfEmptyTime(startTime),
		"end_time":                nilIfEmptyTime(endTime),
	}

	return result, nil
}

// // DeleteNodeInstance 根据Id删除节点实例 这个方法暂时备用
// func (r *MySQLNodeService) DeleteNodeInstance(tx *sql.Tx, id int) error {
// 	query := ` DELETE FROM node_instance WHERE id = ?`
// 	_, err := r.DB.Exec(query, id)
// 	if err != nil {
// 		return fmt.Errorf("failed to delete node instance: %v", err)
// 	}
// 	return nil
// }

func (service *MySQLNodeService) GetTaskForm(processDefinitionName string, executionId string) (string, error) {
	modelMap := GetModelMap()
	var model *Model
	var parseErr error // 外部声明 parseErr 变量
	if (*modelMap)[processDefinitionName] != nil {
		model = (*modelMap)[processDefinitionName]
	} else {
		mySQLRepositoryService := GetMySQLRepositoryService()
		ppd, err := mySQLRepositoryService.GetLatestProcessDefinitionByName(processDefinitionName)
		if err != nil {
			return "", fmt.Errorf("failed to retrieve last insert id: %v", err)
		}

		if ppd == nil {
			return "", fmt.Errorf("no process definition found with name: %s", processDefinitionName)
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

	formdata := model.Tasks[executionId].FormData
	return formdata, nil
}

func (service *MySQLNodeService) ClearProcessData(tx *sql.Tx, processInstanceId int) error {
	query := ` DELETE FROM node_instance WHERE process_instance_id = ?`
	_, err := tx.Exec(query, processInstanceId)
	if err != nil {
		return fmt.Errorf("failed to delete node instance: %v", err)
	}
	return nil
}
