package metadata

import (
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/serf/serf"
)

// called by
// agent/consul/autopilot.go/IsServer
// agent/consul/merge.go/NotifyMerge
// agent/metadata/server.go/IsConsulServer
// Build extracts the Consul version info for a member.
func Build(m *serf.Member) (*version.Version, error) {
	str := versionFormat.FindString(m.Tags["build"])
	return version.NewVersion(str)
}
