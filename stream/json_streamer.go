package stream

import (
	"encoding/json"
	"strings"

	"github.com/sirupsen/logrus"
)

type JsonStreamer struct {
	children JsonToken
}

type JsonToken interface {
	Append(c rune) (bool, bool, error) //finished, needReappend last char, error
	MarshalJSON() ([]byte, error)
}

const (
	_JSON_OBJECT_STEP_NONE = iota
	_JSON_OBJECT_STEP_KEY
	_JSON_OBJECT_STEP_COLON
	_JSON_OBJECT_STEP_VALUE
	_JSON_OBJECT_STEP_VALUE_INNER
	_JSON_OBJECT_STEP_COMMA //,
)

type JsonObject struct {
	step     byte
	values   map[string]JsonToken
	curValue JsonToken
	curKey   string
	builder  strings.Builder
}

func (j *JsonObject) MarshalJSON() ([]byte, error) {
	var filtered = make(map[string]JsonToken)
	for k, v := range j.values {
		if _, ok := v.(*JsonUndefined); !ok {
			filtered[k] = v
		}
	}
	if j.curKey != "" {
		filtered[j.curKey] = j.curValue
	}
	return json.Marshal(filtered)
}

func (j *JsonStreamer) Json() json.RawMessage {
	if j.children == nil {
		return nil
	}
	bts, err := j.children.MarshalJSON()
	if err != nil {
		return nil
	}
	return json.RawMessage(bts)
}

func (j *JsonObject) Append(c rune) (bool, bool, error) {
	if j.step == _JSON_OBJECT_STEP_NONE {
		switch c {
		case '"':
			j.step = _JSON_OBJECT_STEP_KEY
		case '}':
			j.step = _JSON_OBJECT_STEP_NONE
			return true, false, nil
		}
	} else if j.step == _JSON_OBJECT_STEP_KEY {
		switch c {
		case '"':
			j.step = _JSON_OBJECT_STEP_COLON
			j.curKey = j.builder.String()
		default:
			j.builder.WriteRune(c)
		}
	} else if j.step == _JSON_OBJECT_STEP_COLON {
		switch c {
		case ':':
			j.step = _JSON_OBJECT_STEP_VALUE
		}
	} else if j.step == _JSON_OBJECT_STEP_VALUE {
		j.curValue = detectBasic(c)
		if j.curValue == nil {
			return false, false, nil
		}
		j.step = _JSON_OBJECT_STEP_VALUE_INNER
	} else if j.step == _JSON_OBJECT_STEP_VALUE_INNER {
		finished, needReappend, err := j.curValue.Append(c)
		if err != nil {
			return false, false, err
		} else if finished {
			logrus.WithField("key", j.curKey).WithField("value", j.curValue).Debug("append object value")
			j.values[j.curKey] = j.curValue
			j.step = _JSON_OBJECT_STEP_COMMA
			if needReappend {
				j.Append(c)
			}
			return false, needReappend, nil
		}
	} else if j.step == _JSON_OBJECT_STEP_COMMA {
		switch c {
		case ',':
			j.step = _JSON_OBJECT_STEP_NONE
			j.curKey = ""
			j.curValue = nil
			j.builder.Reset()
		case '}':
			j.curKey = ""
			j.curValue = nil
			j.step = _JSON_OBJECT_STEP_NONE
			return true, false, nil
		}
	}

	return false, false, nil
}

type JsonString struct {
	builder strings.Builder
	value   string
}

func (j *JsonString) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.value)
}

func (j *JsonString) Append(c rune) (bool, bool, error) {
	if c == '"' {
		j.value = j.builder.String()
		return true, false, nil
	}
	j.builder.WriteRune(c)
	return false, false, nil
}

type JsonNumber struct {
	jsonSingleItem
}

func (*JsonNumber) IsLegal(c rune) bool {
	return c == '0' || c == '1' || c == '2' || c == '3' || c == '4' || c == '5' || c == '6' || c == '7' || c == '8' || c == '9' || c == '-' || c == '+' || c == '.' || c == 'e' || c == 'E'
}

func (j *JsonNumber) Append(c rune) (bool, bool, error) {
	return _append(j, c)
}

type JsonBool struct {
	jsonSingleItem
}

func (j *JsonBool) IsLegal(c rune) bool {
	var parsedLen = j.builder.Len()
	if parsedLen == 0 {
		return c == 't' || c == 'f'
	}
	var parsed = j.builder.String()
	if strings.Contains("true", parsed) || strings.Contains("false", parsed) {
		return true
	}
	return false
}

func (j *JsonBool) Append(c rune) (bool, bool, error) {
	return _append(j, c)
}

type JsonNull struct {
	jsonSingleItem
}

func (j *JsonNull) IsLegal(c rune) bool {
	var parsedLen = j.builder.Len()
	if parsedLen == 0 {
		return c == 'n'
	}
	var parsed = j.builder.String()
	return strings.Contains("null", parsed)
}

func (j *JsonNull) Append(c rune) (bool, bool, error) {
	return _append(j, c)
}

type JsonUndefined struct {
	jsonSingleItem
}

func (j *JsonUndefined) IsLegal(c rune) bool {
	var parsedLen = j.builder.Len()
	if parsedLen == 0 {
		return c == 'u'
	}
	var parsed = j.builder.String()
	return strings.Contains("undefined", parsed)
}

func (j *JsonUndefined) Append(c rune) (bool, bool, error) {
	return _append(j, c)
}

type jsonSingleItem struct {
	builder strings.Builder
	value   string
}

func (j *jsonSingleItem) MarshalJSON() ([]byte, error) {
	return json.Marshal(json.RawMessage(j.value))
}

type jsonLegalTester interface {
	IsLegal(c rune) bool
	finish()
	append(c rune)
}

func (j *jsonSingleItem) finish() {
	j.value = j.builder.String()
}

func (j *jsonSingleItem) append(c rune) {
	j.builder.WriteRune(c)
}

func _append(j interface{}, c rune) (bool, bool, error) {
	if tester, ok := j.(jsonLegalTester); ok {
		if !tester.IsLegal(c) {
			tester.finish()
			return true, true, nil
		}
		tester.append(c)
	}
	return false, false, nil
}

type JsonArray struct {
	step    byte
	values  []JsonToken
	curItem JsonToken
}

func (j *JsonArray) MarshalJSON() ([]byte, error) {

	if j.curItem == nil {
		return json.Marshal(j.values)
	} else {
		var tempValues = make([]JsonToken, len(j.values)+1)
		copy(tempValues, j.values)
		tempValues[len(j.values)] = j.curItem
		return json.Marshal(tempValues)
	}
}

func (j *JsonArray) Append(c rune) (bool, bool, error) {
	if j.step == _JSON_OBJECT_STEP_VALUE_INNER {
		finished, needReappend, err := j.curItem.Append(c)
		if err != nil {
			return false, false, err
		}
		if finished {
			j.values = append(j.values, j.curItem)
			j.curItem = nil
			if needReappend {
				switch c {
				case ',', ' ', '\t', '\n', '\r':
					j.step = _JSON_OBJECT_STEP_NONE
				case ']':
					return true, false, nil
				}
			} else {
				j.step = _JSON_OBJECT_STEP_COMMA
			}
		}
		return false, false, nil
	} else if j.step == _JSON_OBJECT_STEP_NONE {

		var children = detectBasic(c)
		if children != nil {
			j.curItem = children
			j.step = _JSON_OBJECT_STEP_VALUE_INNER
			return false, false, nil
		}
	} else if j.step == _JSON_OBJECT_STEP_COMMA {
		switch c {
		case ',':
			j.step = _JSON_OBJECT_STEP_NONE
			j.curItem = nil
		case ']':
			return true, false, nil
		}
	}
	return false, false, nil
}

// return true if finished, false if need reappend last char
func (j *JsonStreamer) Append(str string) (bool, error) {

	for _, c := range str {
		finished, err := j.AppendRune(c)
		if err != nil {
			return false, err
		}
		if finished {
			return true, nil
		}
	}
	return false, nil
}

func (j *JsonStreamer) AppendRune(c rune) (bool, error) {
	if j.children != nil {
		finished, _, err := j.children.Append(c)
		if err != nil {
			return false, err
		}
		if finished {
			return true, nil
		}
		return false, nil
	}
	j.children = detectBasic(c)
	return false, nil
}

func (j *JsonStreamer) MarshalJSON() ([]byte, error) {
	return j.children.MarshalJSON()
}

func isSpace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func detectBasic(c rune) JsonToken {
	var value JsonToken
	switch c {
	case '{':
		logrus.Debug("detect object")
		value = &JsonObject{
			values: make(map[string]JsonToken),
		}
	case '[':
		value = &JsonArray{}
	case '"':
		value = &JsonString{}
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-', '+', '.', 'e', 'E':
		value = &JsonNumber{}
		value.Append(c)
	case 't', 'f':
		value = &JsonBool{}
		value.Append(c)
	case 'n':
		value = &JsonNull{}
		value.Append(c)
	case 'u':
		value = &JsonUndefined{}
		value.Append(c)
	default:
		return nil
	}
	return value
}
