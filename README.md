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

## Example configuration

```yml
pipeline:                                                                                                                                                                                      
  kicad:
    image: toroid/drone-kicad
    projects_names:
      - Project1/Project1_BaseName
      - Project2/Project2_BaseName
    schematic: true
    bom: true
    gerber:
      all: true
      protel: true
      splitth: true
```

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
│   └── SCH
│       ├── export_schematic_screencast.ogv
│       └── Project1_BaseName.pdf
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
    └── SCH
        ├── export_schematic_screencast.ogv
        └── Project2_BaseName.pdf
```

## Deploying

You can then take the `CI-BUILD` directory and deploy the results to some server. We use [drone-mella](https://git.toroid.io/drone-plugins/drone-mella) sometimes to upload to [OwnCloud](https://owncloud.org/).

## Contibuting

Don't hesitate to submit issues or pull requests. This is by nature an instable project (see next section).

## Base images

We maintain a squashed Docker image of KiCad develpment version on top of ArchLinux [here](https://hub.docker.com/r/toroid/kicad-base/).

Our libraries and the scripts behind this plugin are added in another image, [here](https://hub.docker.com/r/toroid/kicad/). This way we can accelerate the update process of the plugin without building KiCad every time.