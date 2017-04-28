package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

const (
	pythonexec    = "python2"
	sch_script    = "/bin/ci-scripts/export_schematic.py"
	bom_script    = "/bin/ci-scripts/export_bom.py"
	grb_script    = "/bin/ci-scripts/export_grb.py"
	svg_script    = "/bin/PcbDraw/pcbdraw.py"
	style_dir     = "/bin/PcbDraw/styles"
	default_style = "gatema-green"
)

const (
	DEP_TYPE_LIB      = iota
	DEP_TYPE_PRETTY   = iota
	DEP_TYPE_3D       = iota
	DEP_TYPE_TEMPLATE = iota
	DEP_TYPE_SVG      = iota
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

	Netrc struct {
		Machine  string
		Login    string
		Password string
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
		Sch        bool         // Generate Schematic (pdf)
		Bom        bool         // Generate BOM (xml & xlsx)
		Grb        GerberLayers // Gerber file layers
		GrbGen     bool         // Generate Gerber files
		SvgLibDirs []string     // SVG lib folder to pass to the svg generator
		Svg        bool         // Generate SVG output
		//Brd	bool // Generate PCB plot (pdf)
		//Lyr	bool // Generate plot for each layer (pdf)
		//Wrl	bool // Generate VRML PCB
		//Stp	bool // Generate Step PCB
		//3d	bool // Generate plot of 3D view (png)
	}

	Dependencies struct {
		Libraries  []string // External libraries
		Footprints []string // External footprints
		Modules3d  []string // External 3D models
		Basedir    string   // Base directory
		Templates  []string // External templates
		SvgLibs    []string // External SVG models
	}

	// Plugin defines the KiCad plugin parameters
	Plugin struct {
		Client       Client       // Client configuration
		Projects     Projects     // Projects configuration
		Options      Options      // Plugin options
		Dependencies Dependencies // Projects dependencies
		Netrc        Netrc        // Authentication
	}
)

func (p Plugin) Exec() error {

	err := writeNetrc(p.Netrc.Machine, p.Netrc.Login, p.Netrc.Password)
	if err != nil {
		return err
	}

	var cmds []*exec.Cmd

	if p.Dependencies.Basedir == "" {
		p.Dependencies.Basedir = "/usr/share/kicad"
	}

	for _, dep := range p.Dependencies.Libraries {
		cmds = append(cmds, commandClone(dep, DEP_TYPE_LIB, p.Dependencies.Basedir))
	}

	for _, dep := range p.Dependencies.Footprints {
		cmds = append(cmds, commandClone(dep, DEP_TYPE_PRETTY, p.Dependencies.Basedir))
	}

	for _, dep := range p.Dependencies.Modules3d {
		cmds = append(cmds, commandClone(dep, DEP_TYPE_3D, p.Dependencies.Basedir))
	}

	for _, dep := range p.Dependencies.Templates {
		cmds = append(cmds, commandClone(dep, DEP_TYPE_TEMPLATE, p.Dependencies.Basedir))
	}

	for _, dep := range p.Dependencies.SvgLibs {
		cmds = append(cmds, commandClone(dep, DEP_TYPE_SVG, p.Dependencies.Basedir))
	}

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

	var svg_lib_dirs []string
	if len(p.Options.SvgLibDirs) > 0 {
		for _, lib := range p.Options.SvgLibDirs {
			svg_lib_dirs = append(svg_lib_dirs, path.Join(p.Dependencies.Basedir, "svg-lib", lib))
		}
	}

	if p.Options.Svg {
		for _, pjtname := range p.Projects.Names {
			cmds = append(cmds, commandSVG(pjtname, svg_lib_dirs))
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

func commandSVG(pjtname string, svg_lib_dirs []string) *exec.Cmd {

	var style []string
	style = append(style, default_style, ".json")

	var output []string
	output = append(output, "CI-BUILD/", path.Base(pjtname), "/SVG/", path.Base(pjtname), ".svg")

	err := os.MkdirAll(path.Dir(strings.Join(output, "")), 0777)
	if err != nil {
		fmt.Println("Directory couldn't be created!")
	}

	var board []string
	board = append(board, pjtname, ".kicad_pcb")

	var c = exec.Command(
		pythonexec,
		"-u",
		svg_script,
		path.Join(style_dir, strings.Join(style, "")),
		strings.Join(svg_lib_dirs, ","),
		strings.Join(output, ""),
		strings.Join(board, ""),
	)

	return c
}

func commandClone(depurl string, deptype int, basedir string) *exec.Cmd {

	if deptype == DEP_TYPE_LIB {
		basedir = path.Join(basedir, "library")
	} else if deptype == DEP_TYPE_PRETTY {
		basedir = path.Join(basedir, "footprints")
	} else if deptype == DEP_TYPE_3D {
		basedir = path.Join(basedir, "modules/packages3d")
	} else if deptype == DEP_TYPE_TEMPLATE {
		basedir = path.Join(basedir, "template")
	} else if deptype == DEP_TYPE_SVG {
		basedir = path.Join(basedir, "svg-lib")
	}

	err := os.MkdirAll(basedir, 0777)
	if err != nil {
		fmt.Println("Directory couldn't be created!")
	}

	var cmd []string
	cmd = append(cmd, "cd", basedir, "&&", "git", "clone", depurl)

	return exec.Command(
		"/bin/sh",
		"-c",
		strings.Join(cmd, " "),
	)
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

// helper function to write a netrc file. [From drone-git]
func writeNetrc(machine, login, password string) error {
	if machine == "" {
		return nil
	}
	out := fmt.Sprintf(
		netrcFile,
		machine,
		login,
		password,
	)

	home := "/root"
	u, err := user.Current()
	if err == nil {
		home = u.HomeDir
	}
	path := filepath.Join(home, ".netrc")
	return ioutil.WriteFile(path, []byte(out), 0600)
}

const netrcFile = `
machine %s
login %s
password %s
`
