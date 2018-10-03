[![Build Status](https://bianca.toroid.io/api/badges/Toroid-io/drone-kicad/status.svg?branch=master)](https://bianca.toroid.io/Toroid-io/drone-kicad)
## drone-kicad

`drone-kicad` is a [drone](https://github.com/drone/drone) plugin for
generating KiCad EDA output files.

At the moment you can generate:

 - Schematics
 - Gerbers
 - SVGs

It can also tag your board with the current commit, tag and date.

`drone-kicad` can also create variants from your main PCB provided your
components have a field named `variant` and a corresponding value. This
way you can generate independant output for the main PCB and each
variant of your choice.

## Main Options

These apply to the main PCB:

```
  code: project_code
  main: dir/main_pcb
  options:
    sch: true | false
    bom: true | false
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
    tags:
      all: true | false
      tag: true | false
      commit: true | false
      date: true | false
  variants:
   - {options for variant 1}
   - {options for variant 2}
```

## Variant Options

These apply to each variant individually:

```
  name: variant_name
  content: OPT1,OPT2,OPT3
  options:
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
    tags:
      all: true | false
      tag: true | false
      commit: true | false
      date: true | false
```

## Plugin options

```
  dependencies:
    libraries:
      - https://git.server.com/user/lib           // External libraries
    footprints:
      - https://git.server.com/user/footprints    // External footprints
    modules3d:
      - https://git.server.com/user/modules3d     // External 3D models
    basedir: "/other/than/usr/share/kicad"        // Base directory
    templates:
      - https://git.server.com/user/templates     // External templates
    svglibs:
      - https://git.server.com/user/svglib        // External SVG models
    svglibdirs:
      - relative/path/1                           // SVG lib folder to pass to the svg generator
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
      - main: Project1/Project1_BaseName
        code: PRJ_01
        options:
          sch: true
          gbr:
            all: true
            protel: true
          bom: true
          svg: true
          tags:
            commit: true

## Output

Output defaults to `CI-BUILD` directory in current directory (repo
root). The previous configuration would lead to the following output
tree:

```
```

## Deploying

You can then take the `CI-BUILD` directory and deploy the results to some server. We use [drone-mella](https://github.com/Toroid-io/drone-mella) sometimes to upload to [OwnCloud](https://owncloud.org/).

## Contributing

Don't hesitate to submit issues or pull requests. This is by nature an instable project (see next section).

## Base images

We maintain a squashed Docker image of KiCad development version on top of ArchLinux [here](https://hub.docker.com/r/toroid/kicad-base/).

## License

This project is made available under the GNU General Public License(GPL) version 3 or grater.

KiCad is made available under the GNU General Public License(GPL) version 3 or greater.

PcbDraw is made available under the MIT License.
