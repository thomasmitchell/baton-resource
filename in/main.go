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
}

type Output struct {
	Version  models.Version    `json:"version"`
	Metadata map[string]string `json:"metadata"`
}

func main() {
	dec := json.NewDecoder(os.Stdin)
	cfg := &Config{}
	err := dec.Decode(&cfg)
	if err != nil {
		utils.Bail("Failed to decode input JSON: %s", err)
	}

	drv, err := driver.New(cfg.Source)
	if err != nil {
		utils.Bail("Failed to initialize driver: %s", err)
	}
	payload, err := drv.Read(cfg.Source.Key)
	if err != nil {
		utils.Bail("Error when reading from remote: %s", err)
	}

	outputDir := os.Args[1]

	writeJSON := func(path string, obj interface{}) {
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

	if len(payload.Calls) > 0 {
		writeJSON(filepath.Join(outputDir, "calls"), payload.Calls)

	}

	if len(payload.Callbacks) > 0 {
		writeJSON(filepath.Join(outputDir, "callbacks"), payload.Callbacks)
	}

	//clear out calls/callbacks on upstream resource to avoid future calls to this
	// "function" going through the same stack
	if cfg.Params.ClearStack {
		payload.Calls = []models.Call{}
		payload.Callbacks = []models.Call{}

		err = drv.Write(cfg.Source.Key, *payload)
		if err != nil {
			utils.Bail("Failed to write payload: %s", err)
		}
	}

	output := Output{
		Version: payload.Version,
		Metadata: map[string]string{
			"caller": payload.Caller,
		},
	}
	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(&output)
	if err != nil {
		utils.Bail("Could not encode output JSON: %s", err)
	}
}
