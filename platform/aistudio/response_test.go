package aistudio

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestHandle(t *testing.T) {
	var raw = "{\n  \"body\": {\n    \"code\": \"\n      // 获取页面attributes\n      var email = getAttribute('Email');\n      var username = getAttribute('用户名');\n      var phoneNumber = getAttribute('手机号码');\n      \n      // 登录功能实现\n      function login() {\n        // 使用axios发送登录请求\n        axios.post('/api/login', {\n          email: email,\n          username: username,\n          phoneNumber: phoneNumber\n        })\n        .then(function (response) {\n          // 登录成功，保存token等信息\n          localStorage.setItem('token', response.data.token);\n          \n          // 跳转至主页面\n          router.push('/main');\n        })\n        .catch(function (error) {\n          // 登录失败，处理错误信息\n          console.error(error);\n        });\n      }\n      \n      // 执行登录函数\n      login();\n    \"\n  }\n}\n"
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
