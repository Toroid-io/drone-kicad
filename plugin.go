package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	pythonexec = "python2"
	sch_script = "/bin/ci-scripts/export_schematic.py"
	bom_script = "/bin/ci-scripts/export_bom.py"
)

type (
	// Client defines the client data to be embedded in some documents
	Client struct {
		Code string // Enterprise client code
		Name string // Enterprise client name
	}

	// Project defines the KiCad project
	Project struct {
		Code string // Enterprise project code
		Name string // Enterprise project name
	}

	/* Deploy defines where to store the generated files
	Deploy struct {
		Url	string // Deploy folder URL
		User	string // Deploy server username
		Pass	string // Deploy server password
	}
	*/

	// Options defines what to generate
	Options struct {
		Sch bool // Generate Schematic (pdf)
		Bom bool // Generate BOM (xml & xlsx)
		//Brd	bool // Generate PCB plot (pdf)
		//Grb	bool // Generate Gerber files
		//Lyr	bool // Generate plot for each layer (pdf)
		//Wrl	bool // Generate VRML PCB
		//Stp	bool // Generate Step PCB
		//3d	bool // Generate plot of 3D view (png)
	}

	// Plugin defines the KiCad plugin parameters
	Plugin struct {
		Client  Client  // Client configuration
		Project Project // Project configuration
		//Deploy	Deploy	// Deploy configuration
		Options Options // Plugin options
	}
)

func (p Plugin) Exec() error {

	var cmds []*exec.Cmd

	if p.Options.Sch {
		cmds = append(cmds, commandSchematic(p.Project))
	}
	if p.Options.Bom {
		cmds = append(cmds, commandBOM(p.Project))
	}

	// Set env variables
	os.Setenv("DISPLAY", ":0")
	os.Setenv("DEBIAN_FRONTEND", "noninteractive")

	// execute all commands in batch mode.
	for _, cmd := range cmds {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		trace(cmd)

		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func commandSchematic(pjt Project) *exec.Cmd {

	return exec.Command(
		pythonexec,
		"-u",
		sch_script,
		pjt.Name,
	)
}

func commandBOM(pjt Project) *exec.Cmd {

	return exec.Command(
		pythonexec,
		"-u",
		bom_script,
		pjt.Name,
	)
}

// trace writes each command to stdout with the command wrapped in an xml
// tag so that it can be extracted and displayed in the logs.
func trace(cmd *exec.Cmd) {
	fmt.Fprintf(os.Stdout, "+ %s\n", strings.Join(cmd.Args, " "))
}
