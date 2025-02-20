package stream

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

type Builder struct {
	reasonBuilder   strings.Builder
	desBuilder      strings.Builder
	jsonBuilder     JsonStreamer
	jsonParsingStep int
	desReadOffset   int
	jsonSpliterCnt  int
	jsonPrefixCnt   int
	desReadLock     sync.Mutex
}

func (b *Builder) AppendReason(str string) {
	b.reasonBuilder.WriteString(str)
}

func (b *Builder) AppendRune(c rune) {
	//if str has "```json" then start to parse json
	if b.jsonParsingStep == 0 {
		if c == '`' {
			b.jsonSpliterCnt++
			if b.jsonSpliterCnt == 3 {
				b.jsonParsingStep = 1
				logrus.Debug("start to parse json")
				return
			}
		} else {
			b.desBuilder.WriteRune(c)
		}
	} else {
		if c == '`' {
			b.jsonSpliterCnt++
			if b.jsonSpliterCnt == 3 {
				b.jsonParsingStep = 2
				return
			}
		} else {
			if b.jsonPrefixCnt < 4 {
				if "json"[b.jsonPrefixCnt] != byte(c) {
					b.jsonPrefixCnt = 0
				} else {
					b.jsonPrefixCnt++
				}
			} else {
				b.jsonBuilder.AppendRune(c)
			}
		}
	}
}

type Stat struct {
	Reason      string
	Description string
	Json        json.RawMessage
	Finished    bool
}

func (b *Builder) Stat() *Stat {
	return &Stat{
		Reason:      b.reasonBuilder.String(),
		Description: b.DescriptionTail(),
		Json:        b.jsonBuilder.Json(),
		Finished:    b.jsonParsingStep == 3,
	}
}

// func (b *Builder) Append(str string) {
// 	//if str has "```json" then start to parse json
// 	if b.jsonParsingStep == 0 {
// 		if strings.Contains(str, "```json") {
// 			des, json, _ := strings.Cut(str, "```json")
// 			b.desBuilder.WriteString(des)
// 			b.jsonBuilder.Append(json)
// 			b.jsonParsingStep = 1
// 			return
// 		} else {
// 			b.desBuilder.WriteString(str)
// 		}
// 	} else {
// 		if strings.Contains(str, "```") {
// 			json, des, _ := strings.Cut(str, "```")
// 			b.jsonBuilder.Append(json)
// 			b.desBuilder.WriteString(des)
// 			b.jsonParsingStep = 2
// 			return
// 		} else {
// 			b.jsonBuilder.Append(str)
// 		}
// 	}
// }

func (b *Builder) SetFinished() {
	b.jsonParsingStep = 3
}

func (b *Builder) DescriptionTail() string {
	//lock the desReadOffset
	b.desReadLock.Lock()
	//read the rest of the description from desReadOffset
	defer func() {
		b.desReadOffset = b.desBuilder.Len()
		b.desReadLock.Unlock()
	}()
	return b.desBuilder.String()[b.desReadOffset:]
}

func (b *Builder) JsonParsed() bool {
	return b.jsonParsingStep == 2
}
