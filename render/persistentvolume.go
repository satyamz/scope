package render

import (
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"
)

// PersistentVolumeRenderer is the common renderer for all the storage components.
var PersistentVolumeRenderer = Memoise(
	MakeReduce(
		ConnectionStorageJoin(
			Map2PVCName,
			report.PersistentVolume,
			SelectPersistentVolumeClaim,
		),
		MapEndpoints(
			Map2SC,
			report.PersistentVolumeClaim,
		)))

// ConnectionStorageJoin returns connectionStorageJoin object
func ConnectionStorageJoin(toPV func(report.Node) string, topology string, selector TopologySelector) Renderer {
	return connectionStorageJoin{toPV: toPV, topology: topology, selector: selector}
}

// connectionStorageJoin holds the information about mapping of storage components
// along with TopologySelector
type connectionStorageJoin struct {
	toPV     func(report.Node) string
	topology string
	selector TopologySelector
}

func (c connectionStorageJoin) Render(rpt report.Report) Nodes {
	inputNodes := TopologySelector(c.topology).Render(rpt).Nodes

	var pvNodes = map[string]string{}
	for _, n := range inputNodes {
		pvcName := c.toPV(n)
		pvNodes[pvcName] = n.ID
	}

	return MapStorageEndpoints(
		func(m report.Node) string {
			pvcName, ok := m.Latest.Lookup(kubernetes.Name)
			if !ok {
				return ""
			}
			id := pvNodes[pvcName]
			return id
		}, c.topology, c.selector).Render(rpt)
}

// Map2PVCName returns PVC name for the given Pod.
func Map2PVCName(m report.Node) string {
	pvcName, ok := m.Latest.Lookup(kubernetes.VolumeClaim)
	if !ok {
		pvcName = ""
	}
	return pvcName
}

// Map2SC returns pvc node ID
func Map2SC(n report.Node) string {
	if pvcNodeID, ok := n.Latest.Lookup(report.MakePersistentVolumeClaimNodeID(n.ID)); ok {
		return pvcNodeID
	}
	return ""
}

// mapStorageEndpoints is the Renderer for rendering storage components together.
type mapStorageEndpoints struct {
	f        endpointMapFunc
	topology string
	selector TopologySelector
}

// MapStorageEndpoints instantiates mapStorageEndpoints and returns same
func MapStorageEndpoints(f endpointMapFunc, topology string, selector TopologySelector) Renderer {
	return mapStorageEndpoints{f: f, topology: topology, selector: selector}
}

func (e mapStorageEndpoints) Render(rpt report.Report) Nodes {

	endpoints := e.selector.Render(rpt)
	ret := newJoinResults(TopologySelector(e.topology).Render(rpt).Nodes)

	for _, n := range endpoints.Nodes {
		if id := e.f(n); id != "" {
			ret.addChild(n, id, e.topology)
		}
	}
	return ret.storageResult(endpoints)
}
