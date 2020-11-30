package main

import (
	"encoding/json"
	"os"

	"github.com/thomasmitchell/baton-resource/driver"
	"github.com/thomasmitchell/baton-resource/models"
	"github.com/thomasmitchell/baton-resource/utils"
)

type Config struct {
	Source  models.Source  `json:"source"`
	Version models.Version `json:"version"`
}

func main() {
	dec := json.NewDecoder(os.Stdin)
	cfg := &Config{}
	err := dec.Decode(&cfg)
	if err != nil {
		utils.Bail("Failed to decode input JSON: %s", err)
	}

	drv, err := driver.New(cfg.Source)
	remoteVersion, err := drv.ReadVersion(cfg.Source.Key)
	if err != nil {
		utils.Bail("Error when reading from remote: %s", err)
	}

	if cfg.Version.Number == "" {
		cfg.Version.Number = "0"
	}

	if cfg.Version.LessThan(*remoteVersion) {
		enc := json.NewEncoder(os.Stdout)
		err = enc.Encode([]models.Version{*remoteVersion})
		if err != nil {
			utils.Bail("Error encoding output value: %s", err)
		}
	}
}
