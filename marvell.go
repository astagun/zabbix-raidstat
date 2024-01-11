package main

import (
	"fmt"
	"strings"
)

type MarvellVendor struct {
	execPath string
}

// GetControllersIDs - get number of controllers in the system
func (v MarvellVendor) GetControllersIDs() []string {
	inputData := GetCommandOutput(v.execPath, "info", "-o", "hba")
	return GetRegexpAllSubmatch(inputData, "Adapter ID:[\\s]+(.*)")
}

// GetLogicalDrivesIDs - get number of logical drives for controller with ID 'controllerID'
func (v MarvellVendor) GetLogicalDrivesIDs(controllerID string) []string {
	GetCommandOutput(v.execPath, "adapter", "-i", controllerID)
	inputData := GetCommandOutput(v.execPath, "info", "-o", "ld")
	return GetRegexpAllSubmatch(inputData, "id:[\\s]+(.*)")
}

// GetPhysicalDrivesIDs - get number of physical drives for controller with ID 'controllerID'
func (v MarvellVendor) GetPhysicalDrivesIDs(controllerID string) []string {
	GetCommandOutput(v.execPath, "adapter", "-i", controllerID)
	inputData := GetCommandOutput(v.execPath, "info", "-o", "pd")
	return GetRegexpAllSubmatch(inputData, "PD ID:[\\s]+(.*)")
}

// GetControllerStatus - get controller status
func (v MarvellVendor) GetControllerStatus(controllerID string, indent int) []byte {
	type ReturnData struct {
		Status      string `json:"status"`
		ModelNumber string `json:"modelnumber"`
		PartNumber  string `json:"partnumber"`
	}

	inputData := GetCommandOutput(v.execPath, "info", "-o", "hba", "-i", controllerID)

	healthStatuses := []string{}
	for _, v := range []string{
		"Image health",
		"Autoload image health",
		"Boot loader image health",
		"Firmware image health",
		"Boot ROM image health",
		"HBA info image health",
	} {
		if s := GetRegexpSubmatch(inputData, fmt.Sprintf("%s:[\\s]+(.*)", v)); s != "Healthy" {
			healthStatuses = append(healthStatuses, fmt.Sprintf("%s is %s", v, s))
		}
	}

	var status string
	if len(healthStatuses) == 0 {
		status = "OK"
	} else {
		status = strings.Join(healthStatuses, ", ")
	}

	modelnumber := GetRegexpSubmatch(inputData, "ModelNumber:[\\s]+(.*)")
	partnumber := GetRegexpSubmatch(inputData, "PartNumber:[\\s]+(.*)")

	data := ReturnData{
		Status:      TrimSpacesLeftAndRight(status),
		ModelNumber: TrimSpacesLeftAndRight(modelnumber),
		PartNumber:  TrimSpacesLeftAndRight(partnumber),
	}

	return append(MarshallJSON(data, indent), "\n"...)
}

// GetLDStatus - get logical drive status
func (v MarvellVendor) GetLDStatus(controllerID string, deviceID string, indent int) []byte {
	type ReturnData struct {
		Status   string `json:"status"`
		Name     string `json:"name"`
		Size     string `json:"size"`
		RaidMode string `json:"raidmode"`
	}

	GetCommandOutput(v.execPath, "adapter", "-i", controllerID) // set adapter for next commands (mvcli-specific)
	inputData := GetCommandOutput(v.execPath, "info", "-o", "ld", "-i", deviceID)
	status := GetRegexpSubmatch(inputData, "VD status:[\\s]+(.*)")
	name := GetRegexpSubmatch(inputData, "name:[\\s]+(.*)")
	size := GetRegexpSubmatch(inputData, "size:[\\s]+(.*)")
	raidmode := GetRegexpSubmatch(inputData, "RAID mode:[\\s]+(.*)")

	if status == "optimal" {
		status = "OK"
	}

	data := ReturnData{
		Status:   TrimSpacesLeftAndRight(status),
		Name:     TrimSpacesLeftAndRight(name),
		Size:     TrimSpacesLeftAndRight(size),
		RaidMode: TrimSpacesLeftAndRight(raidmode),
	}

	return append(MarshallJSON(data, indent), "\n"...)
}

// GetPDStatus - get physical drive status
func (v MarvellVendor) GetPDStatus(controllerID string, deviceID string, indent int) []byte {
	type ReturnData struct {
		Status          string `json:"status"`
		Model           string `json:"model"`
		FirmwareVersion string `json:"firmwareversion"`
		Size            string `json:"size"`
		CurrentSpeed    string `json:"currentspeed"`
	}

	inputData := GetCommandOutput(v.execPath, "info", "-o", "pd", "-i", deviceID)
	status := GetRegexpSubmatch(inputData, "PD status:[\\s]+(.*)")
	model := GetRegexpSubmatch(inputData, "model:[\\s]+(.*)")
	firmwareversion := GetRegexpSubmatch(inputData, "Firmware version:[\\s]+(.*)")
	size := GetRegexpSubmatch(inputData, "Size:[\\s]+(.*)")
	currentspeed := GetRegexpSubmatch(inputData, "Current speed:[\\s]+(.*)")

	if status == "online" {
		status = "OK"
	}

	data := ReturnData{
		Status:          TrimSpacesLeftAndRight(status),
		Model:           TrimSpacesLeftAndRight(model),
		FirmwareVersion: TrimSpacesLeftAndRight(firmwareversion),
		Size:            TrimSpacesLeftAndRight(size),
		CurrentSpeed:    TrimSpacesLeftAndRight(currentspeed),
	}

	return append(MarshallJSON(data, indent), "\n"...)
}

func NewMarvellVendor(execPath string) Vendor {
	v := MarvellVendor{execPath: execPath}
	return v
}
