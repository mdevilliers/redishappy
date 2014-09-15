package sentinel

import (
	"github.com/mdevilliers/redishappy/configuration"
	"time"
)

type SentinelTopology struct{
	Sentinels map[string] *SentinelInfo
}

type SentinelInfo struct{
	SentinelLocation string
	LastUpdated time.Time
	KnownClusters []string
	State int
}

type TopologyRequest struct{
	ReplyChannel chan SentinelTopology
}

type SentinelEvent interface{
	GetSentinel() *configuration.Sentinel
}

type SentinelAdded struct{
	sentinel *configuration.Sentinel
}

type SentinelLost struct{
	sentinel *configuration.Sentinel
}

type SentinelPing struct{
    sentinel *configuration.Sentinel
	Clusters []string
}


// TODO : find a better way to implement base type
// functionality
func(s SentinelAdded) GetSentinel() *configuration.Sentinel {
	return s.sentinel
}

func(s SentinelLost) GetSentinel() *configuration.Sentinel {
	return s.sentinel
}

func(s SentinelPing) GetSentinel() *configuration.Sentinel {
	return s.sentinel
}

func(topology SentinelTopology) FindSentinelInfo(sentinel *configuration.Sentinel) (*SentinelInfo, bool) {
	info, ok := topology.Sentinels[sentinel.GetLocation()]
	return info, ok
}

