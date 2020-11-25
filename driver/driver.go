package driver

import "github.com/thomasmitchell/baton-resource/models"

type Driver interface {
	Read(key string) (*models.Payload, error)
	Write(key string, payload models.Payload) error
	ReadVersion(key string) (*models.Version, error)
}

func New(cfg models.Source) (Driver, error) {
	return newS3Driver(cfg)
}
