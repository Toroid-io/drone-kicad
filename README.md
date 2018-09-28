[![Build Status](https://bianca.toroid.io/api/badges/Toroid-io/drone-kicad/status.svg?branch=master)](https://bianca.toroid.io/Toroid-io/drone-kicad)
## drone-kicad

`drone-kicad` is a [drone](https://github.com/drone/drone) plugin for generating KiCad EDA output files.

## Options

- `projects_names`: Collection of KiCad projects. It must be the name of the schematic/board file **without** extension. The `basename` is used to create output folders for each project.
- `schematic` : `true|false`. Enables generation of PDF output of all schematic sheets for all projects.
- `bom` : `true|false`. Enables generation of XML bom output for all projects.
- `gerber` : Enables generation of gerber files for selected layers for all projects. Layers are specified as a collection. Following options are available:
    - `all` : Print all layers
    - `fcu, bcu, fmask, bmask, fsilks, bsilks, edgecuts` : Print specified layer
    - `drl` : Print drill file
    - `splitth` : Print PTH and NPTH to different files. Defaul behavior is to merge them (assuming no NPTH is present).
    - `protel` : Use Protel extensions. If no set, files will have a suffix of the form `-LayerName`, e.g.: `project-F.Cu.gbr`
- `deps_dir`: Base directory for dependencies. Defaults to `/usr/share/kicad`.
- `library`: Collection of extra libraries used by the project. They are cloned into `deps_dir/library`.
- `pretty`: Collection of extra footprints used by the project. They are cloned into `deps_dir/footprints`.
- `3d`: Collection of extra 3D models used by the project. They are cloned into `deps_dir/modules/packages3d`.
- `stp`: Enables generation of 3D board in STEP format.
- `template`: Collection of extra templates used by the project. They are cloned into `deps_dir/template`.
- `svg`: `true|false`. Enables the generation of SVG output.
- `svg_lib`: Collection of svg footprint repos used for svg output generation.
- `svg_lib_dirs`: Paths where to look for svg footprints.
- `tags`: Enables board tagging with one or several of the following data:
    - `all` : Prints all fields
    - `date` : Prints date in text element with text `$date$`
    - `commit` : Prints first 8 characters of commit SHA in text element with text `$commit$`
    - `tag` : Prints tag in text element with text `$tag$`

## Example configuration

```yml
pipeline:
  kicad:
    image: toroid/drone-kicad
    projects_names:
      - Project1/Project1_BaseName
      - Project2/Project2_BaseName
    schematic: true
    deps_dir: "/opt/kicad"
    library: https://git.server.com/username/awesome-kicad-library
    bom: true
    gerber:
      all: true
      protel: true
      splitth: true
    svg: true
    svg_lib:
      - https://git.server.com/username/awesome-svg-library
      - https://git.server.com/username/awesome-svg-library-2
    svg_lib_dirs:
      - awesome-svg-library/Version1
    stp: true
    tags:
      commit: true
```

In this example, `awesome-svg-library` may contain footprints under `Version1` directory and `awesome-svg-library-2` may contain footprints under the root directory.

## Output

Output defaults to `CI-BUILD` directory in current directory (repo root). The previous configuration would lead to the following output tree:

```
CI-BUILD
├── Project1_BaseName
│   ├── BOM
│   │   ├── export_bom_screencast.ogv
│   │   └── Project1_BaseName.xml
│   ├── GBR
│   │   ├── Project1_BaseName.gbl
│   │   ├── Project1_BaseName.gbo
│   │   ├── Project1_BaseName.gbs
│   │   ├── Project1_BaseName.gm1
│   │   ├── Project1_BaseName.gtl
│   │   ├── Project1_BaseName.gto
│   │   ├── Project1_BaseName.gts
│   │   ├── Project1_BaseName-NPTH.drl
│   │   └── Project1_BaseName-PTH.drl
│   ├── SCH
│   │   ├── export_schematic_screencast.ogv
│   │   └── Project1_BaseName.pdf
│   ├── STP
│   │   └── Project1_BaseName.stp
│   └── SVG
│       └── Project1_BaseName.svg
└── Project2_BaseName
    ├── BOM
    │   ├── export_bom_screencast.ogv
    │   └── Project2_BaseName.xml
    ├── GBR
    │   ├── Project2_BaseName.gbl
    │   ├── Project2_BaseName.gbo
    │   ├── Project2_BaseName.gbs
    │   ├── Project2_BaseName.gm1
    │   ├── Project2_BaseName.gtl
    │   ├── Project2_BaseName.gto
    │   ├── Project2_BaseName.gts
    │   ├── Project2_BaseName-NPTH.drl
    │   └── Project2_BaseName-PTH.drl
    ├── SCH
    │   ├── export_schematic_screencast.ogv
    │   └── Project2_BaseName.pdf
    ├── STP
    │   └── Project2_BaseName.stp
    └── SVG
        └── Project2_BaseName.svg
```

## Deploying

You can then take the `CI-BUILD` directory and deploy the results to some server. We use [drone-mella](https://github.com/Toroid-io/drone-mella) sometimes to upload to [OwnCloud](https://owncloud.org/).

## Contibuting

Don't hesitate to submit issues or pull requests. This is by nature an instable project (see next section).

## Base images

We maintain a squashed Docker image of KiCad develpment version on top of ArchLinux [here](https://hub.docker.com/r/toroid/kicad-base/).

## License

This project is made available under the GNU General Public License(GPL) version 3 or grater.

KiCad is made available under the GNU General Public License(GPL) version 3 or greater.
PcbDraw is made available under the MIT License.
