package chatgpt

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

var jsonText = `{
	"aspect": " 编程能力",
    "flags": [],
    "level": "可培养新人",
    "question": "请编写一个Java程序，输出Hello World！",
  }
}`

type Config struct {
	A string
	C string
}

var c Config

func TestJson(t *testing.T) {
	var matched = jsonLastCommaMatcher.FindString(jsonText)

	fmt.Println(matched)

	jsonText = strings.Replace(jsonText, matched, "}}", 1)

	var reader = strings.NewReader(jsonText)

	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&c)
	if err != nil {
		fmt.Println(c)
		fmt.Println(reader.Len())
		panic(err)
	}

	fmt.Println(c)

}
