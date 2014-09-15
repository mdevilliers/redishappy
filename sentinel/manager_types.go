package sentinel

import (
	"github.com/mdevilliers/redishappy/types"
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
	GetSentinel() *types.Sentinel
}

type SentinelAdded struct{
	sentinel *types.Sentinel
}

type SentinelLost struct{
	sentinel *types.Sentinel
}

type SentinelPing struct{
    sentinel *types.Sentinel
	Clusters []string
}

// TODO : find a better way to implement base type
// functionality
func(s SentinelAdded) GetSentinel() *types.Sentinel {
	return s.sentinel
}

func(s SentinelLost) GetSentinel() *types.Sentinel {
	return s.sentinel
}

func(s SentinelPing) GetSentinel() *types.Sentinel {
	return s.sentinel
}

func(topology SentinelTopology) FindSentinelInfo(sentinel *types.Sentinel) (*SentinelInfo, bool) {
	info, ok := topology.Sentinels[sentinel.GetLocation()]
	return info, ok
}

