//go:build !test

package services

//go:generate mockgen -destination=./__mocks__/xplane.go -package=mocks -source=xplane.go

import (
	"bufio"
	"github.com/xairline/goplane/extra"
	"github.com/xairline/goplane/xplm/utilities"
	"github.com/xairline/xa-autoortho/utils/logger"
	"os"
	"path/filepath"
	"sync"
)

type XplaneService interface {
	// init
	onPluginStateChanged(state extra.PluginState, plugin *extra.XPlanePlugin)
	onPluginStart()
	onPluginStop()
}

type xplaneService struct {
	Plugin            *extra.XPlanePlugin
	AutoorthoSvc      AutoorthoService
	Logger            logger.Logger
	LoadFlightOnStart bool
}

var xplaneSvcLock = &sync.Mutex{}
var xplaneSvc XplaneService

func NewXplaneService(
	logger logger.Logger,
) XplaneService {
	if xplaneSvc != nil {
		logger.Info("Xplane SVC has been initialized already")
		return xplaneSvc
	} else {
		logger.Info("Xplane SVC: initializing")
		xplaneSvcLock.Lock()
		defer xplaneSvcLock.Unlock()

		xplaneSvc := &xplaneService{
			Plugin:            extra.NewPlugin("XA Autoortho Launcher", "com.github.xairline.xa-autoortho", "A plugin that automatically launches AutoOrtho "),
			Logger:            logger,
			LoadFlightOnStart: false,
		}
		xplaneSvc.Plugin.SetPluginStateCallback(xplaneSvc.onPluginStateChanged)
		return xplaneSvc
	}
}

func (s *xplaneService) onPluginStateChanged(state extra.PluginState, plugin *extra.XPlanePlugin) {
	switch state {
	case extra.PluginStart:
		s.onPluginStart()
	case extra.PluginStop:
		s.onPluginStop()
	case extra.PluginEnable:
		s.Logger.Infof("Plugin: %s enabled", plugin.GetName())
	case extra.PluginDisable:
		s.Logger.Infof("Plugin: %s disabled", plugin.GetName())
	}
}

func (s *xplaneService) onPluginStart() {
	s.Logger.Info("Plugin started")
	systemPath := utilities.GetSystemPath()
	pluginPath := filepath.Join(systemPath, "Resources", "plugins", "XA-autoortho")

	// read X-Plane preferences: systempath/Output/preferences/X-Plane.prf
	file, err := os.OpenFile(filepath.Join(systemPath, "Output", "preferences", "X-Plane.prf"), os.O_RDONLY, 0644)
	if err != nil {
		s.Logger.Errorf("Some error occured. Err: %s", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text() == "_show_qfl_on_start 0" {
			s.LoadFlightOnStart = true
			break
		}
	}

	s.Logger.Infof("XP is set to load flight on start-up: %t", s.LoadFlightOnStart)

	s.AutoorthoSvc = NewAutoorthoService(s.Logger, pluginPath, s.LoadFlightOnStart)
	err = s.AutoorthoSvc.LaunchAutoortho()
	if err != nil {
		s.Logger.Errorf("Some error occured. Err: %s", err)
	}
}

func (s *xplaneService) onPluginStop() {
	s.AutoorthoSvc.Umount()
	s.Logger.Info("Plugin stopped")
}
