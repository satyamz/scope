package render

import (
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"
)

var PVCRenderer = Memoise(MakeReduce(ConnectionStorageJoin(MapPVC2PVName, report.PersistentVolume)))

func ConnectionStorageJoin(toPV func(report.Node) string, topology string) Renderer {
	return connectionStorageJoin{toPV: toPV, topology: topology}
}

type connectionStorageJoin struct {
	toPV     func(report.Node) string
	topology string
}

func (c connectionStorageJoin) Render(rpt report.Report) Nodes {
	inputNodes := TopologySelector(c.topology).Render(rpt).Nodes //All PV nodes

	var pvNodes = map[string]string{} // Map to store information
	for _, n := range inputNodes {
		for _, pvcName := range c.toPV(n) {
			pvNodes[string(pvcName)] = n.ID
		}
	}
	return MapStorageEndpoints(
		func(m report.Node) string {

			//Function to get PV id for PVCName in PVC Node in argument
			pvName, ok := m.Latest.Lookup(kubernetes.Name)
			if !ok {
				return "" //pvName not found then return empty id
			}
			id := pvNodes[pvName] // Return PVC ID
			return id
		}, c.topology).Render(rpt)
}

func MapPVC2PVName(m report.Node) string {
	//return PVName associated with the given PVC Node
	pvcName, ok := m.Latest.Lookup(kubernetes.PersistentVolumeClaimName)
	if !ok {
		pvcName = ""
	}
	return pvcName
}
