package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	pythonexec = "python2"
	sch_script = "/bin/ci-scripts/export_schematic.py"
	bom_script = "/bin/ci-scripts/export_bom.py"
	grb_script = "/bin/ci-scripts/export_grb.py"
	tag_script = "/bin/ci-scripts/tag_board.py"
	dlf_script = "/bin/ci-scripts/delete_footprints.py"
	ftr_script = "/bin/ci-scripts/footprints_to_remove.sh"
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
		Code string `json:"code"` // Enterprise client code
		Name string `json:"name"` // Enterprise client name
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
		All     bool `json:"all"`
		Tag     bool `json:"tag"`
		Commit  bool `json:"commit"`
		Date    bool `json:"date"`
		Sed     bool `json:"sed"`
		Variant bool `json:"variant"`
	}

	// Options for projects
	ProjectOptions struct {
		Sch  bool         // Generate Schematic (pdf)
		Bom  bool         // Generate BOM xml
		Grb  GerberLayers // Gerber layers enabled
		Svg  bool         // Generate SVG output
		Tags Tags         // Tags enabled
		Pcb  bool         // Export PCB file
		Wait int          // Delay before variant generation (allows Pcbnew to fully load)
	}

	// Options for variants
	VariantOptions struct {
		Grb  GerberLayers // Gerber layers enabled
		Svg  bool         // Generate SVG output
		Tags Tags         // Tags enabled
		Pcb  bool         // Export PCB file
		Wait int          // Delay before variant generation (allows Pcbnew to fully load)
		//Brd	bool // Generate PCB plot (pdf)
		//Lyr	bool // Generate plot for each layer (pdf)
		//3d	bool // Generate plot of 3D view (png)
	}

	// Dependencies defines project dependencies to be cloned
	Dependencies struct {
		Libraries  []string `json:"libraries"`  // External libraries
		Footprints []string `json:"footprints"` // External footprints
		Modules3d  []string `json:"modules3d"`  // External 3D models
		Basedir    string   `json:"basedir"`    // Base directory
		Templates  []string `json:"templates"`  // External templates
		Svglibs    []string `json:"svglibs"`    // External SVG models
		Svglibdirs []string `json:"svglibdirs"` // SVG lib folder to pass to the svg generator
	}

	// Commit handles commit information
	Commit struct {
		Tag string // tag if tag event
		Sha string // commit sha
	}

	// Variant defines a varaint in the project
	Variant struct {
		Name    string
		Content string
		Options VariantOptions
	}

	Project struct {
		Code         string         `json:"code"`         // Enterprise project code
		Main         string         `json:"main"`         // Enterprise project name
		Client       Client         `json:"client"`       // Enterprise client code
		Dependencies Dependencies   `json:"dependencies"` // Projects dependencies
		Variants     []Variant      `json:"variants"`     // Project variants
		Options      ProjectOptions `json:"options"`      // Project options
	}

	// Plugin defines the KiCad plugin parameters
	Plugin struct {
		Projects []Project // Projects configuration
		Netrc    Netrc     // Authentication
		Commit   Commit    // Commit information
	}
)

func (p Plugin) Exec() error {

	err := writeNetrc(p.Netrc.Machine, p.Netrc.Login, p.Netrc.Password)
	if err != nil {
		return err
	}

	var cmds []*exec.Cmd

	for _, project := range p.Projects {

		if project.Dependencies.Basedir == "" {
			project.Dependencies.Basedir = "/usr/share/kicad"
		}

		for _, dep := range project.Dependencies.Libraries {
			cmds = append(cmds, commandClone(dep, DEP_TYPE_LIB, project.Dependencies.Basedir))
		}

		for _, dep := range project.Dependencies.Footprints {
			cmds = append(cmds, commandClone(dep, DEP_TYPE_PRETTY, project.Dependencies.Basedir))
		}

		for _, dep := range project.Dependencies.Modules3d {
			cmds = append(cmds, commandClone(dep, DEP_TYPE_3D, project.Dependencies.Basedir))
		}

		for _, dep := range project.Dependencies.Templates {
			cmds = append(cmds, commandClone(dep, DEP_TYPE_TEMPLATE, project.Dependencies.Basedir))
		}

		for _, dep := range project.Dependencies.Svglibs {
			cmds = append(cmds, commandClone(dep, DEP_TYPE_SVG, project.Dependencies.Basedir))
		}

		var svg_lib_dirs []string
		if len(project.Dependencies.Svglibdirs) > 0 {
			for _, lib := range project.Dependencies.Svglibdirs {
				svg_lib_dirs = append(svg_lib_dirs, path.Join(project.Dependencies.Basedir, "svg-lib", lib))
			}
		}

		// Export schematic
		if project.Options.Sch {
			cmds = append(cmds, commandSchematic(project))
		}

		// Export BOM (xml)
		if project.Options.Bom {
			cmds = append(cmds, commandBOM(project))
		}

		// Process each variant
		for _, variant := range project.Variants {

			// Create a variant PCB file for each variant
			cmds = append(cmds, commandVariant(variant, project))

			// Tag board
			if variant.Options.Tags.Sed {
				cmds = append(cmds, commandSed("\\$commit\\$", p.Commit.Sha[0:8], project.Main, variant.Name))
				if len(p.Commit.Tag) > 0 {
					cmds = append(cmds, commandSed("\\$tag\\$", p.Commit.Tag, project.Main, variant.Name))
				} else {
					cmds = append(cmds, commandSed("\\$tag\\$", "\"\"", project.Main, variant.Name))
				}
				year, month, day := time.Now().Date()
				date := fmt.Sprintf("%d/%d/%d", day, month, year)
				cmds = append(cmds, commandSed("\\$date\\$", date, project.Main, variant.Name))
			} else {
				cmds = append(cmds, commandTag(p.Commit, project.Main, variant.Name, variant.Options.Tags))
			}

			// Export PCB
			if variant.Options.Pcb {
				cmds = append(cmds, commandCopyPcb(project.Main, variant.Name))
			}

			// Export SVG
			if variant.Options.Svg {
				cmds = append(cmds, commandSVG(project.Main, variant.Name, svg_lib_dirs))
			}

			// Export Gerbers
			cmds = append(cmds, commandGerber(project.Main, variant.Name, variant.Options.Grb))
		}

		// Tag board
		if project.Options.Tags.Sed {
			cmds = append(cmds, commandSed("\\$commit\\$", p.Commit.Sha[0:8], project.Main, ""))
			if len(p.Commit.Tag) > 0 {
				cmds = append(cmds, commandSed("\\$tag\\$", p.Commit.Tag, project.Main, ""))
			} else {
				cmds = append(cmds, commandSed("\\$tag\\$", "\"\"", project.Main, ""))
			}
			year, month, day := time.Now().Date()
			date := fmt.Sprintf("%d/%d/%d", day, month, year)
			cmds = append(cmds, commandSed("\\$date\\$", date, project.Main, ""))
		} else {
			cmds = append(cmds, commandTag(p.Commit, project.Main, "", project.Options.Tags))
		}

		// Export PCB
		if project.Options.Pcb {
			cmds = append(cmds, commandCopyPcb(project.Main, ""))
		}

		// Export Gerbers
		cmds = append(cmds, commandGerber(project.Main, "", project.Options.Grb))

		// Export SVG
		if project.Options.Svg {
			cmds = append(cmds, commandSVG(project.Main, "", svg_lib_dirs))
		}
	}

	// execute all commands in batch mode.
	for _, cmd := range cmds {
		if cmd != nil {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			trace(cmd)

			err := cmd.Run()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func commandCopyPcb(pjtname string, variant string) *exec.Cmd {

	var board []string
	if len(variant) > 0 {
		board = append(board, pjtname, "_", variant, ".kicad_pcb")
	} else {
		board = append(board, pjtname, ".kicad_pcb")
	}

	var folder []string
	if len(variant) > 0 {
		folder = append(folder, "CI-BUILD/", path.Base(pjtname), "_", variant, "/PCB")
	} else {
		folder = append(folder, "CI-BUILD/", path.Base(pjtname), "/PCB")
	}

	var cmd []string
	cmd = append(cmd, "mkdir", "-p", strings.Join(folder, ""), "&&", "cp", strings.Join(board, ""), strings.Join(folder, ""))

	return exec.Command(
		"/bin/sh",
		"-c",
		strings.Join(cmd, " "),
	)
}

func commandVariant(variant Variant, project Project) *exec.Cmd {

	var schematic []string
	schematic = append(schematic, project.Main, ".sch")

	var board []string
	board = append(board, project.Main, ".kicad_pcb")

	var options []string
	options = append(options, project.Main)
	if len(variant.Content) > 0 {
		options = append(options, strings.Split(variant.Content, ",")...)
	}

	fpToRemove := exec.Command(
		ftr_script,
		options...,
	)
	var stdout bytes.Buffer
	fpToRemove.Stdout = &stdout
	err := fpToRemove.Run()
	if err != nil {
		fmt.Printf("%s", err)
	}
	outStr := string(stdout.Bytes())

	var options2 []string
	options2 = append(options2, "-u")
	options2 = append(options2, dlf_script)
	options2 = append(options2, "--brd")
	options2 = append(options2, project.Main)
	options2 = append(options2, "--footprints")
	options2 = append(options2, outStr)
	options2 = append(options2, "--variant")
	options2 = append(options2, variant.Name)
	if variant.Options.Wait > 0 {
		options2 = append(options2, "--wait_init", strconv.Itoa(variant.Options.Wait))
	} else if project.Options.Wait > 0 {
		options2 = append(options2, "--wait_init", strconv.Itoa(project.Options.Wait))
	}

	return exec.Command(
		pythonexec,
		options2...,
	)
}

func commandSVG(pjtname string, variant string, svg_lib_dirs []string) *exec.Cmd {

	var output []string
	if len(variant) > 0 {
		output = append(output, "CI-BUILD/", path.Base(pjtname), "_", variant, "/SVG/", path.Base(pjtname), "_", variant, ".svg")
	} else {
		output = append(output, "CI-BUILD/", path.Base(pjtname), "/SVG/", path.Base(pjtname), ".svg")
	}
	err := os.MkdirAll(path.Dir(strings.Join(output, "")), 0777)
	if err != nil {
		fmt.Println("Directory couldn't be created!")
	}

	var board []string
	if len(variant) > 0 {
		board = append(board, pjtname, "_", variant, ".kicad_pcb")
	} else {
		board = append(board, pjtname, ".kicad_pcb")
	}

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

func commandSed(regex string, repl string, prjname string, variant string) *exec.Cmd {

	var reg_repl []string
	reg_repl = append(reg_repl, "s|", regex, "|", repl, "|")

	var board []string
	if len(variant) > 0 {
		board = append(board, prjname, "_", variant, ".kicad_pcb")
	} else {
		board = append(board, prjname, ".kicad_pcb")
	}

	return exec.Command(
		"sed",
		"-i",
		"-e",
		strings.Join(reg_repl, ""),
		strings.Join(board, ""),
	)
}

func commandTag(c Commit, pjtname string, variant string, tags Tags) *exec.Cmd {

	var options []string
	var sha string
	if len(c.Sha) > 7 {
		sha = c.Sha[0:8]
	} else if len(c.Sha) > 0 {
		sha = c.Sha
	} else {
		sha = "dummy"
	}

	options = append(options, "-u", tag_script)

	var board []string
	if len(variant) > 0 {
		board = append(board, pjtname, "_", variant)
	} else {
		board = append(board, pjtname)
	}
	options = append(options, "--brd", strings.Join(board, ""))
	options = append(options, "--commit", sha)
	if len(variant) > 0 {
		options = append(options, "--variant", variant)
	} else {
		options = append(options, "--variant", "dummy")
	}

	if tags.All {
		options = append(options,
			"--tag-date",
			"--tag-commit")
		if len(c.Tag) > 0 {
			options = append(options, "--tag-tag", "--tag", c.Tag)
		} else {
			options = append(options, "--tag", "dummy")
		}
		if len(variant) > 0 {
			options = append(options, "--tag-variant")
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
		options = append(options, "--tag-commit")
	}
	if tags.Tag && len(c.Tag) > 0 {
		options = append(options, "--tag-tag", "--tag", c.Tag)
	} else {
		options = append(options, "--tag", "dummy")
	}
	if tags.Variant && len(variant) > 0 {
		options = append(options, "--tag-variant")
	}

	if len(options) > 8 {
		return exec.Command(
			pythonexec,
			options...,
		)
	} else {
		return nil
	}
}

func commandGerber(pjtname string, variant string, lyr GerberLayers) *exec.Cmd {

	var options []string

	options = append(options, "-u", grb_script)
	if len(variant) > 0 {
		options = append(options, "--brd", strings.Join([]string{pjtname, "_", variant}, ""))
	} else {
		options = append(options, "--brd", pjtname)
	}

	var dir []string
	if len(variant) > 0 {
		dir = append(dir, "CI-BUILD/", path.Base(pjtname), "_", variant, "/GRB")
	} else {
		dir = append(dir, "CI-BUILD/", path.Base(pjtname), "/GRB")
	}
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

	if len(options) > 6 {
		return exec.Command(
			pythonexec,
			options...,
		)
	} else {
		return nil
	}
}

func commandSchematic(project Project) *exec.Cmd {

	var options []string
	options = append(options, "-u", sch_script, project.Main)
	if project.Options.Wait > 0 {
		options = append(options, strconv.Itoa(project.Options.Wait))
	}
	var c = exec.Command(
		pythonexec,
		options...,
	)

	c.Env = os.Environ()
	c.Env = append(c.Env, "DEBIAN_FRONTEND=noninteractive")
	c.Env = append(c.Env, "DISPLAY=:0")

	return c
}

func commandBOM(project Project) *exec.Cmd {

	var options []string
	options = append(options, "-u", bom_script, project.Main)
	if project.Options.Wait > 0 {
		options = append(options, strconv.Itoa(project.Options.Wait))
	}
	var c = exec.Command(
		pythonexec,
		options...,
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
