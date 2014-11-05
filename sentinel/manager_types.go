package sentinel

import (
	"time"

	"github.com/mdevilliers/redishappy/types"
)

type SentinelTopology struct {
	Sentinels map[string]*SentinelInfo `json:"sentinels"`
}

type SentinelInfo struct {
	SentinelLocation string    `json:"sentinelLocation"`
	LastUpdated      time.Time `json:"lastUpdated"`
	Clusters         []string  `json:"clusters"`
	State            int       `json:"state"`
}

type TopologyRequest struct {
	ReplyChannel chan SentinelTopology
}

type SentinelEvent interface {
	GetSentinel() types.Sentinel
}

type SentinelAdded struct {
	Sentinel types.Sentinel
}

type SentinelLost struct {
	Sentinel types.Sentinel
}

type SentinelUnknown struct {
	Sentinel types.Sentinel
}

type SentinelPing struct {
	Sentinel types.Sentinel
}

type SentinelClustersMonitoredUpdate struct {
	Sentinel types.Sentinel
	Clusters []string
}

func (s SentinelAdded) GetSentinel() types.Sentinel {
	return s.Sentinel
}

func (s SentinelLost) GetSentinel() types.Sentinel {
	return s.Sentinel
}

func (s SentinelPing) GetSentinel() types.Sentinel {
	return s.Sentinel
}

func (s SentinelUnknown) GetSentinel() types.Sentinel {
	return s.Sentinel
}

func (s SentinelClustersMonitoredUpdate) GetSentinel() types.Sentinel {
	return s.Sentinel
}

func (topology SentinelTopology) FindSentinelInfo(sentinel types.Sentinel) (*SentinelInfo, bool) {
	key := topology.createKey(sentinel)
	info, ok := topology.Sentinels[key]
	return info, ok
}
func (topology SentinelTopology) createKey(sentinel types.Sentinel) string {
	return sentinel.GetLocation()
}
