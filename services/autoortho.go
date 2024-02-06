package services

//go:generate mockgen -destination=./__mocks__/Autoortho.go -package=mocks -source=Autoortho.go
import (
	"context"
	"github.com/xairline/xa-snow/utils/logger"
	"gopkg.in/ini.v1"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"
	"sync"
)

var autoorthoSvcLock = &sync.Mutex{}
var autoorthoSvc AutoorthoService

type AutoorthoService interface {
	LaunchAutoortho() error
	getMounts() []string
	Umount()
}

type autoorthoService struct {
	Logger logger.Logger
	dir    string
	pyPath string
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func (a *autoorthoService) Umount() {
	a.cancel()
	a.wg.Wait()
}

func (a *autoorthoService) LaunchAutoortho() error {
	mounts := a.getMounts()
	a.Logger.Infof("Mounts: %v", mounts)
	autoUnmount := os.Getenv("AUTO_UNMOUNT")
	a.Logger.Infof("Auto unmount: %s", autoUnmount)
	user, _ := user.Current()
	for _, mount := range mounts {
		a.wg.Add(1)
		go func(mount string) {
			defer a.wg.Done()
			file, err := os.OpenFile(path.Join(user.HomeDir, "autoortho.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatalf("failed to open file: %v", err)
			}
			defer file.Close()
			cmd := exec.Command(
				a.pyPath,
				a.dir+"/autoortho/autoortho_fuse.py",
				strings.Split(mount, "|")[0],
				strings.Split(mount, "|")[1],
			)
			cmd.Stdout = file
			cmd.Stderr = file
			// Start the command without waiting for it to complete
			if err := cmd.Start(); err != nil {
				a.Logger.Errorf("Error starting command: %v", err)
				return
			}
			select {
			case <-a.ctx.Done():
				if autoUnmount == "true" {
					if cmd.Process != nil {
						err := cmd.Process.Signal(os.Interrupt)
						if err != nil {
							a.Logger.Errorf("Error sending interrupt: %v", err)
						}
					}
				}
			}
			// Wait for the command to finish
			err = cmd.Wait()
			if err != nil {
				a.Logger.Errorf("Command finished with error: %v", err)
			} else {
				a.Logger.Infof("Command finished successfully")
			}
		}(mount)
	}

	return nil
}

func (a *autoorthoService) getMounts() []string {
	// read toml file to bytes
	var res []string
	{
	}
	user, _ := user.Current()
	// Load the configuration file
	cfg, err := ini.Load(path.Join(user.HomeDir, ".autoortho"))
	if err != nil {
		log.Fatalf("Fail to read file: %v", err)
	}

	// Get a specific section
	section, err := cfg.GetSection("paths")
	if err != nil {
		log.Fatalf("Fail to get section: %v", err)
	}

	// Get key-value pairs from the section
	XPlanePath := section.Key("xplane_path").String()
	SceneryPath := section.Key("scenery_path").String()
	a.Logger.Infof("XPlanePath: %s, SceneryPath: %s", XPlanePath, SceneryPath)

	folders, _ := os.ReadDir(SceneryPath + "/z_autoortho/scenery")
	for _, region := range folders {
		if region.IsDir() {
			root := path.Join(SceneryPath, "z_autoortho", "scenery", region.Name())
			mount := path.Join(XPlanePath, "Custom Scenery", region.Name())
			res = append(res, root+"|"+mount)
		}
	}

	return res
}

func NewAutoorthoService(logger logger.Logger, dir string, pyPath string) AutoorthoService {
	if autoorthoSvc != nil {
		logger.Info("Autoortho SVC has been initialized already")
		return autoorthoSvc
	} else {
		logger.Info("Autoortho SVC: initializing")
		autoorthoSvcLock.Lock()
		defer autoorthoSvcLock.Unlock()
		logger.Infof("Autoortho SVC: initializing with folder %s", dir)
		ctx, cancel := context.WithCancel(context.Background())
		autoorthoSvc = &autoorthoService{
			Logger: logger,
			dir:    dir,
			pyPath: pyPath,
			ctx:    ctx,
			cancel: cancel,
		}
		return autoorthoSvc
	}
}
