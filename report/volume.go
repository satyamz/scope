package report

//VolumeRecord maitains volume name
type VolumeRecord struct {
	VolumeName string `json:"volume_name,omitempty"`
	VolumeClaim string `json:"volume_claim,omitempty"`
}
