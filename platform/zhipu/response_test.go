package zhipu

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestHandle(t *testing.T) {
	var raw = "根据用户的问题，我们可以判断该问题属于'submit_feed_tower_compliment'类型，因为用户提供了类似格式的卸料完成信息。下面是根据用户的描述生成的操作参数，按照您提供的JSON格式：\\n\\n```json\\n{\\n  \\\"body\\\": {\\n    \\\"action\\\": \\\"submit_feed_tower_compliment\\\",\\n    \\\"device_type\\\": \\\"feedTower\\\",\\n    \\\"time_range\\\": \\\"\\\"\\n  }\\n}\\n```\\n\\n注意：时间范围（time_range）在这里是非必填项，因为用户问题中没有提供具体的时间段信息，而且对于这种类型的提交操作通常不需要时间段数据。"
	raw = strings.ReplaceAll(raw, "\\n", "\n")
	raw = strings.ReplaceAll(raw, "\\\"", "\"")
	var jsonStr string
	if strings.Contains(raw, "```json") {
		_, jsonStr, _ = strings.Cut(raw, "```json")
		jsonStr, _, _ = strings.Cut(jsonStr, "```")
	} else {
		jsonStr = raw
	}
	jsonStr = findLineBreakAfterComments(jsonStr)
	jsonStr = strings.ReplaceAll(jsonStr, "\n", "")
	jsonStr = strings.ReplaceAll(jsonStr, "\r", "\\n")
	logrus.Info(jsonStr)
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	logrus.Info(result, err)
}
