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
	"time"
)

const (
	pythonexec = "python2"
	sch_script = "/bin/ci-scripts/export_schematic.py"
	bom_script = "/bin/ci-scripts/export_bom.py"
	grb_script = "/bin/ci-scripts/export_grb.py"
	tag_script = "/bin/ci-scripts/tag_board.py"
	svg_script = "/bin/PcbDraw/pcbdraw.py"
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

	// GerberLayers defines the options for exporting Gerber files
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

	// Tags defines wich tags to add to the board
	Tags struct {
		All    bool `json:"all"`
		Tag    bool `json:"tag"`
		Commit bool `json:"commit"`
		Date   bool `json:"date"`
		Sed    bool `json:"sed"`
	}

	// Options defines what to generate
	Options struct {
		Sch        bool         // Generate Schematic (pdf)
		Bom        bool         // Generate BOM (xml & xlsx)
		Grb        GerberLayers // Gerber layers enabled
		GrbGen     bool         // Generate Gerber files
		SvgLibDirs []string     // SVG lib folder to pass to the svg generator
		Svg        bool         // Generate SVG output
		Tags       Tags         // Tags enabled
		Tag        bool         // Tag board
		//Brd	bool // Generate PCB plot (pdf)
		//Lyr	bool // Generate plot for each layer (pdf)
		//3d	bool // Generate plot of 3D view (png)
	}

	// Dependencies defines project dependencies to be cloned
	Dependencies struct {
		Libraries  []string // External libraries
		Footprints []string // External footprints
		Modules3d  []string // External 3D models
		Basedir    string   // Base directory
		Templates  []string // External templates
		SvgLibs    []string // External SVG models
	}

	// Commit handles commit information
	Commit struct {
		Tag string // tag if tag event
		Sha string // commit sha
	}

	// Plugin defines the KiCad plugin parameters
	Plugin struct {
		Client       Client       // Client configuration
		Projects     Projects     // Projects configuration
		Options      Options      // Plugin options
		Dependencies Dependencies // Projects dependencies
		Netrc        Netrc        // Authentication
		Commit       Commit       // Commit information
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

	if p.Options.Tag {
		for _, pjtname := range p.Projects.Names {
			if p.Options.Tags.Sed {
				brd := strings.Join([]string{pjtname, ".kicad_pcb"}, "")
				cmds = append(cmds, commandSed("\\$commit\\$", p.Commit.Sha[0:8], brd))
				if len(p.Commit.Tag) > 0 {
					cmds = append(cmds, commandSed("\\$tag\\$", p.Commit.Tag, brd))
				} else {
					cmds = append(cmds, commandSed("\\$tag\\$", "\"\"", brd))
				}
				year, month, day := time.Now().Date()
				date := fmt.Sprintf("%d/%d/%d", day, month, year)
				cmds = append(cmds, commandSed("\\$date\\$", date, brd))
			} else {
				cmds = append(cmds, commandTag(p.Commit, pjtname, p.Options.Tags))
			}
		}
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

	var output []string
	output = append(output, "CI-BUILD/", path.Base(pjtname), "/SVG/", path.Base(pjtname), ".svg")

	err := os.MkdirAll(path.Dir(strings.Join(output, "")), 0777)
	if err != nil {
		fmt.Println("Directory couldn't be created!")
	}

	var board []string
	board = append(board, pjtname, ".kicad_pcb")

	return exec.Command(
		pythonexec,
		"-u",
		svg_script,
		strings.Join(svg_lib_dirs, ","),
		strings.Join(board, ""),
		strings.Join(output, ""),
	)
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

func commandSed(regex string, repl string, file string) *exec.Cmd {

	var reg_repl []string
	reg_repl = append(reg_repl, "s|", regex, "|", repl, "|")

	return exec.Command(
		"sed",
		"-i",
		"-e",
		strings.Join(reg_repl, ""),
		file,
	)
}

func commandTag(c Commit, pjtname string, tags Tags) *exec.Cmd {

	var options []string
	var sha string = c.Sha[0:8]

	options = append(options, "-u", tag_script)
	options = append(options, "--brd", pjtname)

	if tags.All {
		options = append(options, "--tag-date",
			"--tag-commit", "--commit", sha)
		if c.Tag != "" {
			options = append(options,
				"--tag-tag", "--tag", c.Tag)
		}
		return exec.Command(
			pythonexec,
			options...,
		)
	}

	if tags.Date {
		options = append(options, "--tag-date")
	}
	if tags.Commit {
		options = append(options, "--tag-commit", "--commit", sha)
	}
	if tags.Tag && c.Tag != "" {
		options = append(options, "--tag-tag", "--tag", c.Tag)
	}

	return exec.Command(
		pythonexec,
		options...,
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
