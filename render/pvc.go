package render

import (
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"
)

var PVCRenderer = Memoise(MakeReduce(ConnectionStorageJoin(MapPV2PVCName, report.PersistentVolume, SelectPersistentVolumeClaim), ConnectionStorageJoin(MapPod2PVCName, report.ApplicationPod, SelectPersistentVolumeClaim), ConnectionStorageJoin(MapPVC2SCName, report.PersistentVolumeClaim, SelectStorageClass), MapEndpoints(endpoint2PVC, report.StorageClass)))

func ConnectionStorageJoin(toPV func(report.Node) string, topology string, selector TopologySelector) Renderer {
	return connectionStorageJoin{toPV: toPV, topology: topology, selector: selector}
}

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

func MapPV2PVCName(m report.Node) string {
	pvcName, ok := m.Latest.Lookup(kubernetes.Claim)
	if !ok {
		pvcName = ""
	}
	return pvcName
}

func MapPod2PVCName(m report.Node) string {
	pvcName, ok := m.Latest.Lookup(kubernetes.VolumeClaimName)
	if !ok {
		pvcName = ""
	}
	return pvcName
}

func MapPVC2SCName(m report.Node) string {
	scName, ok := m.Latest.Lookup(kubernetes.StorageClassName)
	if !ok {
		scName = ""
	}
	return scName
}

// endpoint2PVC returns pvc node ID
func endpoint2PVC(n report.Node) string {
	if pvcNodeID, ok := n.Latest.Lookup(report.MakePersistentVolumeClaimNodeID(n.ID)); ok {
		return pvcNodeID
	}
	return ""
}

type mapStorageEndpoints struct {
	f        endpointMapFunc
	topology string
	selector TopologySelector
}

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
