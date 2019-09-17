package cloud

import (
	"log"

	"github.com/tus/tusd"
)

// Config ...
type Config struct {
	Router  string
	Bucket  string
	Service S3API
	Logger  *log.Logger
}

// NewHandler ...
func NewHandler(conf *Config) (*tusd.Handler, error) {
	store := New(conf.Bucket, conf.Service)
	storeComposer := tusd.NewStoreComposer()
	store.UseIn(storeComposer)
	storeComposer.UseGetReader(nil)

	tusdCfg := tusd.Config{
		StoreComposer:           storeComposer,
		BasePath:                conf.Router,
		NotifyUploadProgress:    true,
		NotifyCompleteUploads:   true,
		NotifyTerminatedUploads: true,
		RespectForwardedHeaders: true,
		Logger:                  conf.Logger,
	}

	return tusd.NewHandler(tusdCfg)
}
