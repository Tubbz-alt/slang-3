package daemon

import (
	"net/http"
	"encoding/json"
	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"io/ioutil"
	"path/filepath"
	"strings"
	"gopkg.in/yaml.v2"
	"strconv"
	"math/rand"
)

var port = 12345
var runningInstances = make(map[int64]struct {
	port int
	op   *core.Operator
})
var rnd = rand.New(rand.NewSource(99))

var RunnerService = &DaemonService{map[string]*DaemonEndpoint{
	"/": {func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			type runInstructionJSON struct {
				Cwd   string          `json:"cwd"`
				Fqn   string          `json:"fqn"`
				Props core.Properties `json:"props"`
				Gens  core.Generics   `json:"gens"`
			}

			type outJSON struct {
				URL    string `json:"url,omitempty"`
				Handle int64 `json:"handle,omitempty"`
				Status string `json:"status"`
				Error  *Error `json:"error,omitempty"`
			}

			var data outJSON

			decoder := json.NewDecoder(r.Body)
			var ri runInstructionJSON
			err := decoder.Decode(&ri)
			if err != nil {
				data = outJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0001"}}
				writeJSON(w, &data)
				return
			}

			env := api.NewEnviron(ri.Cwd)
			httpDef, err := api.ConstructHttpEndpoint(env, port, ri.Fqn, ri.Gens, ri.Props)
			if err != nil {
				data = outJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0002"}}
				writeJSON(w, &data)
				return
			}

			packagedOperator := strings.Replace(ri.Fqn+"Packed", ".", string(filepath.Separator), -1) + ".yaml"

			bytes, _ := yaml.Marshal(httpDef)
			ioutil.WriteFile(
				filepath.Join(env.WorkingDir(), packagedOperator),
				bytes,
				0644,
			)

			op, err := env.BuildAndCompileOperator(packagedOperator, nil, nil)
			if err != nil {
				data = outJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0003"}}
				writeJSON(w, &data)
				return
			}

			handle := rnd.Int63()
			runningInstances[handle] = struct {
				port int
				op   *core.Operator
			}{port,op}

			op.Start()
			op.Main().In().Push(nil) // Start server

			data.Status = "success"
			data.Handle = handle
			data.URL = "http://localhost:" + strconv.Itoa(port)

			port++

			writeJSON(w, &data)
		} else if r.Method == "DELETE" {
			type stopInstructionJSON struct {
				Handle int64 `json:"handle"`
			}

			type outJSON struct {
				Status string `json:"status"`
				Error  *Error `json:"error,omitempty"`
			}

			var data outJSON

			decoder := json.NewDecoder(r.Body)
			var si stopInstructionJSON
			err := decoder.Decode(&si)
			if err != nil {
				data = outJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0001"}}
				writeJSON(w, &data)
				return
			}

			if ii, ok := runningInstances[si.Handle]; !ok {
				data = outJSON{Status: "error", Error: &Error{Msg: "Unknown handle", Code: "E0002"}}
				writeJSON(w, &data)
				return
			} else {
				ii.op.Stop()
				delete(runningInstances, si.Handle)

				data.Status = "success"
				writeJSON(w, &data)
			}
		}
	}},
}}
