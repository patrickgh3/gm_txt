package main

import (
    "github.com/sqweek/dialog"

    "fmt"
    "os"
    "os/signal"
    "bufio"
    "path/filepath"
    "syscall"
    "strings"
    "errors"
)

var projectPath  string
var projectDir   string
var humanDir     string
var gmObjectsDir string
var gmScriptsDir string

const usage string = `Usage:
gm_txt.exe [project_file]
If you provide no arguments, it'll open a file picker dialog.
`

const helpMessage string = `
Translation supports:
1. both objects and scripts,
2. both modifying and creating files,
3. in both directions (GM <--> NiceObjects)

.gml files are scripts and are simply copied back and forth.
.gmo files are translated objects. Type "objects" for the .gmo format.

Note that creating new .gmo or .gml files adds them to the project.
Also note that physics properties are not preserved.

This window will log each translation, as well as translation errors.
`

func main () {
    InitTranslations()

    // Start listening for SIGINT (Ctrl-C)

    sigchan := make(chan os.Signal, 2)
    signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)

    // Parse arguments

    // No args opens windows file picker for project path
    if len(os.Args) < 2 {
        var err error
        projectPath, err = dialog.File().
                Title("Select GMS project file").
                SetStartDir(".").
                Filter("GMS project file (*.project.gmx)", "project.gmx").
                Load()
        if err != nil {
            fmt.Printf("Error in project file select dialog: %v\n", err)
            return
        }

    } else if len(os.Args) == 2 && os.Args[1] == "--help" {
        fmt.Println(usage)
        return

    // One arg (non-help) specifies project path
    } else if len(os.Args) == 2 {
        projectPath = os.Args[1]

    } else {
        fmt.Println(usage)
        return
    }

    // Verify project path is okay.

    fi, err := os.Stat(projectPath)
    if err != nil {
        fmt.Printf("Error Stat-ing project file %v : %v\n",
            projectPath, err)
        return
    }
    if os.IsNotExist(err) {
        fmt.Printf("Project file %v does not exist\n", projectPath)
        return
    }
    if fi.IsDir() {
        fmt.Printf("Must select project file, not directory\n")
        return
    }

    // Compute relative paths.

    projectDir = filepath.Dir(projectPath)
    humanDir     = filepath.Join(projectDir, "gm_txt")
    gmObjectsDir = filepath.Join(projectDir, "objects")
    gmScriptsDir = filepath.Join(projectDir, "scripts")

    // Create human folder

    err = os.MkdirAll(humanDir, os.ModePerm)
    if err != nil {
        fmt.Printf("Error creating gm_txt directory &v: %v\n", humanDir, err)
        return
    }

    // Read GM project file.

    f, err := os.Open(projectPath)
    if err != nil {
        fmt.Printf("Error opening file %v: %v", projectPath, err)
        return
    }
    defer f.Close()

    proj := GMProject{}
    err = ReadGMProject(f, &proj)
    if err != nil {
        fmt.Printf("Error reading project file %v: %v", projectPath, err)
        return
    }

    // Translate all GM objects.

    err = WalkNode (proj.ObjectsRoot, func(subpath string) error {

        // Compute paths.
        resourceName := strings.TrimPrefix(subpath, "objects\\")
        path := filepath.Join(projectDir, subpath)+".object.gmx"
        destPath := filepath.Join(humanDir, resourceName+".gmo")

        // Translate.
        err := GMObjectFileToHumanObjectFile(path, destPath)
        if err != nil {
            return errors.New(fmt.Sprintf("Error translating %v: %v\n",
                    resourceName, err))
        }

        return nil
    })
    if err != nil {
        fmt.Printf("Error during initial translation: %v\n", err)
        return
    }

    // Copy over all GM scripts.

    err = WalkNode (proj.ScriptsRoot, func(subpath string) error {

        // Compute paths.
        path := filepath.Join(projectDir, subpath)
        resourceName := strings.TrimSuffix(
                strings.TrimPrefix(subpath, "scripts\\"),
                ".gml")
        destPath := filepath.Join(humanDir, resourceName+".gml")

        // Translate.
        err := cp(destPath, path)
        if err != nil {
            return errors.New(fmt.Sprintf("Error copying %v: %v\n",
                    resourceName, err))
        }

        return nil
    })
    if err != nil {
        fmt.Printf("Error during initial translation: %v\n", err)
        return
    }

    // Generate cheatsheet.txt
    
    obj := blankObject()
    obj.SpriteName = "sprBread"
    obj.Visible = 0
    obj.Solid = 1
    obj.Persistent = 1
    obj.Depth = 4
    obj.ParentName = "objWheat"
    obj.MaskName = "sprBread"

    e := blankEvent()
    e.Type = 7
    e.Number = 11
    e.Actions[0].Arguments.Arguments[0].String =
`// If you don't include some of the properties listed above, it'll assume the
// default values. (no sprite, visible, non-solid, non-persistent, depth 0,
// no parent, mask same as sprite)

// Below is a list of all event names.
`
    obj.Events.Events = append(obj.Events.Events, e)

    for _, code := range eventCodes {
        name := eventCodeToName[code]
        e := blankEvent()
        e.Type = code.Type
        e.Number = code.Number

        if name == "Alarm" {
            e.Number = 0
        } else if name == "Collision" {
            e.ObjectName = "objBread"
        } else if name == "User Defined" {
            e.Number = 10
        } else if name == "Outside View" {
            e.Number = 40
        } else if name == "Boundary View" {
            e.Number = 50
        }

        obj.Events.Events = append(obj.Events.Events, e)
    }

    destPath := filepath.Join(humanDir, "cheatsheet.txt")
    f, err = os.Create(destPath)
    if err != nil {
        fmt.Printf("Error creating file %v: %v", destPath, err)
    }
    defer f.Close()
    w := bufio.NewWriter(f)
    err = WriteHumanObject(*obj, w, false)
    if err != nil {
        fmt.Printf("Error writing cheatsheet.txt: %v", err)
    }
    w.Flush()

    // Start monitoring files for changes.
    // (the watcher must be in a different goroutine)
    go watch()

    // Surrender this goroutine to the watcher, and wait for control signal

    <-sigchan

    // Upon control signal, remove created directory

    err = os.RemoveAll(humanDir)
    if err != nil {
        fmt.Printf("Error removing gm_txt directory %v:\n%v\n", humanDir, err)
        return
    }
    fmt.Println("Success")
}
