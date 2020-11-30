package models

import (
	"strconv"

	"github.com/thomasmitchell/baton-resource/utils"
)

type Payload struct {
	Version       Version `json:"version"`
	ReturnSuccess bool    `json:"return_success"`
	Caller        string  `json:"callers"`
	Calls         []Call  `json:"calls"`
	Callbacks     []Call  `json:"callbacks"`
}

type Call struct {
	Key       string `json:"key"`
	Calls     []Call `json:"calls"`
	Callbacks []Call `json:"callbacks"`
}

type Source struct {
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Region          string `json:"region"`
	Key             string `json:"key"`
	Endpoint        string `json:"endpoint"`
	TLSSkipVerify   bool   `json:"tls_skip_verify"`
}

type Version struct {
	Number string `json:"number"`
}

func (v Version) LessThan(v2 Version) bool {
	v1u, err := strconv.ParseUint(v.Number, 10, 64)
	if err != nil {
		utils.Bail("Unsupported number format for (lhs) version `%s': %s", v.Number, err)
	}

	v2u, err := strconv.ParseUint(v2.Number, 10, 64)
	if err != nil {
		utils.Bail("Unsupported number format for (rhs) version `%s': %s", v.Number, err)
	}

	return v1u < v2u
}

func (v Version) Bump() Version {
	vu, err := strconv.ParseUint(v.Number, 10, 64)
	if err != nil {
		utils.Bail("Unsupported number format for version `%s': %s", v.Number, err)
	}

	return Version{
		Number: strconv.FormatUint(vu+1, 10),
	}
}
