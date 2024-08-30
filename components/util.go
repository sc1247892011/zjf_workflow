package components

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"

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

// 解析json为map字符串
func ParseJSON(input string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(input), &result)
	return result, err
}

// govaluate表达式计算
func EvaluateExpression(expression string, parameters map[string]interface{}) (interface{}, error) {
	// 创建表达式解析器
	expr, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %v", err)
	}

	// 评估表达式
	result, err := expr.Evaluate(parameters)
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
