package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thomasmitchell/baton-resource/driver"
	"github.com/thomasmitchell/baton-resource/models"
	"github.com/thomasmitchell/baton-resource/utils"
)

type Config struct {
	Source models.Source `json:"source"`
	Params Params        `json:"params"`
}

type Params struct {
	Caller          string        `json:"caller"`
	ContinueFrom    string        `json:"continue_from"`
	CallbackFrom    string        `json:"callback_from"`
	CallbackFailure bool          `json:"callback_failure"`
	Calls           []models.Call `json:"calls"`
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

	if len(cfg.Params.Calls) > 0 && cfg.Params.ContinueFrom != "" {
		utils.Bail("Cannot have ContinueFrom and Calls specified in params")
	}

	packaged := []keyPayloadPair{}
	if len(cfg.Params.Calls) > 0 {
		packaged = append(
			packaged,
			packageCalls(cfg.Params.Calls, false)...,
		)
	}

	if cfg.Params.ContinueFrom != "" {
		calls, err := getContinuations(cfg.Params.ContinueFrom)
		if err != nil {
			utils.Bail(err.Error())
		}

		packaged = append(
			packaged,
			packageCalls(calls, false)...,
		)
	}

	if cfg.Params.CallbackFrom != "" {
		calls, err := getCallbacks(cfg.Params.CallbackFrom)
		if err != nil {
			utils.Bail(err.Error())
		}

		packaged = append(
			packaged,
			packageCalls(calls, cfg.Params.CallbackFailure)...,
		)
	}

	drv, err := driver.New(cfg.Source)
	if err != nil {
		utils.Bail("Failed to initialize driver: %s", err)
	}

	for _, pkg := range packaged {
		curVersion, err := drv.ReadVersion(pkg.Key)
		if err != nil {
			utils.Bail("Could not retrieve version from key `%s': %s", err)
		}

		pkg.Payload.Version = curVersion.Bump()
		err = drv.Write(pkg.Key, pkg.Payload)
		if err != nil {
			utils.Bail("Could not write version to key `%s': %s", pkg.Key, err)
		}
	}

	//Get this resource's version (Because concourse will want it for the git
	// step even though we don't need it...)
	version, err := drv.ReadVersion(cfg.Source.Key)
	if err != nil {
		utils.Bail("Erred getting our own version")
	}

	output := Output{
		Version:  *version,
		Metadata: map[string]string{},
	}
	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(&output)
	if err != nil {
		utils.Bail("Could not encode output JSON: %s", err)
	}
}

func getContinuations(dir string) ([]models.Call, error) {
	return readJSONFromFile(filepath.Join(dir, "calls"))
}

func getCallbacks(dir string) ([]models.Call, error) {
	return readJSONFromFile(filepath.Join(dir, "callbacks"))
}

func readJSONFromFile(path string) ([]models.Call, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			if _, statErr := os.Stat(filepath.Dir(path)); statErr != nil {
				if os.IsNotExist(statErr) {
					//Resource input isn't there
					return nil,
						fmt.Errorf(
							"Missing resource input `%s'",
							filepath.Base(filepath.Dir(path)),
						)
				}
			}

			//Path file isn't there, which can just mean there's nothing to return
			return nil, nil
		}

		//Actual IO error
		return nil, fmt.Errorf("Error opening file `%s': %s", path, err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	ret := []models.Call{}
	err = dec.Decode(&ret)
	if err != nil {
		return nil, fmt.Errorf("Error decoding JSON from file `%s': %s", path, err)
	}

	return ret, nil
}

type keyPayloadPair struct {
	Key     string
	Payload models.Payload
}

//Version is not set yet after this call
func packageCalls(calls []models.Call, failed bool) []keyPayloadPair {
	ret := []keyPayloadPair{}
	for _, call := range calls {
		ret = append(
			ret,
			keyPayloadPair{
				Key: call.Key,
				Payload: models.Payload{
					ReturnSuccess: !failed,
					Caller:        utils.CallerName(),
					Calls:         call.Calls,
					Callbacks:     call.Callbacks,
				},
			},
		)
	}

	return ret
}
