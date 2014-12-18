package main

import (
	"sync"

	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/services/template"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
)

type HAProxyFlipperClient struct {
	configurationManager *configuration.ConfigurationManager
	lock                 *sync.Mutex
	state                *types.MasterDetailsCollection
}

func NewHAProxyFlipper(configuration *configuration.ConfigurationManager) *HAProxyFlipperClient {
	state := types.NewMasterDetailsCollection()
	return &HAProxyFlipperClient{configurationManager: configuration, lock: &sync.Mutex{}, state: &state}
}

func (flipper *HAProxyFlipperClient) InitialiseRunningState(details *types.MasterDetailsCollection) {
	flipper.lock.Lock()
	defer flipper.lock.Unlock()

	if details.IsEmpty() {
		logger.Info.Print("Empty initial state : Nothing to do")
		return
	}

	configuration := flipper.configurationManager.GetCurrentConfiguration()
	_, err := flipper.renderAndReload(configuration, details)

	if err != nil {
		logger.Error.Panicf("Unable to initilise and write state : %s", err.Error())
	}
	flipper.state = details
}

func (flipper *HAProxyFlipperClient) Orchestrate(switchEvent types.MasterSwitchedEvent) {

	flipper.lock.Lock()
	defer flipper.lock.Unlock()

	logger.NoteWorthy.Printf("Redis cluster {%s} master failover detected from {%s}:{%d} to {%s}:{%d}.", switchEvent.Name, switchEvent.OldMasterIp, switchEvent.OldMasterPort, switchEvent.NewMasterIp, switchEvent.NewMasterPort)
	logger.NoteWorthy.Printf("Master Switched : %s", util.String(switchEvent))

	configuration := flipper.configurationManager.GetCurrentConfiguration()
	logger.Info.Printf("Current Configuration : %s", util.String(configuration.Clusters))

	cluster, err := configuration.FindClusterByName(switchEvent.Name)

	if err != nil {
		logger.Error.Printf("Redis cluster called %s not found in configuration.", switchEvent.Name)
		return
	}

	logger.Info.Printf("Cluster found : %s", util.String(cluster))

	detail := &types.MasterDetails{
		ExternalPort: cluster.ExternalPort,
		Name:         switchEvent.Name,
		Ip:           switchEvent.NewMasterIp,
		Port:         switchEvent.NewMasterPort}

	flipper.state.AddOrReplace(detail)

	flipper.renderAndReload(configuration, flipper.state)
}

func (flipper *HAProxyFlipperClient) renderAndReload(config configuration.Configuration, details *types.MasterDetailsCollection) (bool, error) {

	configDetails := config.HAProxy

	outputPath := configDetails.OutputPath
	templatepath := configDetails.TemplatePath
	reloadCommand := configDetails.ReloadCommand

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
		logger.Info.Printf("HAProxy output : %s", string(output))
		logger.Error.Printf("Error reloading haproxy with command %s : %s\n", reloadCommand, err.Error())
		return false, err
	}

	logger.NoteWorthy.Printf("HAProxy reload completed.")

	return true, nil
}

func renderTemplate(details *types.MasterDetailsCollection, outputPath string, templatepath string) (bool, error) {

	logger.Info.Printf("Details %s", util.String(details))
	renderedTemplate, err := template.RenderTemplate(templatepath, details)

	if err != nil {
		logger.Error.Printf("Error rendering tempate at %s.", templatepath)
		return false, err
	}

	if util.FileExists(outputPath) {
		newFileHash := util.HashString(renderedTemplate)
		oldFileHash, err := util.HashFile(outputPath)

		if err != nil {
			logger.Error.Printf("Error hashing existing HAProxy config file at %s.", outputPath)
			return false, err
		}

		if newFileHash == oldFileHash {
			logger.NoteWorthy.Printf("Existing config file up todate. New file hash : %s == Old file hash %s. Nothing to do.", newFileHash, oldFileHash)
			return true, nil
		}

		logger.Info.Printf("Updating config file. New file hash : %s != Old file hash %s", newFileHash, oldFileHash)
	}

	err = util.WriteFile(outputPath, renderedTemplate)

	if err != nil {
		logger.Error.Printf("Error writing file to %s : %s\n", outputPath, err.Error())
		return false, err
	}

	return true, nil
}
