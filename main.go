package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ps78674/docopt.go"
)

const configFile = "config.json"

var (
	indent       int
	toolVendor   string
	toolBinary   string
	operation    string
	argOption    string
	controllerID string
	deviceID     string
)

var vendors = []string{"adaptec", "megacli", "hp", "marvell", "sas2ircu"}

func init() {
	var (
		discoveryOption string
		statusOption    string
		options         []string
	)

	discoveryOptions := []string{"ct", "ld", "pd"}
	statusOptions := []string{"ct,<CONTROLLER_ID>", "ld,<CONTROLLER_ID>,<LD_ID>", "pd,<CONTROLLER_ID>,<PD_ID>"}

	var programName = filepath.Base(os.Args[0])
	var usage = fmt.Sprintf(`%[1]s: parse raid vendor tool output and format it as json

Usage:
  %[1]s (-v <VENDOR>) (-d <OPTION> | -s <OPTION>) [-i <INT>]

Options:
  -v, --vendor <VENDOR>    raid tool vendor, one of: %[2]s
  -d, --discover <OPTION>  discovery option, one of: %[3]s
  -s, --status <OPTION>    status option, one of: %[4]s
  -i, --indent <INT>       indent json output level [default: 0]

  -h, --help               show this screen
	`, programName, strings.Join(vendors, " | "), strings.Join(discoveryOptions, " | "), strings.Join(statusOptions, " | "))

	cmdOpts, err := docopt.ParseDoc(usage)
	if err != nil {
		fmt.Printf("error parsing options: %s\n", err)
		os.Exit(1)
	}

	toolVendor, _ = cmdOpts.String("--vendor")
	discoveryOption, _ = cmdOpts.String("--discover")
	statusOption, _ = cmdOpts.String("--status")
	indent, _ = cmdOpts.Int("--indent")

	for i, v := range vendors {
		if v != toolVendor {
			if i == len(vendors)-1 {
				fmt.Printf("Vendors must be one of '%s' (ex.: -v adaptec), got '%s'.\n", strings.Join(vendors, " | "), toolVendor)
				docopt.PrintHelpOnly(nil, usage)
				os.Exit(1)
			}
			continue
		}
		break
	}

	if len(discoveryOption) != 0 {
		operation = "Discovery"
		options = discoveryOptions
		argOption = discoveryOption
	} else if len(statusOption) != 0 {
		operation = "Status"
		options = statusOptions
		argOption = statusOption
	}

	for i, v := range options {
		rangeValues := strings.Split(v, ",")
		argOptionValues := strings.SplitN(argOption, ",", 3)

		if argOptionValues[0] != rangeValues[0] || len(argOptionValues) != len(rangeValues) {
			if i == len(options)-1 {
				fmt.Printf("%s option must be one of '%s', got '%s'.\n", operation, strings.Join(options, " | "), argOption)
				docopt.PrintHelpOnly(nil, usage)
				os.Exit(1)
			}

			continue
		}

		if len(argOptionValues) > 1 {
			argOption = argOptionValues[0]
		}

		if len(argOptionValues) == 2 || len(argOptionValues) == 3 {
			controllerID = argOptionValues[1]
		}

		if len(argOptionValues) == 3 {
			controllerID = argOptionValues[1]
			deviceID = argOptionValues[2]
		}

		break
	}
}

func discoverControllers(v Vendor) {
	type (
		Element struct {
			CT string `json:"{#CT_ID}"`
		}
		Reply struct {
			Data []Element `json:"data"`
		}
	)

	var (
		d    []Element
		JSON []byte
		jErr error
	)

	controllersIDs := v.GetControllersIDs()

	for _, v := range controllersIDs {
		d = append(d, Element{CT: v})
	}

	if indent > 0 {
		JSON, jErr = json.MarshalIndent(Reply{d}, "", strings.Repeat(" ", indent))
	} else {
		JSON, jErr = json.Marshal(Reply{d})
	}

	if jErr != nil {
		fmt.Println(jErr)
		os.Exit(1)
	}

	os.Stdout.Write(append(JSON, "\n"...))
}

func discoverLogicalDrives(v Vendor) {
	type (
		Element struct {
			CT string `json:"{#CT_ID}"`
			LD string `json:"{#LD_ID}"`
		}
		Reply struct {
			Data []Element `json:"data"`
		}
	)

	var (
		d    []Element
		JSON []byte
		jErr error
	)

	controllersIDs := v.GetControllersIDs()
	for _, ctID := range controllersIDs {
		logicalDrivesIDs := v.GetLogicalDrivesIDs(ctID)
		for _, ldID := range logicalDrivesIDs {
			d = append(d, Element{CT: ctID, LD: ldID})
		}

	}

	if indent > 0 {
		JSON, jErr = json.MarshalIndent(Reply{d}, "", strings.Repeat(" ", indent))
	} else {
		JSON, jErr = json.Marshal(Reply{d})
	}

	if jErr != nil {
		fmt.Println(jErr)
		os.Exit(1)
	}

	os.Stdout.Write(append(JSON, "\n"...))
}

func discoverPhysicalDrives(v Vendor) {
	type (
		Element struct {
			CT string `json:"{#CT_ID}"`
			PD string `json:"{#PD_ID}"`
		}
		Reply struct {
			Data []Element `json:"data"`
		}
	)

	var (
		d    []Element
		JSON []byte
		jErr error
	)

	controllersIDs := v.GetControllersIDs()
	for _, ctID := range controllersIDs {
		logicalDrivesIDs := v.GetPhysicalDrivesIDs(ctID)
		for _, pdID := range logicalDrivesIDs {
			d = append(d, Element{CT: ctID, PD: pdID})
		}

	}

	if indent > 0 {
		JSON, jErr = json.MarshalIndent(Reply{d}, "", strings.Repeat(" ", indent))
	} else {
		JSON, jErr = json.Marshal(Reply{d})
	}

	if jErr != nil {
		fmt.Println(jErr)
		os.Exit(1)
	}

	os.Stdout.Write(append(JSON, "\n"...))
}

func getControllerStatus(v Vendor, controllerID string) {
	os.Stdout.Write(v.GetControllerStatus(controllerID, indent))
}

func getLDStatus(v Vendor, controllerID string, deviceID string) {
	os.Stdout.Write(v.GetLDStatus(controllerID, deviceID, indent))
}

func getPDStatus(v Vendor, controllerID string, deviceID string) {
	os.Stdout.Write(v.GetPDStatus(controllerID, deviceID, indent))
}

type Vendor interface {
	GetControllersIDs() []string
	GetLogicalDrivesIDs(string) []string
	GetPhysicalDrivesIDs(string) []string
	GetControllerStatus(string, int) []byte
	GetLDStatus(string, string, int) []byte
	GetPDStatus(string, string, int) []byte
}

func main() {
	var v Vendor
	switch toolVendor {
	case "adaptec":
		v = NewAdaptecVendor("arcconf")
	case "megacli":
		v = NewMegacliVendor("megacli")
	case "hp":
		v = NewHPVendor("ssacli")
	case "marvell":
		v = NewMarvellVendor("mvcli")
	case "sas2ircu":
		v = NewSAS2IrcuVendor("sas2ircu")
	default:
		fmt.Printf("unknown vendor %q", toolVendor)
		os.Exit(1)
	}

	switch argOption {
	case "ct":
		switch operation {
		case "Discovery":
			discoverControllers(v)
		case "Status":
			getControllerStatus(v, controllerID)
		}
	case "ld":
		switch operation {
		case "Discovery":
			discoverLogicalDrives(v)
		case "Status":
			getLDStatus(v, controllerID, deviceID)
		}
	case "pd":
		switch operation {
		case "Discovery":
			discoverPhysicalDrives(v)
		case "Status":
			getPDStatus(v, controllerID, deviceID)
		}
	}
}
