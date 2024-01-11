package main

import (
	"fmt"
)

type HPVendor struct {
	execPath string
}

// GetControllersIDs - get number of controllers in the system
func (v HPVendor) GetControllersIDs() []string {
	inputData := GetCommandOutput(v.execPath, "ctrl", "all", "show")
	return GetRegexpAllSubmatch(inputData, "in Slot (.*?)[\\s]")
}

// GetLogicalDrivesIDs - get number of logical drives for controller with ID 'controllerID'
func (v HPVendor) GetLogicalDrivesIDs(controllerID string) []string {
	inputData := GetCommandOutput(v.execPath, "ctrl", fmt.Sprintf("slot=%s", controllerID), "ld", "all", "show")
	return GetRegexpAllSubmatch(inputData, "logicaldrive (.*?)[\\s]")
}

// GetPhysicalDrivesIDs - get number of physical drives for controller with ID 'controllerID'
func (v HPVendor) GetPhysicalDrivesIDs(controllerID string) []string {
	inputData := GetCommandOutput(v.execPath, "ctrl", fmt.Sprintf("slot=%s", controllerID), "pd", "all", "show")
	return GetRegexpAllSubmatch(inputData, "physicaldrive (.*?)[\\s]")
}

// GetControllerStatus - get controller status
func (v HPVendor) GetControllerStatus(controllerID string, indent int) []byte {
	type ReturnData struct {
		Status        string `json:"status"`
		Model         string `json:"model"`
		BatteryStatus string `json:"batterystatus"`
		CacheStatus   string `json:"cachestatus"`
	}

	inputData := GetCommandOutput(v.execPath, "ctrl", fmt.Sprintf("slot=%s", controllerID), "show", "status")
	status := GetRegexpSubmatch(inputData, "Controller Status *: (.*)")
	model := GetRegexpSubmatch(inputData, "(.*) in Slot")
	batteryStatus := GetRegexpSubmatch(inputData, "Battery/Capacitor Status *: (.*)")
	cacheStatus := GetRegexpSubmatch(inputData, "Cache Status *: (.*)")

	data := ReturnData{
		Status:        TrimSpacesLeftAndRight(status),
		Model:         TrimSpacesLeftAndRight(model),
		BatteryStatus: TrimSpacesLeftAndRight(batteryStatus),
		CacheStatus:   TrimSpacesLeftAndRight(cacheStatus),
	}

	return append(MarshallJSON(data, indent), "\n"...)
}

// GetLDStatus - get logical drive status
func (v HPVendor) GetLDStatus(controllerID string, deviceID string, indent int) []byte {
	type ReturnData struct {
		Status string `json:"status"`
		Size   string `json:"size"`
	}

	inputData := GetCommandOutput(v.execPath, "ctrl", fmt.Sprintf("slot=%s", controllerID), "ld", deviceID, "show", "detail")
	status := GetRegexpSubmatch(inputData, "Status *: (.*)")
	size := GetRegexpSubmatch(inputData, "Size *: (.*)")

	data := ReturnData{
		Status: TrimSpacesLeftAndRight(status),
		Size:   TrimSpacesLeftAndRight(size),
	}

	return append(MarshallJSON(data, indent), "\n"...)
}

// GetPDStatus - get physical drive status
func (v HPVendor) GetPDStatus(controllerID string, deviceID string, indent int) []byte {
	type ReturnData struct {
		Status             string `json:"status"`
		Model              string `json:"model"`
		Size               string `json:"size"`
		CurrentTemperature string `json:"currenttemperature"`
		MaximumTemperature string `json:"maximumtemperature"`
	}

	inputData := GetCommandOutput(v.execPath, "ctrl", fmt.Sprintf("slot=%s", controllerID), "pd", deviceID, "show", "detail")
	status := GetRegexpSubmatch(inputData, "[\\s]{2}Status: (.*)")
	model := GetRegexpSubmatch(inputData, "Model: (.*)")
	size := GetRegexpSubmatch(inputData, "[\\s]{2}Size: (.*)")
	currentTemperature := GetRegexpSubmatch(inputData, "Current Temperature \\(C\\): (.*)")
	maximumTemperature := GetRegexpSubmatch(inputData, "Maximum Temperature \\(C\\): (.*)")

	data := ReturnData{
		Status:             TrimSpacesLeftAndRight(status),
		Model:              TrimSpacesLeftAndRight(model),
		Size:               TrimSpacesLeftAndRight(size),
		CurrentTemperature: TrimSpacesLeftAndRight(currentTemperature),
		MaximumTemperature: TrimSpacesLeftAndRight(maximumTemperature),
	}

	return append(MarshallJSON(data, indent), "\n"...)
}

func NewHPVendor(execPath string) Vendor {
	v := HPVendor{execPath: execPath}
	return v
}
