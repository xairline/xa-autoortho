//go:build !test

package main

import (
	"github.com/xairline/goplane/extra/logging"
	"github.com/xairline/goplane/xplm/plugins"
	"github.com/xairline/goplane/xplm/utilities"
	"github.com/xairline/xa-autoortho/services"
	"github.com/xairline/xa-autoortho/utils/logger"
	"path/filepath"
)

// @BasePath  /apis

func main() {
}

func init() {
	xplaneLogger := logger.NewXplaneLogger()
	plugins.EnableFeature("XPLM_USE_NATIVE_PATHS", true)
	logging.MinLevel = logging.Info_Level
	logging.PluginName = "X Airline Autoortho Launcher"
	// get plugin path
	systemPath := utilities.GetSystemPath()
	pluginPath := filepath.Join(systemPath, "Resources", "plugins", "XA-autoortho")
	xplaneLogger.Infof("Plugin path: %s", pluginPath)

	// entrypoint
	services.NewXplaneService(
		xplaneLogger,
	)
}
