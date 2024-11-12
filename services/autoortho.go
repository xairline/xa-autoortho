package services

//go:generate mockgen -destination=./__mocks__/Autoortho.go -package=mocks -source=Autoortho.go
import (
	"bufio"
	"bytes"
	"context"
	"github.com/xairline/xa-autoortho/utils/logger"
	"gopkg.in/ini.v1"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"
	"sync"
	"time"
)

var autoorthoSvcLock = &sync.Mutex{}
var autoorthoSvc AutoorthoService

type AutoorthoService interface {
	LaunchAutoortho() error
	getMounts() []string
	Umount()
}

type autoorthoService struct {
	Logger     logger.Logger
	pluginPath string
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func (a *autoorthoService) Umount() {
	a.cancel()
	a.wg.Wait()
}

func (a *autoorthoService) LaunchAutoortho() error {
	mounts := a.getMounts()
	a.Logger.Infof("Mounts: %v", mounts)
	current, _ := user.Current()

	// Create a new WaitGroup for cmd.Start()
	var cmdStartWG sync.WaitGroup
	cmdStartWG.Add(len(mounts))

	for _, mount := range mounts {
		a.wg.Add(1)
		go func(mount string) {
			defer a.wg.Done()
			poisonFile := path.Join(strings.Split(mount, "|")[0], ".poison")
			file, err := os.OpenFile(path.Join(current.HomeDir, ".autoortho-data", "logs", "autoortho.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatalf("failed to open file: %v", err)
			}
			defer file.Close()
			os.Remove(poisonFile)

			// incase we have left over fuse mount
			err = exec.Command("umount", strings.Split(mount, "|")[1]).Run()
			if err != nil {
				a.Logger.Errorf("Warning unmounting: %v", err)
			}
			cmd := exec.Command(
				a.pluginPath+"/autoortho_fuse",
				strings.Split(mount, "|")[0],
				strings.Split(mount, "|")[1],
			)
			cmd.Stdout = file
			cmd.Stderr = file
			// Start the command without waiting for it to complete
			if err := cmd.Start(); err != nil {
				a.Logger.Errorf("Error starting command: %v", err)
				// Decrement the cmdStartWG even if there's an error
				cmdStartWG.Done()
				return
			}
			// wait until it is actually mounted by checking a file
			for {
				isFuseMount, err := a.isFuseMountPoint(strings.Split(mount, "|")[1])
				if err != nil {
					a.Logger.Errorf("Error checking mount point: %v", err)
					time.Sleep(1 * time.Second)
					continue
				}

				if !isFuseMount {
					a.Logger.Infof("Autoortho service is not ready: %s", strings.Split(mount, "|")[1])
					time.Sleep(1 * time.Second)
				} else {
					a.Logger.Infof("Autoortho service is ready: %s", strings.Split(mount, "|")[1])
					break
				}
			}
			// Indicate that cmd.Start() has completed
			cmdStartWG.Done()
			select {
			case <-a.ctx.Done():
				a.Logger.Infof("Autoortho service is stopping: %s", strings.Split(mount, "|")[1])
				a.wg.Done()
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
	// Wait for all cmd.Start() calls to complete
	cmdStartWG.Wait()
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

func NewAutoorthoService(logger logger.Logger, pluginPath string) AutoorthoService {
	if autoorthoSvc != nil {
		logger.Info("Autoortho SVC has been initialized already")
		return autoorthoSvc
	} else {
		logger.Info("Autoortho SVC: initializing")
		autoorthoSvcLock.Lock()
		defer autoorthoSvcLock.Unlock()
		ctx, cancel := context.WithCancel(context.Background())
		autoorthoSvc = &autoorthoService{
			Logger:     logger,
			pluginPath: pluginPath,
			ctx:        ctx,
			cancel:     cancel,
			wg:         sync.WaitGroup{},
		}
		return autoorthoSvc
	}
}

func (a *autoorthoService) isFuseMountPoint(path string) (bool, error) {
	cmd := exec.Command("mount")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		// The mount command output typically has lines like:
		// <device> on <mountpoint> (<fstype>, <options>)
		if strings.Contains(line, " on "+path+" ") {
			// Extract the filesystem type from the line
			start := strings.Index(line, "(")
			end := strings.Index(line, ")")
			if start != -1 && end != -1 && end > start {
				fsInfo := line[start+1 : end]
				// fsInfo might look like "fusefs_osxfuse, local, nodev, nosuid, synchronous"
				fsType := strings.Fields(fsInfo)[0]
				if strings.HasPrefix(fsType, "fuse") || strings.Contains(fsType, "osxfuse") || strings.Contains(fsType, "macfuse") {
					return true, nil
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}
