package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/utils"
	"net/http"
	"bytes"
	"io/ioutil"
)

var netHTTPClientCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In:  func() core.TypeDef {
					req := HTTP_REQUEST_DEF.Copy()
					delete(req.Map, "params")
					delete(req.Map, "path")
					delete(req.Map, "query")
					req.Map["url"] = &core.TypeDef{Type: "string"}
					return req
				}(),
				Out: HTTP_RESPONSE_DEF.Copy(),
			},
		},
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

			req := i.(map[string]interface{})
			method := req["method"].(string)
			url := req["url"].(string)
			body := req["body"].(utils.Binary)
			headers := req["headers"].([]interface{})

			r, err := http.NewRequest(method, url, bytes.NewReader(body))
			if err != nil {
				out.Push(nil)
				continue
			}
			for _, header := range headers {
				entry := header.(map[string]interface{})
				if entry["value"] == nil {
					continue
				}
				r.Header.Set(entry["key"].(string), entry["value"].(string))
			}

			resp, err := http.DefaultClient.Do(r)
			if err != nil {
				out.Push(nil)
				continue
			}

			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				out.Push(nil)
				continue
			}

			out.Map("status").Push(float64(resp.StatusCode))
			out.Map("body").Push(utils.Binary(respBody))

			out.Map("headers").PushBOS()
			for key := range resp.Header {
				out.Map("headers").Stream().Map("key").Push(key)
				out.Map("headers").Stream().Map("value").Push(resp.Header.Get(key))
			}
			out.Map("headers").PushEOS()
		}
	},
}