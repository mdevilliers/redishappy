package flipper

import (
	"github.com/mdevilliers/redishappy/configuration"
	"github.com/mdevilliers/redishappy/sentinel"
	"github.com/mdevilliers/redishappy/template"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
	"log"
	"sync"
)

type FlipperClient struct {
	configuration *configuration.Configuration
	lock *sync.Mutex
}

func New(configuration *configuration.Configuration) *FlipperClient {
	return &FlipperClient{configuration : configuration, lock : &sync.Mutex{} }
}

func (flipper *FlipperClient) Orchestrate(switchEvent sentinel.MasterSwitchedEvent){
	
	flipper.lock.Lock()
	defer flipper.lock.Unlock()

	log.Printf( "Redis cluster {%s} master failover detected from {%s}:{%d} to {%s}:{%d}.", switchEvent.Name, switchEvent.OldMasterIp, switchEvent.OldMasterPort,switchEvent.NewMasterIp, switchEvent.NewMasterPort)
	
	log.Printf("Master Switched : %s\n",  util.String(switchEvent))
	log.Printf("Current Configuration : %s\n",  util.String(flipper.configuration.Clusters))

	configuration :=flipper.configuration
	path := configuration.HAProxy.OutputPath
	templatepath := configuration.HAProxy.TemplatePath
	reloadCommand := configuration.HAProxy.ReloadCommand

	cluster, err := configuration.FindClusterByName(switchEvent.Name)

	if err != nil {
		log.Printf("Redis cluster called %s not found in configuration.", switchEvent.Name)
		return
	}
	
	log.Printf("Cluster found : %s\n",  util.String(cluster))

	details := types.MasterDetails {
						ExternalPort: cluster.MasterPort,
						Name :switchEvent.Name,
						Ip : switchEvent.NewMasterIp,
						Port :switchEvent.NewMasterPort}

	//render template
	// TODO : look into HAProxy supporting multiple config files....

	arr := []types.MasterDetails{details}
	renderedTemplate, err := template.RenderTemplate(templatepath, &arr)

	if err != nil {
		log.Printf("Error rendering tempate at %s.", templatepath)
		return
	}

	//get hash of the new and old files
	newFileHash := util.HashString(renderedTemplate)
	oldFileHash, err := util.HashFile(path)

	if err != nil {
		log.Printf("Error hashing existing HAProxy config file at %s.", path)
		return
	}

	if newFileHash == oldFileHash {
		log.Printf("Existing config file up todate. New file hash : %s == Old file hash %s. Nothing to do.", newFileHash, oldFileHash )
		return
	}

	log.Printf("Updating config file. New file hash : %s == Old file hash %s", newFileHash, oldFileHash )

	//TODO : check we have permission to update file

	err = template.WriteFile(path, renderedTemplate)

	if err != nil {
		log.Printf("error writing file to %s : %s\n", path, err.Error())
		return
	}

	//reload haproxy
	output, err := util.ExecuteCommand(reloadCommand)

	if err != nil {
		log.Printf("error reloading haproxy with command %s : %s\n", reloadCommand, err.Error())
		return
	}
	log.Printf("HAProxy output : %s", string(output))
	log.Printf("HAProxy reload completed.")	
}