package logger

import (
	"encoding/json"

	"go.uber.org/zap"
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

func init() {
	// TODO: get this from a config file
	rawJSON := []byte(`{
	  "level": "debug",
	  "encoding": "json",
	  "outputPaths": ["stdout", "/tmp/logs"],
	  "errorOutputPaths": ["stderr"],
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}

	Logger := zap.Must(cfg.Build())
	Sugar = Logger.Sugar()
}
