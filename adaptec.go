package main

import (
	"fmt"
	"os"
	"strings"
)

type AdaptecVendor struct {
	execPath string
}

// GetControllersIDs - get number of controllers in the system
func (v AdaptecVendor) GetControllersIDs() []string {
	inputData := GetCommandOutput(v.execPath, "list")
	return GetRegexpAllSubmatch(inputData, "Controller ([^a-zA-Z].*?):")
}

// GetLogicalDrivesIDs - get number of logical drives for controller with ID 'controllerID'
func (v AdaptecVendor) GetLogicalDrivesIDs(controllerID string) []string {
	inputData := GetCommandOutput(v.execPath, "getconfig", controllerID, "ld")
	return GetRegexpAllSubmatch(inputData, "Logical Device number (.*)[\\s]")
}

// GetPhysicalDrivesIDs - get number of physical drives for controller with ID 'controllerID'
func (v AdaptecVendor) GetPhysicalDrivesIDs(controllerID string) []string {
	inputData := GetCommandOutput(v.execPath, "getconfig", controllerID, "pd")
	return GetRegexpAllSubmatch(inputData, "Device is a Hard drive[\\s\\S]*?Reported Channel,Device\\(T:L\\)[\\s]*[:][\\s](.*?)\\(.*\\)[\\s]")
}

// GetControllerStatus - get controller status
func (v AdaptecVendor) GetControllerStatus(controllerID string, indent int) []byte {
	type ReturnData struct {
		Status      string `json:"status"`
		Model       string `json:"model"`
		Temperature string `json:"temperature"`
	}

	inputData := GetCommandOutput(v.execPath, "getconfig", controllerID, "ad")
	status := GetRegexpSubmatch(inputData, "Controller Status *: (.*)")
	model := GetRegexpSubmatch(inputData, "Controller Model *: (.*)")
	temperature := GetRegexpSubmatch(inputData, "Temperature *: (.*) C")

	if status == "Optimal" {
		status = "OK"
	}

	data := ReturnData{
		Status:      TrimSpacesLeftAndRight(status),
		Model:       TrimSpacesLeftAndRight(model),
		Temperature: TrimSpacesLeftAndRight(temperature),
	}

	return append(MarshallJSON(data, indent), "\n"...)
}

// GetLDStatus - get logical drive status
func (v AdaptecVendor) GetLDStatus(controllerID string, deviceID string, indent int) []byte {
	type ReturnData struct {
		Status string `json:"status"`
		Size   string `json:"size"`
	}

	inputData := GetCommandOutput(v.execPath, "getconfig", controllerID, "ld", deviceID)
	status := GetRegexpSubmatch(inputData, "Status of Logical Device *: (.*)")
	size := GetRegexpSubmatch(inputData, "Size *: (.*)")

	if status == "Optimal" {
		status = "OK"
	}

	data := ReturnData{
		Status: TrimSpacesLeftAndRight(status),
		Size:   TrimSpacesLeftAndRight(size),
	}

	return append(MarshallJSON(data, indent), "\n"...)
}

// GetPDStatus - get physical drive status
func (v AdaptecVendor) GetPDStatus(controllerID string, deviceID string, indent int) []byte {
	type ReturnData struct {
		Status      string `json:"status"`
		Model       string `json:"model"`
		Smart       string `json:"smart"`
		SmartWarn   string `json:"smartwarnings"`
		TotalSize   string `json:"totalsize"`
		Temperature string `json:"temperature"`
	}

	deviceData := strings.Split(deviceID, ",")
	if len(deviceData) < 2 {
		fmt.Printf("Error - wrong device id '%s'.\n", deviceID)
		os.Exit(1)

	}

	inputData := GetCommandOutput(v.execPath, "getconfig", controllerID, "pd", deviceData[0], deviceData[1])
	status := GetRegexpSubmatch(inputData, "[\\s]{2}State *: (.*)")
	model := GetRegexpSubmatch(inputData, "Model *: (.*)")
	smart := GetRegexpSubmatch(inputData, "S.M.A.R.T. *: (.*)")
	smartWarn := GetRegexpSubmatch(inputData, "S.M.A.R.T. warnings *: (.*)")
	totalSize := GetRegexpSubmatch(inputData, "Total Size *: (.*)")
	temperature := GetRegexpSubmatch(inputData, "Temperature *: (.*) C")

	if status == "Online" {
		status = "OK"
	}

	if smart == "No" {
		smart = "OK"
	}

	data := ReturnData{
		Status:      TrimSpacesLeftAndRight(status),
		Model:       TrimSpacesLeftAndRight(model),
		Smart:       TrimSpacesLeftAndRight(smart),
		SmartWarn:   TrimSpacesLeftAndRight(smartWarn),
		TotalSize:   TrimSpacesLeftAndRight(totalSize),
		Temperature: TrimSpacesLeftAndRight(temperature),
	}

	return append(MarshallJSON(data, indent), "\n"...)
}

func NewAdaptecVendor(execPath string) Vendor {
	v := AdaptecVendor{execPath: execPath}
	return v
}
