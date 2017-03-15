package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

const (
	pythonexec = "python2"
	sch_script = "/bin/ci-scripts/export_schematic.py"
	bom_script = "/bin/ci-scripts/export_bom.py"
	grb_script = "/bin/ci-scripts/export_grb.py"
)

type (
	// Client defines the client data to be embedded in some documents
	Client struct {
		Code string // Enterprise client code
		Name string // Enterprise client name
	}

	// Projects defines the KiCad projects
	Projects struct {
		Codes []string // Enterprise project code
		Names []string // Enterprise project name
	}

	// Gerber defines the options for exporting Gerber files
	GerberLayers struct {
		All      bool `json:"all"`
		Protel   bool `json:"protel"`
		Fcu      bool `json:"fcu"`
		Bcu      bool `json:"bcu"`
		Fmask    bool `json:"fmask"`
		Bmask    bool `json:"bmask"`
		Fsilks   bool `json:"fsilks"`
		Bsilks   bool `json:"bsilks"`
		Edgecuts bool `json:"edgecuts"`
		Drl      bool `json:"drl"`
		Splitth  bool `json:"splitth"`
	}

	// Options defines what to generate
	Options struct {
		Sch bool // Generate Schematic (pdf)
		Bom bool // Generate BOM (xml & xlsx)
		//Brd	bool // Generate PCB plot (pdf)
		Grb    GerberLayers // Gerber file layers
		GrbGen bool         // Generate Gerber files
		//Lyr	bool // Generate plot for each layer (pdf)
		//Wrl	bool // Generate VRML PCB
		//Stp	bool // Generate Step PCB
		//3d	bool // Generate plot of 3D view (png)
	}

	// Plugin defines the KiCad plugin parameters
	Plugin struct {
		Client   Client   // Client configuration
		Projects Projects // Projects configuration
		Options  Options  // Plugin options
	}
)

func (p Plugin) Exec() error {

	var cmds []*exec.Cmd

	if p.Options.Sch {
		for _, pjtname := range p.Projects.Names {
			cmds = append(cmds, commandSchematic(pjtname))
		}
	}
	if p.Options.Bom {
		for _, pjtname := range p.Projects.Names {
			cmds = append(cmds, commandBOM(pjtname))
		}
	}
	if p.Options.GrbGen {
		for _, pjtname := range p.Projects.Names {
			cmds = append(cmds, commandGerber(pjtname, p.Options.Grb))
		}
	}

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

func commandGerber(pjtname string, lyr GerberLayers) *exec.Cmd {

	var options []string

	options = append(options, "-u", grb_script)
	options = append(options, "--brd", pjtname)

	var dir []string
	dir = append(dir, "CI-BUILD/", path.Base(pjtname), "/GBR")
	options = append(options, "--dir", strings.Join(dir, ""))

	if lyr.Splitth {
		options = append(options, "--splitth")
	}
	if lyr.Protel {
		options = append(options, "--protel")
	}
	if lyr.All {
		options = append(options, "--all")
		return exec.Command(
			pythonexec,
			options...,
		)
	}
	if lyr.Fcu {
		options = append(options, "--fcu")
	}
	if lyr.Bcu {
		options = append(options, "--bcu")
	}
	if lyr.Fsilks {
		options = append(options, "--fsilks")
	}
	if lyr.Bsilks {
		options = append(options, "--bsilks")
	}
	if lyr.Fmask {
		options = append(options, "--fmask")
	}
	if lyr.Bmask {
		options = append(options, "--bmask")
	}
	if lyr.Edgecuts {
		options = append(options, "--gko")
	}
	if lyr.Drl {
		options = append(options, "--drl")
	}

	return exec.Command(
		pythonexec,
		options...,
	)
}

func commandSchematic(pjtname string) *exec.Cmd {

	var c = exec.Command(
		pythonexec,
		"-u",
		sch_script,
		pjtname,
	)

	c.Env = os.Environ()
	c.Env = append(c.Env, "DEBIAN_FRONTEND=noninteractive")
	c.Env = append(c.Env, "DISPLAY=:0")

	return c
}

func commandBOM(pjtname string) *exec.Cmd {

	var c = exec.Command(
		pythonexec,
		"-u",
		bom_script,
		pjtname,
	)

	c.Env = os.Environ()
	c.Env = append(c.Env, "DEBIAN_FRONTEND=noninteractive")
	c.Env = append(c.Env, "DISPLAY=:0")

	return c
}

// trace writes each command to stdout with the command wrapped in an xml
// tag so that it can be extracted and displayed in the logs.
func trace(cmd *exec.Cmd) {
	fmt.Fprintf(os.Stdout, "+ %s\n", strings.Join(cmd.Args, " "))
}
