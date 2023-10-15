package qianfan

import (
	"encoding/json"
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
