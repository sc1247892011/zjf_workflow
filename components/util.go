package components

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/Knetic/govaluate"
)

// 读取文件内容
func ReadXMLFile(filePath string) ([]byte, error) {
	xmlContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read XML file: %v", err)
	}

	return xmlContent, nil
}

// 解析json字符串为json对象，在go里json对象是map[string]interface{}的形式
func ParseJSON(input string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(input), &result)
	return result, err
}

// 将 map 转换为 JSON 字符串
func ToJsonString(data map[string]interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error converting map to JSON:", err)
		return "", err
	}
	return string(jsonData), nil
}

// govaluate表达式计算
func EvaluateExpression(expression string, parameters map[string]interface{}) (interface{}, error) {
	for key, value := range parameters {
		// 将参数中的值替换到表达式中
		strValue := fmt.Sprintf("'%v'", value)
		expression = strings.ReplaceAll(expression, key, strValue)
	}
	// 创建表达式解析器
	expr, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %v", err)
	}

	// 评估表达式
	result, err := expr.Evaluate(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %v", err)
	}

	return result, nil
}

// 从govaluate表达式字符串中提取属性名
func ExtractAttributes(expression string) []string {
	// 定义正则表达式，用于匹配属性名，例如 A.a 或 data.value
	regex := regexp.MustCompile(`\b[A-Za-z_]\w*\.\w+\b`)
	matches := regex.FindAllString(expression, -1)

	// 去重处理
	uniqueMatches := make(map[string]bool)
	var attributes []string
	for _, match := range matches {
		if _, exists := uniqueMatches[match]; !exists {
			uniqueMatches[match] = true
			attributes = append(attributes, match)
		}
	}

	return attributes
}

// 辅助函数用于处理 sql.NullString
func nilIfEmpty(ns sql.NullString) interface{} {
	if ns.Valid {
		return ns.String
	}
	return nil
}

// 辅助函数用于处理 sql.NullTime
func nilIfEmptyTime(nt sql.NullTime) interface{} {
	if nt.Valid {
		return nt.Time
	}
	return nil
}
