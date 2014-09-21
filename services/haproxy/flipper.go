package haproxy

import (
	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/template"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
	"sync"
)

type HAProxyFlipperClient struct {
	configuration *configuration.Configuration
	lock          *sync.Mutex
}

func NewFlipper(configuration *configuration.Configuration) *HAProxyFlipperClient {
	return &HAProxyFlipperClient{configuration: configuration, lock: &sync.Mutex{}}
}

func (flipper *HAProxyFlipperClient) Orchestrate(switchEvent types.MasterSwitchedEvent) {

	flipper.lock.Lock()
	defer flipper.lock.Unlock()

	logger.Info.Printf("Redis cluster {%s} master failover detected from {%s}:{%d} to {%s}:{%d}.", switchEvent.Name, switchEvent.OldMasterIp, switchEvent.OldMasterPort, switchEvent.NewMasterIp, switchEvent.NewMasterPort)
	logger.Info.Printf("Master Switched : %s", util.String(switchEvent))
	logger.Info.Printf("Current Configuration : %s", util.String(flipper.configuration.Clusters))

	configuration := flipper.configuration
	path := configuration.HAProxy.OutputPath
	templatepath := configuration.HAProxy.TemplatePath
	reloadCommand := configuration.HAProxy.ReloadCommand

	cluster, err := configuration.FindClusterByName(switchEvent.Name)

	if err != nil {
		logger.Error.Printf("Redis cluster called %s not found in configuration.", switchEvent.Name)
		return
	}

	logger.Info.Printf("Cluster found : %s", util.String(cluster))

	details := types.MasterDetails{
		ExternalPort: cluster.MasterPort,
		Name:         switchEvent.Name,
		Ip:           switchEvent.NewMasterIp,
		Port:         switchEvent.NewMasterPort}

	arr := []types.MasterDetails{details}
	renderedTemplate, err := template.RenderTemplate(templatepath, &arr)

	if err != nil {
		logger.Error.Printf("Error rendering tempate at %s.", templatepath)
		return
	}

	newFileHash := util.HashString(renderedTemplate)
	oldFileHash, err := util.HashFile(path)

	if err != nil {
		logger.Error.Printf("Error hashing existing HAProxy config file at %s.", path)
		return
	}

	if newFileHash == oldFileHash {
		logger.Info.Printf("Existing config file up todate. New file hash : %s == Old file hash %s. Nothing to do.", newFileHash, oldFileHash)
		return
	}

	logger.Info.Printf("Updating config file. New file hash : %s == Old file hash %s", newFileHash, oldFileHash)

	//TODO : check we have permission to update file

	err = template.WriteFile(path, renderedTemplate)

	if err != nil {
		logger.Error.Printf("Error writing file to %s : %s\n", path, err.Error())
		return
	}

	//reload haproxy
	output, err := util.ExecuteCommand(reloadCommand)

	if err != nil {
		logger.Error.Printf("Error reloading haproxy with command %s : %s\n", reloadCommand, err.Error())
		return
	}
	logger.Info.Printf("HAProxy output : %s", string(output))
	logger.Info.Printf("HAProxy reload completed.")
}
