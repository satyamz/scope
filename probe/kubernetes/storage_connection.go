package kubernetes

import (
	log "github.com/Sirupsen/logrus"
	"github.com/weaveworks/scope/report"
)

// volumeName  holds volume name and claim
type volumeTuple struct {
	VolumeName, ClaimName string
}

// storageConnection the storage connection tracker.
type storageConnection struct {
	HostID   string
	HostName string
}

func newstorageConnection(hostID, hostName string) storageConnection {
	sc := storageConnection{
		HostID:   hostID,
		HostName: hostName,
	}
	return sc
}

// ReportConnections calls trackers according to the configuration.
func (t *storageConnection) ReportConnections(rpt *report.Report) {
	hostNodeID := report.MakeHostNodeID(t.HostID) // Create Host Id
	// How to find whether current node is PVC or PV or even current node itself?

}

func (t *storageConnection) addStorageConnection(rpt *report.Report, IsPVName bool, vt volumeTuple, namespaceID string, extraFromNode, extraToNode map[string]string) {
	// Based on the current Node perform following operation
	// IsPV if node is PV then find volumeName
	// IsPVC if node is PVC then find ClaimName
	// Create dummy node with namespace, hostID
	var (
		fromNode = t.makeVolumeNode(namespaceID, vt.VolumeName, vt.ClaimName, extraFromNode)
		toNode   = t.makeVolumeNode(namespaceID, vt.VolumeName, vt.ClaimName, extraToNode)
	)
	log.Infof("%+v", toNode)
	rpt.Endpoint.AddNode(fromNode.WithAdjacent(toNode.ID))
	rpt.Endpoint.AddNode(toNode)
	// Write a different struct to containing ClaimName and VolumeName
	//Based on the node pass the mapping here to addVolume.
	//If node is PVC or PV then add volumeName
	t.addVolume(rpt, vt.VolumeName)

	//If node is PVC then add ClaimName mandatory
	t.addVolume(rpt, vt.VolumeName)

}

func (t *storageConnection) makeVolumeNode(namespaceID string, volumeName string, volumeClaim string, extra map[string]string) report.Node {

	node := report.MakeNodeWith(report.MakeVolumeNodeID(t.HostID, namespaceID, volumeParam), nil)
	if extra != nil {
		node = node.WithLatests(extra)
	}
	return node
}

// addVolume should add VolumeName for the current name
func (t *storageConnection) addVolume(rpt *report.Report, volumeParam string) {
	// Should create a map of NodeID and VolumeName
}
