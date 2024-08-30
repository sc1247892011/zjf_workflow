package components

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

// StartNodeInstance 创建一个新的节点实例
func (r *MySQLNodeService) InitNodeInstance(processInstanceId int, nodeName string, executionId string, previousExecutionId string, assignee string) (int, error) {
	query := `
        INSERT INTO node_instance (process_instance_id, node_name, execution_id, previous_execution_id, assignee, start_time)
        VALUES (?, ?, ?, ?, ?, ?)`

	startTime := time.Now()
	result, err := r.DB.Exec(query, processInstanceId, nodeName, executionId, previousExecutionId, assignee, startTime)
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
func (r *MySQLNodeService) GetAttributeByExpression(expression string, processInstanceId int) (map[string]interface{}, error) {
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
		query := `SELECT output_data FROM node_instance WHERE execution_id = ? and process_instance_id = ?  ORDER BY start_time DESC LIMIT 1`
		row := r.DB.QueryRow(query, executionId, processInstanceId)

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
func (r *MySQLNodeService) CountParallelGatewayIncoming(processInstanceId int, executionId string) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM node_instance
		WHERE process_instance_id = ? AND execution_id = ?
	`
	err := r.DB.QueryRow(query, processInstanceId, executionId).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count parallel gateway incoming: %v", err)
	}
	return count, nil
}

// GetNodeInstanceById 根据Id获取节点实例
func (r *MySQLNodeService) GetNodeInstanceById(id int) (*NodeInstance, error) {
	query := `SELECT id, process_instance_id, node_name, execution_id, output_data, previous_execution_id, start_time, end_time FROM node_instance WHERE id = ?`
	instance := &NodeInstance{}
	err := r.DB.QueryRow(query, id).Scan(&instance.Id, &instance.ProcessInstanceId, &instance.NodeName, &instance.ExecutionId, &instance.OutputData, &instance.PreviousExecutionId, &instance.StartTime, &instance.EndTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get node instance by Id: %v", err)
	}
	return instance, nil
}

// GetNodeInstancesByProcessInstanceId 根据流程实例Id获取节点实例列表
func (r *MySQLNodeService) GetNodeInstancesByProcessInstanceId(processInstanceId int) ([]*NodeInstance, error) {
	query := `SELECT id, process_instance_id, node_name, execution_id, output_data, previous_execution_id, start_time, end_time FROM node_instance WHERE process_instance_id = ?`
	rows, err := r.DB.Query(query, processInstanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to get node instances by process instance Id: %v", err)
	}
	defer rows.Close()

	var instances []*NodeInstance
	for rows.Next() {
		instance := &NodeInstance{}
		err := rows.Scan(&instance.Id, &instance.ProcessInstanceId, &instance.NodeName, &instance.ExecutionId, &instance.OutputData, &instance.PreviousExecutionId, &instance.StartTime, &instance.EndTime)
		if err != nil {
			return nil, fmt.Errorf("failed to scan node instance: %v", err)
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

// UpdateNodeInstanceOutput 更新节点实例的输出数据
func (r *MySQLNodeService) UpdateNodeInstanceOutput(id int, outputData string) error {
	query := `
        UPDATE node_instance
        SET output_data = ?
        WHERE id = ?
    `
	_, err := r.DB.Exec(query, outputData, id)
	if err != nil {
		return fmt.Errorf("failed to update node instance output: %v", err)
	}
	return nil
}

// CompleteNodeInstance 完成节点实例
func (r *MySQLNodeService) CompleteNodeInstance(id int) error {
	endTime := time.Now()
	query := `
        UPDATE node_instance
        SET end_time = ?
        WHERE id = ?
    `
	_, err := r.DB.Exec(query, endTime, id)
	if err != nil {
		return fmt.Errorf("failed to complete node instance: %v", err)
	}
	return nil
}

// DeleteNodeInstance 根据Id删除节点实例
func (r *MySQLNodeService) DeleteNodeInstance(id int) error {
	query := ` DELETE FROM node_instance WHERE id = ?`
	_, err := r.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete node instance: %v", err)
	}
	return nil
}
