package zhipu

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestHandle(t *testing.T) {
	var raw = "根据用户的问题，我们需要获取特定饲料种类的重量信息，并确认饲料的卸料情况。问题中提到的饲料种类代码“rzb/566”和对应的重量以及卸料信息暗示我们应该使用`submit_feed_tower_compliment`类型操作来提交这些数据。\n\n由于用户的输入数据格式类似我们设定的判定条件，因此下一步的具体参数应如下所示：\n\n```json\n{\n  \"body\": {\n    \"action\": \"submit_feed_tower_compliment\",\n    \"device_type\": \"feedTower\",\n    \"time_range\": \"1d\"  // 这里假设我们只需要提交当天的数据，时间范围可以根据需要调整\n  }\n}\n```\n\n这个JSON提交信息指示系统执行一个`submit_feed_tower_compliment`操作，涉及到的设备类型是`feedTower`（料塔），并且我们关注的是当天（`1d`）的数据。当然，如果需要具体到某一时间段内的数据，可以根据用户的具体要求调整`time_range`的值。由于用户没有提供具体的时间范围要求，这里假设为`1d`。"
	raw = strings.ReplaceAll(raw, "\\n", "\n")
	raw = strings.ReplaceAll(raw, "\\\"", "\"")
	var jsonStr string
	if strings.Contains(raw, "```json") {
		_, jsonStr, _ = strings.Cut(raw, "```json")
		jsonStr, _, _ = strings.Cut(jsonStr, "```")
	} else {
		jsonStr = raw
	}
	if strings.Contains(jsonStr, "/*") {
		var after string
		jsonStr, after, _ = strings.Cut(jsonStr, "/*")
		_, after, _ = strings.Cut(after, "*/")
		jsonStr += after
	}
	jsonStr = findLineBreakAfterComments(jsonStr)
	jsonStr = strings.ReplaceAll(jsonStr, "\n", "")
	jsonStr = strings.ReplaceAll(jsonStr, "\r", "\n")
	logrus.Info(jsonStr)
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	logrus.Info(result, err)
}
