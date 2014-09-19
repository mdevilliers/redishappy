package sentinel

import (
	"github.com/mdevilliers/redishappy/types"
	"time"
)

type SentinelTopology struct {
	Sentinels map[string]*SentinelInfo `json:"sentinels"`
}

type SentinelInfo struct {
	SentinelLocation string    `json:"sentinelLocation"`
	LastUpdated      time.Time `json:"lastUpdated"`
	KnownClusters    []string  `json:"knownClusters"`
	State            int       `json:"state"`
}

type TopologyRequest struct {
	ReplyChannel chan SentinelTopology
}

type SentinelEvent interface {
	GetSentinel() *types.Sentinel
}

type SentinelAdded struct {
	Sentinel *types.Sentinel
}

type SentinelLost struct {
	Sentinel *types.Sentinel
}

type SentinelPing struct {
	Sentinel *types.Sentinel
	Clusters []string
}

// TODO : find a better way to implement
// base type functionality
func (s SentinelAdded) GetSentinel() *types.Sentinel {
	return s.Sentinel
}

func (s SentinelLost) GetSentinel() *types.Sentinel {
	return s.Sentinel
}

func (s SentinelPing) GetSentinel() *types.Sentinel {
	return s.Sentinel
}

func (topology SentinelTopology) FindSentinelInfo(sentinel *types.Sentinel) (*SentinelInfo, bool) {
	key := topology.createKey(sentinel)
	info, ok := topology.Sentinels[key]
	return info, ok
}
func (topology SentinelTopology) createKey(sentinel *types.Sentinel) string {
	return sentinel.GetLocation()
}
