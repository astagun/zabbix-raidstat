package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type MegacliVendor struct {
	execPath string
}

// GetControllersIDs - get number of controllers in the system
func (v MegacliVendor) GetControllersIDs() []string {
	inputData := GetCommandOutput(v.execPath, "-AdpGetPciInfo", "-aALL")
	return GetRegexpAllSubmatch(inputData, "for Controller (\\d*)")
}

// GetLogicalDrivesIDs - get number of logical drives for controller with ID 'controllerID'
func (v MegacliVendor) GetLogicalDrivesIDs(controllerID string) []string {
	inputData := GetCommandOutput(v.execPath, "-LdInfo", "-Lall", fmt.Sprintf("-a%s", controllerID), "-NoLog")
	return GetRegexpAllSubmatch(inputData, "Virtual Drive: (.*?)[\\s]")
}

// GetPhysicalDrivesIDs - get number of physical drives for controller with ID 'controllerID'
func (v MegacliVendor) GetPhysicalDrivesIDs(controllerID string) []string {
	inputData := GetCommandOutput(v.execPath, "-PDList", fmt.Sprintf("-a%s", controllerID), "-NoLog")

	result := regexp.MustCompile("Enclosure Device ID: (\\d+)\\nSlot Number: (\\d+)").FindAllStringSubmatch(string(inputData), -1)

	if os.Getenv("RAIDSTAT_DEBUG") == "y" {
		fmt.Printf("Regexp is '%s'\n", "Enclosure Device ID: (\\d+)\\nSlot Number: (\\d+)")
		fmt.Printf("Result is '%s'\n", result)
	}

	data := []string{}

	if len(result) > 0 {
		for _, v := range result {
			data = append(data, fmt.Sprintf("%s:%s", v[1], v[2]))
		}
	}

	return data
}

// GetControllerStatus - get controller status
func (v MegacliVendor) GetControllerStatus(controllerID string, indent int) []byte {
	type ReturnData struct {
		Status        string `json:"status"`
		Model         string `json:"model"`
		BatteryStatus string `json:"batterystatus"`
	}

	inputData := GetCommandOutput(v.execPath, "-AdpAllInfo", fmt.Sprintf("-a%s", controllerID), "-NoLog")
	model := GetRegexpSubmatch(inputData, "roduct Name[\\s]+: (.*)")

	healthStatuses := []string{}
	for _, v := range []string{
		"Degraded",
		"Offline",
		"Critical Disks",
		"Failed Disks",
	} {
		s := GetRegexpSubmatch(inputData, fmt.Sprintf("%s[\\s]+: (.*)", v))

		if TrimSpacesLeftAndRight(s) != "0" {
			healthStatuses = append(healthStatuses, fmt.Sprintf("%s is %s", v, TrimSpacesLeftAndRight(s)))
		}
	}

	var status string
	if len(healthStatuses) == 0 {
		status = "OK"
	} else {
		status = strings.Join(healthStatuses, ", ")
	}

	inputData = GetCommandOutput(v.execPath, "-AdpBbuCmd", "-GetBbuStatus", fmt.Sprintf("-a%s", controllerID), "-NoLog")
	batteryStatus := GetRegexpSubmatch(inputData, "Battery State: (.*)")

	data := ReturnData{
		Status:        TrimSpacesLeftAndRight(status),
		Model:         TrimSpacesLeftAndRight(model),
		BatteryStatus: TrimSpacesLeftAndRight(batteryStatus),
	}

	return append(MarshallJSON(data, indent), "\n"...)
}

// GetLDStatus - get logical drive status
func (v MegacliVendor) GetLDStatus(controllerID string, deviceID string, indent int) []byte {
	type ReturnData struct {
		Status string `json:"status"`
		Size   string `json:"size"`
	}

	inputData := GetCommandOutput(v.execPath, "-LdInfo", fmt.Sprintf("-L%s", deviceID), fmt.Sprintf("-a%s", controllerID), "-NoLog")
	status := GetRegexpSubmatch(inputData, "State *: (.*)")
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
func (v MegacliVendor) GetPDStatus(controllerID string, deviceID string, indent int) []byte {
	type ReturnData struct {
		Status             string `json:"status"`
		Model              string `json:"model"`
		Size               string `json:"size"`
		CurrentTemperature string `json:"currenttemperature"`
		Smart              string `json:"smart"`
	}

	inputData := GetCommandOutput(v.execPath, "-pdInfo", fmt.Sprintf("-PhysDrv[%s]", deviceID), fmt.Sprintf("-a%s", controllerID), "-NoLog")
	status := TrimSpacesLeftAndRight(GetRegexpSubmatch(inputData, "Firmware state: (.*)"))
	model := GetRegexpSubmatch(inputData, "Inquiry Data: (.*)")
	size := GetRegexpSubmatch(inputData, "Raw Size: (.*) \\[")
	currentTemperature := GetRegexpSubmatch(inputData, "Drive Temperature :(\\d+)C")
	smart := TrimSpacesLeftAndRight(GetRegexpSubmatch(inputData, "Drive has flagged a S.M.A.R.T alert : (.*)"))

	if status == "Online, Spun Up" {
		status = "OK"
	}

	if smart == "No" {
		smart = "OK"
	}

	data := ReturnData{
		Status:             status,
		Model:              TrimSpacesLeftAndRight(model),
		Size:               TrimSpacesLeftAndRight(size),
		CurrentTemperature: TrimSpacesLeftAndRight(currentTemperature),
		Smart:              smart,
	}

	return append(MarshallJSON(data, indent), "\n"...)
}

func NewMegacliVendor(execPath string) Vendor {
	v := MegacliVendor{execPath: execPath}
	return v
}
