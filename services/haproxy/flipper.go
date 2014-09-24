package haproxy

import (
	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/services/template"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
	"sync"
)

type HAProxyFlipperClient struct {
	configuration *configuration.Configuration
	lock          *sync.Mutex
	state         *types.MasterDetailsCollection
}

func NewFlipper(configuration *configuration.Configuration) *HAProxyFlipperClient {
	return &HAProxyFlipperClient{configuration: configuration, lock: &sync.Mutex{}}
}

func (flipper *HAProxyFlipperClient) InitialiseRunningState(details *types.MasterDetailsCollection) {
	flipper.lock.Lock()
	defer flipper.lock.Unlock()
	_, err := flipper.renderAndReload(details)

	if err != nil {
		logger.Error.Panicf("Unable to initilise and write state : %s", err.Error())
	}
	flipper.state = details
}

func (flipper *HAProxyFlipperClient) Orchestrate(switchEvent types.MasterSwitchedEvent) {

	flipper.lock.Lock()
	defer flipper.lock.Unlock()

	logger.Info.Printf("Redis cluster {%s} master failover detected from {%s}:{%d} to {%s}:{%d}.", switchEvent.Name, switchEvent.OldMasterIp, switchEvent.OldMasterPort, switchEvent.NewMasterIp, switchEvent.NewMasterPort)
	logger.Info.Printf("Master Switched : %s", util.String(switchEvent))
	logger.Info.Printf("Current Configuration : %s", util.String(flipper.configuration.Clusters))

	configuration := flipper.configuration
	cluster, err := configuration.FindClusterByName(switchEvent.Name)

	if err != nil {
		logger.Error.Printf("Redis cluster called %s not found in configuration.", switchEvent.Name)
		return
	}

	logger.Info.Printf("Cluster found : %s", util.String(cluster))

	detail := &types.MasterDetails{
		ExternalPort: cluster.MasterPort,
		Name:         switchEvent.Name,
		Ip:           switchEvent.NewMasterIp,
		Port:         switchEvent.NewMasterPort}

	flipper.state.AddOrReplace(detail)

	flipper.renderAndReload(flipper.state)
}

func (flipper *HAProxyFlipperClient) renderAndReload(details *types.MasterDetailsCollection) (bool, error) {

	configuration := flipper.configuration
	outputPath := configuration.HAProxy.OutputPath
	templatepath := configuration.HAProxy.TemplatePath
	reloadCommand := configuration.HAProxy.ReloadCommand

	ok, err := renderTemplate(details, outputPath, templatepath)

	if ok {
		ok2, err := executeHAproxyCommand(reloadCommand)
		return ok2, err
	}
	return ok, err
}

func executeHAproxyCommand(reloadCommand string) (bool, error) {

	output, err := util.ExecuteCommand(reloadCommand)

	if err != nil {
		logger.Error.Printf("Error reloading haproxy with command %s : %s\n", reloadCommand, err.Error())
		return false, err
	}

	logger.Info.Printf("HAProxy output : %s", string(output))
	logger.Info.Printf("HAProxy reload completed.")

	return true, nil
}

func renderTemplate(details *types.MasterDetailsCollection, outputPath string, templatepath string) (bool, error) {

	renderedTemplate, err := template.RenderTemplate(templatepath, details)

	if err != nil {
		logger.Error.Printf("Error rendering tempate at %s.", templatepath)
		return false, err
	}

	newFileHash := util.HashString(renderedTemplate)
	oldFileHash, err := util.HashFile(outputPath)

	if err != nil {
		logger.Error.Printf("Error hashing existing HAProxy config file at %s.", outputPath)
		return false, err
	}

	if newFileHash == oldFileHash {
		logger.Info.Printf("Existing config file up todate. New file hash : %s == Old file hash %s. Nothing to do.", newFileHash, oldFileHash)
		return true, nil
	}

	logger.Info.Printf("Updating config file. New file hash : %s == Old file hash %s", newFileHash, oldFileHash)

	//TODO : check we have permission to update file

	err = template.WriteFile(outputPath, renderedTemplate)

	if err != nil {
		logger.Error.Printf("Error writing file to %s : %s\n", outputPath, err.Error())
		return false, err
	}

	return true, nil
}
