package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/thomasmitchell/baton-resource/driver"
	"github.com/thomasmitchell/baton-resource/models"
	"github.com/thomasmitchell/baton-resource/utils"
)

type Config struct {
	Source  models.Source  `json:"source"`
	Version models.Version `json:"version"`
	Params  Params         `json:"params"`
}

type Params struct {
	ClearStack bool `json:"clear_stack"`
	Skip       bool `json:"skip"`
}

type Output struct {
	Version  models.Version      `json:"version"`
	Metadata []map[string]string `json:"metadata"`
}

func main() {
	dec := json.NewDecoder(os.Stdin)
	cfg := &Config{}
	err := dec.Decode(&cfg)
	if err != nil {
		utils.Bail("Failed to decode input JSON: %s", err)
	}

	output := Output{}

	if !cfg.Params.Skip {
		output = get(cfg.Source, cfg.Params.ClearStack)
	} else {
		output = Output{
			Version: models.Version{Number: "0"},
			Metadata: []map[string]string{
				{"name": "skipped", "value": "true"},
			},
		}
	}

	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(&output)
	if err != nil {
		utils.Bail("Could not encode output JSON: %s", err)
	}
}

func writeJSON(path string, obj interface{}) {
	f, err := os.Create(path)
	if err != nil {
		utils.Bail("Could not open up file at `%s': %s", err)
	}

	defer f.Close()

	enc := json.NewEncoder(f)
	err = enc.Encode(&obj)
	if err != nil {
		utils.Bail("Could not encode JSON to file: %s")
	}
}

func get(source models.Source, clearStack bool) Output {
	drv, err := driver.New(source)
	if err != nil {
		utils.Bail("Failed to initialize driver: %s", err)
	}
	payload, err := drv.Read(source.Key)
	if err != nil {
		utils.Bail("Error when reading from remote: %s", err)
	}

	outputDir := os.Args[1]

	if len(payload.Calls) > 0 {
		writeJSON(filepath.Join(outputDir, "calls"), payload.Calls)
	}

	//clear out calls/callbacks on upstream resource to avoid future calls to this
	// "function" going through the same stack
	if clearStack {
		payload.Calls = []models.Call{}

		err = drv.Write(source.Key, *payload)
		if err != nil {
			utils.Bail("Failed to write payload: %s", err)
		}
	}

	return Output{
		Version: payload.Version,
		Metadata: []map[string]string{
			{"name": "caller", "value": payload.Caller},
		},
	}
}
