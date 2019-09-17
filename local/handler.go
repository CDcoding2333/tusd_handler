package local

import (
	"log"

	"github.com/tus/tusd"
)

// Config ...
type Config struct {
	FilePath string
	Router   string
	Logger   *log.Logger
}

// NewHandler ...
func NewHandler(conf *Config) (*tusd.Handler, error) {
	store := FileStore{
		Path: conf.FilePath,
	}
	composer := tusd.NewStoreComposer()
	store.UseIn(composer)
	composer.UseGetReader(nil)

	tusdCfg := tusd.Config{
		StoreComposer:           composer,
		BasePath:                conf.Router,
		NotifyUploadProgress:    true,
		NotifyCompleteUploads:   true,
		NotifyTerminatedUploads: true,
		RespectForwardedHeaders: true,
		Logger:                  conf.Logger,
	}

	return tusd.NewHandler(tusdCfg)
}
