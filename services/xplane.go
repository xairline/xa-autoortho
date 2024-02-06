//go:build !test

package services

//go:generate mockgen -destination=./__mocks__/xplane.go -package=mocks -source=xplane.go

import (
	"github.com/joho/godotenv"
	"github.com/xairline/goplane/extra"
	"github.com/xairline/goplane/xplm/processing"
	"github.com/xairline/goplane/xplm/utilities"
	"github.com/xairline/xa-autoortho/utils/logger"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type XplaneService interface {
	// init
	onPluginStateChanged(state extra.PluginState, plugin *extra.XPlanePlugin)
	onPluginStart()
	onPluginStop()
	// flight loop
	flightLoop(elapsedSinceLastCall, elapsedTimeSinceLastFlightLoop float32, counter int, ref interface{}) float32
}

type xplaneService struct {
	Plugin       *extra.XPlanePlugin
	AutoorthoSvc AutoorthoService
	Logger       logger.Logger
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
			Plugin: extra.NewPlugin("X Airline Autoortho Launcher", "com.github.xairline.xa-autoortho", "A plugin that automatically launches AutoOrtho "),
			Logger: logger,
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
	err := godotenv.Load(filepath.Join(pluginPath, "config"))
	if err != nil {
		s.Logger.Errorf("Some error occured. Err: %s", err)
	}
	autoortho_dir := os.Getenv("AUTOORTHO_DIR")
	if autoortho_dir == "" {
		s.Logger.Errorf("AUTOORTHO_DIR is not set")
	}
	s.Logger.Infof("Autoortho dir: %s", autoortho_dir)

	python_executable := os.Getenv("PYTHON_EXECUTABLE")
	if python_executable == "" {
		s.Logger.Warningf("PYTHON_VIRTUALENV is not set")
		// get default python executable
		python_executable = "python3"
		output, err := exec.Command("which", python_executable).CombinedOutput()
		if err != nil {
			s.Logger.Errorf("Can't find python! Some error occured. Err: %s", err)
		}
		python_executable = string(output)
	}
	s.Logger.Infof("Python executable: %s", python_executable)

	s.AutoorthoSvc = NewAutoorthoService(s.Logger, autoortho_dir, python_executable)
	err = s.AutoorthoSvc.LaunchAutoortho()
	if err != nil {
		s.Logger.Errorf("Some error occured. Err: %s", err)
	}

	processing.RegisterFlightLoopCallback(s.flightLoop, -1, nil)
}

func (s *xplaneService) onPluginStop() {
	s.AutoorthoSvc.Umount()
	s.Logger.Info("Plugin stopped")
}

func (s *xplaneService) flightLoop(
	elapsedSinceLastCall,
	elapsedTimeSinceLastFlightLoop float32,
	counter int,
	ref interface{},
) float32 {
	//s.AutoorthoSvc.GetStats()
	return 5
}
