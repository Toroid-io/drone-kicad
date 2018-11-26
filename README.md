[![Build Status](https://bianca.toroid.io/api/badges/Toroid-io/drone-kicad/status.svg?branch=master)](https://bianca.toroid.io/Toroid-io/drone-kicad)
## drone-kicad

`drone-kicad` is a [drone](https://github.com/drone/drone) plugin for
generating KiCad EDA output files.

At the moment you can generate:

 - Schematics
 - Gerbers
 - SVGs
 - XML and CSV BOMs

It can also tag your board with the current commit, tag and date.

`drone-kicad` can also create variants from your main PCB provided your
components have a field named `variant` and a corresponding value. This
way you can generate independent output for the main PCB and each
variant of your choice.

## Main Options

These apply to the main PCB:

```yml
main: dir/main_pcb              # Path to main file, no extension
options:
  sch: true | false             # Generate schematic pdf outptut
  bom: true | false             # Generate XML and CSV BOMs
  grb:
    all: true | false           # Generate all gerber outputs
    protel: true | false        # Use protel extensions
    fcu: true | false           # Front copper
    bcu: true | false           # Bottom copper
    fmask: true | false         # Front mask
    bmask: true | false         # Bottom mask
    fsilks: true | false        # Front Silk
    bsilks: true | false        # Bottom Silk
    edgecuts: true | false      # Edge Cuts
    drl: true | false           # Drill file
    splitth: true | false       # Split plated/non-plated through holes
  svg: true | false             # Generate SVG plot (Front)
  pcb: true | false             # Export PCB file
  tags:
    all: true | false           # Print all
    tag: true | false           # Print tag
    commit: true | false        # Print commit
    date: true | false          # Print date
  wait: int                     # Delay before variant generation (allows Pcbnew to fully load)
variants:
 - {options for variant 1}
 - {options for variant 2}
```

## Variant Options

These apply to each variant individually:

```yml
name: variant_name              # Your awesome variant name
content: OPT1,OPT2,OPT3         # Comma separated variant field to match
options:                        # Same as before
  grb:
    all: true | false
    protel: true | false
    fcu: true | false
    bcu: true | false
    fmask: true | false
    bmask: true | false
    fsilks: true | false
    bsilks: true | false
    edgecuts: true | false
    drl: true | false
    splitth: true | false
  svg: true | false
  pcb: true | false
  tags:
    all: true | false
    tag: true | false
    commit: true | false
    date: true | false
  wait: int
```

If no `content` is given, all symbols with a non-empty variant field
will be removed.

## Plugin options

```yml
dependencies:
  libraries:
    - https://git.server.com/user/lib           # External libraries
  footprints:
    - https://git.server.com/user/footprints    # External footprints
  modules3d:
    - https://git.server.com/user/modules3d     # External 3D models
  basedir: "/other/than/usr/share/kicad"        # Base directory
  templates:
    - https://git.server.com/user/templates     # External templates
  svglibs:
    - https://git.server.com/user/svglib        # External SVG models
  svglibdirs:
    - relative/path/1                           # SVG lib folder to pass to the svg generator
```

## Tagging

Currently, `drone-kicad` expects a footprint with some text modules with
this content:

 - `$commit$` will be replaced by an eight character commit reference
 - `$tag$` will be replaced by the current tag
 - `$date$` will be replaced by the current date (dd/mm/yyyy)

## Example configuration

```yml
pipeline:
  kicad:
    image: toroid/drone-kicad
    dependencies:
      basedir: "/opt/toroid"
      libraries:
        - https://github.com/toroid-io/toroid-kicad-library
      svg_lib:
        - https://git.server.com/username/awesome-svg-library
        - https://git.server.com/username/awesome-svg-library-2
      svg_lib_dirs:
        - awesome-svg-library/Version1
    projects:
      - main: Project1/project_name
        options:
          bom: true
          pcb: true
          sch: true
          gbr:
            all: true
          svg: true
        variants:
          - name: "Variant1"
            content: "OPT1,OPT2"
            options:
              pcb: true
              grb:
                all: true
          - name: "Variant2"
            content: "OPT1,OPT3,OPT4"
            options:
              svg: true
```

## Output

Output defaults to `CI-BUILD` directory in current directory (repo
root). The previous configuration would lead to the following output
tree:

```
├── project_name
│   ├── GBR
│   │   ├── project_name-B.Cu.gbr
│   │   ├── project_name-B.Mask.gbr
│   │   ├── project_name-B.SilkS.gbr
│   │   ├── project_name.drl
│   │   ├── project_name-Edge.Cuts.gbr
│   │   ├── project_name-F.Cu.gbr
│   │   ├── project_name-F.Mask.gbr
│   │   └── project_name-F.SilkS.gbr
│   ├── SCH
│   │   ├── export_schematic_screencast.ogv
│   │   └── project_name.pdf
│   ├── BOM
│   │   ├── project_name.csv
│   │   ├── project_name.xml
│   │   └── export_bom_screencast.ogv
│   ├── SVG
│   │   └── project_name.svg
│   └── PCB
│       └── project_name.kicad_pcb
├── project_name_Variant1
│   ├── GBR
│   │   ├── project_name_Variant1-B.Cu.gbr
│   │   ├── project_name_Variant1-B.Mask.gbr
│   │   ├── project_name_Variant1-B.SilkS.gbr
│   │   ├── project_name_Variant1.drl
│   │   ├── project_name_Variant1-Edge.Cuts.gbr
│   │   ├── project_name_Variant1-F.Cu.gbr
│   │   ├── project_name_Variant1-F.Mask.gbr
│   │   └── project_name_Variant1-F.SilkS.gbr
│   ├── DLF
│   │   └── dlfVariant1.ogv
│   └── PCB
│       └── project_name_Variant1.kicad_pcb
├── project_name_Variant2
│   ├── SVG
│   │   └── project_name_Variant2.svg
│   └── DLF
│       └── dlfVariant2.ogv
```

The `DLF` folder contains the screen cast for the variants generation process.

## Deploying

You can then take the `CI-BUILD` directory and deploy the results to some server. We use [drone-mella](https://github.com/Toroid-io/drone-mella) sometimes to upload to [OwnCloud](https://owncloud.org/).

## Contributing

Don't hesitate to submit issues or pull requests.

## Base images

We maintain a Docker image of KiCad v5 on top of ArchLinux [here](https://hub.docker.com/r/toroid/kicad-base/).

## License

This project is made available under the GNU General Public License(GPL) version 3 or grater.

KiCad is made available under the GNU General Public License(GPL) version 3 or greater.

PcbDraw is made available under the MIT License.
