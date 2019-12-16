package elem

import (
	"math/rand"
	"time"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var randRangeId = uuid.MustParse("30e3a788-b5ec-4c0f-9338-4a78fe63bd9f")
var randRangeCfg = &builtinConfig{
	blueprint: core.Blueprint{
		Id: randRangeId,
		Meta: core.BlueprintMetaDef{
			Name:             "generate a random number between two given values",
			ShortDescription: "generate a random number between two given values, including those values e.g. [a, b]",
			Icon:             "random",
			Tags:             []string{"data", "random"},
			DocURL:           "https://bitspark.de/slang/docs/operator/rand-range",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"min": {
							Type: "number",
						},
						"max": {
							Type: "number",
						},
					},
				},
				Out: core.TypeDef{
					Type: "number",
				},
			},
		},
		PropertyDefs: map[string]*core.TypeDef{},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			data := i.(map[string]interface{})
			max := data["max"].(int)
			min := data["min"].(int)
			// This generates values for the following interval [min, max] e.g. {min <= x <= max}
			out.Push(min + rand.Intn(max-min+1))
		}
	},
	opConnFunc: func(op *core.Operator, dst, src *core.Port) error {
		// So we seed random when the operator gets assembled
		rand.Seed(time.Now().UnixNano())
		return nil
	},
}
