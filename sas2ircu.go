package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type SAS2IrcuVendor struct {
	execPath string
}

// GetControllersIDs - get number of controllers in the system
func (v SAS2IrcuVendor) GetControllersIDs() []string {
	inputData := GetCommandOutput(v.execPath, "list")
	return GetRegexpAllSubmatch(inputData, "\\s+(\\d+)\\s+.*")
}

// GetLogicalDrivesIDs - get number of logical drives for controller with ID 'controllerID'
func (v SAS2IrcuVendor) GetLogicalDrivesIDs(controllerID string) []string {
	inputData := GetCommandOutput(v.execPath, controllerID, "display")
	return GetRegexpAllSubmatch(inputData, "IR volume (\\d+)")
}

// GetPhysicalDrivesIDs - get number of physical drives for controller with ID 'controllerID'
func (v SAS2IrcuVendor) GetPhysicalDrivesIDs(controllerID string) []string {
	inputData := GetCommandOutput(v.execPath, controllerID, "display")
	sliceArr := GetArraySliceByte(inputData, "Device is a Hard disk", "Drive Type")
	data := []string{}

	if len(sliceArr) > 0 {
		for _, v := range sliceArr {
			enclosure := GetRegexpSubmatch([]byte(v), "Enclosure # *: (.*)")
			slot := GetRegexpSubmatch([]byte(v), "Slot # *: (.*)")

			if len(enclosure) > 0 && len(slot) > 0 {
				data = append(data, fmt.Sprintf("%s:%s", enclosure, slot))
			}
		}
	}

	return data
}

// GetControllerStatus - get controller status
func (v SAS2IrcuVendor) GetControllerStatus(controllerID string, indent int) []byte {
	type ReturnData struct {
		Status string `json:"status"`
		Model  string `json:"model"`
		// Temperature string `json:"temperature"`
	}

	inputData := GetCommandOutput(v.execPath, controllerID, "display")
	model := GetRegexpSubmatch(inputData, "Controller type *: (.*)")

	result := regexp.MustCompile("Status of volume\\s+: .*\\((.*)\\)").FindAllStringSubmatch(string(inputData), -1)

	if os.Getenv("RAIDSTAT_DEBUG") == "y" {
		fmt.Printf("Regexp is '%s'\n", "Status of volume\\s+: .*\\((.*)\\)")
		fmt.Printf("Result is '%s'\n", result)
	}

	healthStatuses := []string{}

	if len(result) > 0 {
		for _, v := range result {
			if v[1] != "OKY" {
				healthStatuses = append(healthStatuses, fmt.Sprintf("%s", v))
			}
		}
	}

	var status string
	if len(healthStatuses) == 0 {
		status = "OK"
	} else {
		status = strings.Join(healthStatuses, ", ")
	}

	data := ReturnData{
		Status: TrimSpacesLeftAndRight(status),
		Model:  TrimSpacesLeftAndRight(model),
	}

	return append(MarshallJSON(data, indent), "\n"...)
}

// GetLDStatus - get logical drive status
func (v SAS2IrcuVendor) GetLDStatus(controllerID string, deviceID string, indent int) []byte {
	type ReturnData struct {
		Status string `json:"status"`
		Size   string `json:"size"`
	}

	inputData := GetCommandOutput(v.execPath, controllerID, "display")
	sliceData := GetSliceByte(inputData, "IR volume "+deviceID, "Physical")

	status := GetRegexpSubmatch(sliceData, "Status of volume *: (.*)")
	size := GetRegexpSubmatch(sliceData, "Size \\(in MB\\) *: (.*)")

	if status == "Okay (OKY)" {
		status = "OK"
	}

	data := ReturnData{
		Status: TrimSpacesLeftAndRight(status),
		Size:   TrimSpacesLeftAndRight(size),
	}

	return append(MarshallJSON(data, indent), "\n"...)
}

// GetPDStatus - get physical drive status
func (v SAS2IrcuVendor) GetPDStatus(controllerID string, deviceID string, indent int) []byte {
	type ReturnData struct {
		Status    string `json:"status"`
		Model     string `json:"model"`
		TotalSize string `json:"totalsize"`
	}

	deviceData := strings.Split(deviceID, ":")
	if len(deviceData) < 2 {
		fmt.Printf("Error - wrong device id '%s'.\n", deviceID)
		os.Exit(1)
	}

	inputData := GetCommandOutput(v.execPath, controllerID, "display")
	sliceArr := GetArraySliceByte(inputData, "Device is a Hard disk", "Drive Type")

	if len(sliceArr) > 0 {
		for _, v := range sliceArr {
			enclosure := GetRegexpSubmatch([]byte(v), "Enclosure # *: (.*)")
			slot := GetRegexpSubmatch([]byte(v), "Slot # *: (.*)")

			if enclosure == deviceData[0] && slot == deviceData[1] {
				status := GetRegexpSubmatch([]byte(v), "[\\s]{2}State *: (.*)")
				model := GetRegexpSubmatch([]byte(v), "Model Number *: (.*)")
				totalSize := GetRegexpSubmatch([]byte(v), "Size \\(in MB\\)/\\(in sectors\\) *: (\\d+)/\\d+")

				if status == "Optimal (OPT)" {
					status = "OK"
				}

				data := ReturnData{
					Status:    TrimSpacesLeftAndRight(status),
					Model:     TrimSpacesLeftAndRight(model),
					TotalSize: TrimSpacesLeftAndRight(totalSize),
				}

				return append(MarshallJSON(data, indent), "\n"...)
			}
		}
	}

	return []byte("")
}

func GetSliceByte(buf []byte, start string, end string) []byte {
	lines := strings.Split(string(buf), "\n")
	capture := false
	var sliceData []byte

	if len(lines) > 0 {
		for _, v := range lines {
			if strings.Contains(v, start) {
				capture = true
			}

			if capture {
				sliceData = append(sliceData, v+"\n"...)

				if strings.Contains(v, end) {
					break
				}
			}
		}
	}

	return sliceData
}

func GetArraySliceByte(buf []byte, start string, end string) (data []string) {
	lines := strings.Split(string(buf), "\n")
	capture := false
	var sliceData []byte

	if len(lines) > 0 {
		for _, v := range lines {
			if strings.Contains(v, start) {
				if capture {
					data = append(data, string(sliceData))
					sliceData = nil
				} else {
					capture = true
				}
			}

			if strings.Contains(v, end) {
				if capture {
					data = append(data, string(sliceData))
					sliceData = nil
				}

				capture = false
			}

			if capture {
				sliceData = append(sliceData, v+"\n"...)
			}
		}
	}

	return
}

func NewSAS2IrcuVendor(execPath string) Vendor {
	v := SAS2IrcuVendor{execPath: execPath}
	return v
}
