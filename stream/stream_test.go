package stream

import (
	"encoding/json"
	"testing"
)

// func TestSteamParse(t *testing.T) {
// 	var tested = "{\"logId\":\"89b4b1c50e408f916d3af6368239edd8\",\"errorCode\":0,\"errorMsg\":\"success\",\"result\":{\"id\":\"as-xnhq4ztejv\",\"object\":\"chat.completion\",\"created\":1722171804,\"result\":\"```json\\n{\\\"body\\\":{\\\"description\\\":\\\"这是一款专为外国人设计的旅游辅助手机应用及其后台管理软件。该软件旨在帮助外国游客更好地了解中国，规划他们的旅游行程，并能在旅行过程中获得及时的帮助和信息。后台管理系统则帮助管理人员监控应用的使用情况，管理旅游信息和用户数据。\\\",\\\"name\\\":\\\"旅行小助手及后台管理\\\",\\\"subsystems\\\":[{\\\"description\\\":\\\"该子系统服务于外国游客，提供旅游规划、景点推荐、实时翻译等功能，帮助游客更好地游览中国。\\\",\\\"name\\\":\\\"游客服务系统\\\",\\\"pages\\\":[{\\\"description\\\":\\\"提供用户注册和登录功能。\\\",\\\"name\\\":\\\"登录页\\\",\\\"type\\\":\\\"login-panel\\\"},{\\\"description\\\":\\\"展示用户的个性化旅游推荐、景点信息、行程规划等。\\\",\\\"name\\\":\\\"工作台\\\",\\\"type\\\":\\\"work-space\\\"},{\\\"description\\\":\\\"展示各个景点的详细信息，包括位置、历史背景、游玩建议等。\\\",\\\"name\\\":\\\"景点信息\\\",\\\"type\\\":\\\"single-table\\\"},{\\\"description\\\":\\\"用户可以通过此页面输入需要翻译的文本，并获得即时翻译结果。\\\",\\\"name\\\":\\\"实时翻译\\\",\\\"type\\\":\\\"interact-panel\\\"},{\\\"description\\\":\\\"用户可以在此页面规划和调整自己的旅游行程。\\\",\\\"name\\\":\\\"行程规划\\\",\\\"type\\\":\\\"form-panel\\\"}],\\\"user\\\":\\\"外国游客\\\"},{\\\"description\\\":\\\"该子系统为管理人员提供数据监控、用户管理、内容管理等功能，确保应用的正常运行和内容的准确性。\\\",\\\"name\\\":\\\"后台管理系统\\\",\\\"pages\\\":[{\\\"description\\\":\\\"提供管理员登录功能。\\\",\\\"name\\\":\\\"登录页\\\",\\\"type\\\":\\\"login-panel\\\"},{\\\"description\\\":\\\"展示应用的使用情况统计、用户活跃度等信息。\\\",\\\"name\\\":\\\"数据总览\\\",\\\"type\\\":\\\"dash-board\\\"},{\\\"description\\\":\\\"管理人员可以在此页面查看、编辑和管理用户信息。\\\",\\\"name\\\":\\\"用户管理\\\",\\\"type\\\":\\\"single-table\\\"},{\\\"description\\\":\\\"用于管理和更新应用内的景点信息、推荐内容等。\\\",\\\"name\\\":\\\"内容管理\\\",\\\"type\\\":\\\"work-space\\\"},{\\\"description\\\":\\\"展示应用的异常日志、用户反馈等信息，帮助管理人员及时发现问题并进行处理。\\\",\\\"name\\\":\\\"问题与反馈\\\",\\\"type\\\":\\\"interact-panel\\\"}],\\\"user\\\":\\\"管理人员\\\"}]}}\\n```\",\"usage\":{\"prompt_tokens\":407,\"completion_tokens\":405,\"total_tokens\":812},\"is_truncated\":false,\"finish_reason\":\"normal\",\"need_clear_history\":false}}"
// 	var parser = NewStreamParser()
// 	parser.Init()
// 	for i := 0; i < len(tested); i++ {
// 		parser.Append(tested[i])
// 	}
// 	parser.Finish()
// }

func TestStreamParse(t *testing.T) {
	var parser JsonStreamer
	parser.Append("   {")
	parser.Append("   \"name\": \"test\"")
	parser.Append("   }")
	parser.Append("   ")
	parser.Append("   ")
	bts, err := json.Marshal(&parser)
	if err != nil {
		t.Error(err)
	}
	if string(bts) != `{"name":"test"}` {
		t.Error("unexpected result", string(bts))
	}
}

func TestStreamParseUnfinishedObject(t *testing.T) {
	var parser JsonStreamer
	parser.Append("   {")
	parser.Append("   \"name\": \"test\",")
	parser.Append("   \"name1\": \"")
	parser.Append("   ")
	parser.Append("   ")
	bts, err := json.Marshal(&parser)
	if err != nil {
		t.Error(err)
	}
	if string(bts) != `{"name":"test","name1":""}` {
		t.Error("unexpected result", string(bts))
	}
}

func TestStreamParseUnfinishedObjectWithNumber(t *testing.T) {
	var parser JsonStreamer
	parser.Append("   {")
	parser.Append("   \"name\": \"test\",")
	parser.Append("   \"name1\": 1,")
	parser.Append("   \"name2\": \"")
	parser.Append("   ")
	parser.Append("   ")
	bts, err := json.Marshal(&parser)
	if err != nil {
		t.Error(err)
	}
	if string(bts) != `{"name":"test","name1":1,"name2":""}` {
		t.Error("unexpected result", string(bts))
	}
}

func TestStreamParseObjectWithNumber(t *testing.T) {
	var parser JsonStreamer
	parser.Append("   {")
	parser.Append("   \"name\": \"test\",")
	parser.Append("   \"name1\": 1")
	parser.Append("  } ")
	parser.Append("   ")
	bts, err := json.Marshal(&parser)
	if err != nil {
		t.Error(err)
	}
	if string(bts) != `{"name":"test","name1":1}` {
		t.Error("unexpected result", string(bts))
	}
}

func TestStreamParseObjectWithFloatNumber(t *testing.T) {
	var parser JsonStreamer
	parser.Append("   {")
	parser.Append("   \"name\": \"test\",")
	parser.Append("   \"name1\": 1.2")
	parser.Append("  } ")
	parser.Append("   ")
	bts, err := json.Marshal(&parser)
	if err != nil {
		t.Error(err)
	}
	if string(bts) != `{"name":"test","name1":1.2}` {
		t.Error("unexpected result", string(bts))
	}
}

func TestStreamParseObjectWithBoolNumber(t *testing.T) {
	var parser JsonStreamer
	parser.Append("   {")
	parser.Append("   \"name\": \"test\",")
	parser.Append("   \"name1\": false")
	parser.Append("  } ")
	parser.Append("   ")
	bts, err := json.Marshal(&parser)
	if err != nil {
		t.Error(err)
	}
	if string(bts) != `{"name":"test","name1":false}` {
		t.Error("unexpected result", string(bts))
	}
}

func TestStreamParseObjectWithNull(t *testing.T) {
	var parser JsonStreamer
	parser.Append("   {")
	parser.Append("   \"name\": \"test\",")
	parser.Append("   \"name1\": null")
	parser.Append("  } ")
	parser.Append("   ")
	bts, err := json.Marshal(&parser)
	if err != nil {
		t.Error(err)
	}
	if string(bts) != `{"name":"test","name1":null}` {
		t.Error("unexpected result", string(bts))
	}
}

func TestStreamParseObjectWithUndefined(t *testing.T) {
	var parser JsonStreamer
	parser.Append("   {")
	parser.Append("   \"name\": \"test\",")
	parser.Append("   \"name1\": undefined")
	parser.Append("  } ")
	parser.Append("   ")
	bts, err := json.Marshal(&parser)
	if err != nil {
		t.Error(err)
	}
	if string(bts) != `{"name":"test"}` {
		t.Error("unexpected result", string(bts))
	}
}

func TestStreamParseArrayInObject(t *testing.T) {
	var parser JsonStreamer
	parser.Append("   {")
	parser.Append("   \"name\": \"test\",")
	parser.Append("   \"name1\": [1,2,3]")
	parser.Append("  } ")
	parser.Append("   ")
	bts, err := json.Marshal(&parser)
	if err != nil {
		t.Error(err)
	}
	if string(bts) != `{"name":"test","name1":[1,2,3]}` {
		t.Error("unexpected result", string(bts))
	}
}

func TestStreamParseArrayOfDifferentTypeInObject(t *testing.T) {
	var parser JsonStreamer
	parser.Append("   {")
	parser.Append("   \"name\": \"test\",")
	parser.Append("   \"name1\": [1,2,3, \"test\"]")
	parser.Append("  } ")
	parser.Append("   ")
	bts, err := json.Marshal(&parser)
	if err != nil {
		t.Error(err)
	}
	if string(bts) != `{"name":"test","name1":[1,2,3,"test"]}` {
		t.Error("unexpected result", string(bts))
	}
}

// Unsupported: "undefined" in Go
// func TestStreamParseArrayHasUndefinedInObject(t *testing.T) {
// 	var parser JsonStreamer
// 	parser.Append("   {")
// 	parser.Append("   \"name\": \"test\",")
// 	parser.Append("   \"name1\": [1,2, undefined, \"test\"]")
// 	parser.Append("  } ")
// 	parser.Append("   ")
// 	bts, err := json.Marshal(&parser)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	if string(bts) != `{"name":"test","name1":[1,2,3,"test"]}` {
// 		t.Error("unexpected result", string(bts))
// 	}
// }

func TestStreamParseArray(t *testing.T) {
	var parser JsonStreamer
	parser.Append("   [")
	parser.Append("   \"name\",")
	parser.Append("   \"test\"")
	parser.Append("  ] ")
	parser.Append("   ")
	bts, err := json.Marshal(&parser)
	if err != nil {
		t.Error(err)
	}
	if string(bts) != `["name","test"]` {
		t.Error("unexpected result", string(bts))
	}
}

func TestStreamParseArrayWithNumber(t *testing.T) {
	var parser JsonStreamer
	parser.Append("   [")
	parser.Append("   \"name\",")
	parser.Append("   1")
	parser.Append("  ] ")
	parser.Append("   ")
	bts, err := json.Marshal(&parser)
	if err != nil {
		t.Error(err)
	}
	if string(bts) != `["name",1]` {
		t.Error("unexpected result", string(bts))
	}
}

func TestStreamParseObjectInObject(t *testing.T) {
	var parser JsonStreamer
	parser.Append("   {")
	parser.Append("   \"name\": \"test\",")
	parser.Append("   \"name1\": {\"name\": \"test\"}")
	parser.Append("  } ")
	parser.Append("   ")
	bts, err := json.Marshal(&parser)
	if err != nil {
		t.Error(err)
	}
	if string(bts) != `{"name":"test","name1":{"name":"test"}}` {
		t.Error("unexpected result", string(bts))
	}
}

func TestStreamParseObjectInObject2(t *testing.T) {
	var parser JsonStreamer
	parser.Append("   {")
	parser.Append("   \"name1\": {\"name\": \"test\"}")
	parser.Append("  } ")
	parser.Append("   ")
	bts, err := json.Marshal(&parser)
	if err != nil {
		t.Error(err)
	}
	if string(bts) != `{"name1":{"name":"test"}}` {
		t.Error("unexpected result", string(bts))
	}
}

func TestStreamParseObjectInObject3(t *testing.T) {
	var parser JsonStreamer
	parser.Append("   {")
	parser.Append("   \"body\": {\"name\": \"test\",")
	parser.Append("  \"description\": \"test\",")
	parser.Append("  \"subsystems\": [")
	parser.Append("  {\"name\": \"test\"}")
	parser.Append("  ]")
	parser.Append("  } ")
	bts, err := json.Marshal(&parser)
	if err != nil {
		t.Error(err)
	}
	if string(bts) != `{"body":{"description":"test","name":"test","subsystems":[{"name":"test"}]}}` {
		t.Error("unexpected result", string(bts))
	}
}
