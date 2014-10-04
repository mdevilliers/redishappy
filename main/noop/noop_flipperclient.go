package main

import (
	"github.com/mdevilliers/redishappy/services/logger"
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/util"
)

type NoOpFlipperClient struct {
}

func NewNoOpFlipper() *NoOpFlipperClient {
	return &NoOpFlipperClient{}
}

func (*NoOpFlipperClient) InitialiseRunningState(state *types.MasterDetailsCollection) {
	logger.Info.Printf("InitialiseRunningState called : %s", util.String(state.Items()))
}

func (*NoOpFlipperClient) Orchestrate(switchEvent types.MasterSwitchedEvent) {
	logger.Info.Printf("Orchestrate called : %s", util.String(switchEvent))
}
