package qianfan

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestLoadDef(t *testing.T) {
	def, err := openapi3.NewLoader().LoadFromFile("../../deploy_bce/def.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var path = def.Paths["/api/page/update"]
	var post = path.Post
	var req = post.RequestBody.Value
	parameters := schemaToParameterDescriptions(req.Content.Get("application/json").Schema.Value)

	bts, err := json.Marshal(parameters)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bts))

}

func TestWrongJsonFix(t *testing.T) {
	var tested = "```json\n{\n  \"body\": [\n    {\n      \"Name\": \"用户名\",\n      \"Type\": \"string\",\n      \"Description\": \"用户的唯一标识符\",\n      \"fill_by\": \"input\"\n    }\n  ]\n}"
	_, jsonStr, _ := strings.Cut(tested, "```json")
	jsonStr, _, _ = strings.Cut(jsonStr, "```")
	jsonStr = tryToCleanJsonError(strings.TrimSpace(jsonStr))
	var fixed = tryToCleanJsonError(jsonStr)
	t.Log(fixed)
}
