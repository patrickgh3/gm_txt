package main

import (
    "github.com/sqweek/dialog"

    "fmt"
    "os"
    "os/signal"
    "bufio"
    "path/filepath"
    "strings"
    "syscall"
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

    // Translate all GM objects.

    // implements filepath.WalkFunc
    f := func (path string, info os.FileInfo, err error) error {
        // Skip directories and extraneous files.
        if info.IsDir() || !strings.HasSuffix(path, ".object.gmx") {
            return nil
        }

        // Compute translated file path.
        resourceName := strings.TrimSuffix(filepath.Base(path), ".object.gmx")
        destPath := filepath.Join(humanDir, resourceName+".gmo")

        // Translate.
        err = GMObjectFileToHumanObjectFile(path, destPath)
        if err != nil {
            fmt.Printf("Error initially translating %v: %v\n",
                    resourceName, err)
            return err
        }
        return err
    }

    err = filepath.Walk(gmObjectsDir, f)
    if err != nil {
        fmt.Printf("Error during initial translation of all GM objects "+
                "to human objects: %v\n", err)
        return
    }

    // Copy over all GM scripts.

    // implements filepath.WalkFunc
    f = func (path string, info os.FileInfo, err error) error {
        // Skip directories and extraneous files.
        if info.IsDir() || filepath.Ext(path) != ".gml" {
            return nil
        }

        destPath := filepath.Join(humanDir, filepath.Base(path))
        return cp(destPath, path)
    }

    err = filepath.Walk(gmScriptsDir, f)
    if err != nil {
        fmt.Printf("Error during initial copying of scripts: %v\n", err)
        return
    }

    // Start monitoring files for changes
    go watch()

    // Listen for typed commands on Stdin
    go func () {
        scan := bufio.NewScanner(os.Stdin)
        for scan.Scan() {
            text := scan.Text()
            if text == "help" {
                fmt.Println(helpMessage)
            } else if text == "objects" {
                fmt.Println(objectsHelpMessage)
            } else if text == "events" {
                fmt.Println(eventsHelpMessage())
            }
        }
    }()

    // Surrender this goroutine to the watcher, and wait for control signal

    fmt.Println("Listening (Ctrl-C to quit) (type \"help\" for help)")

    <-sigchan

    // Upon control signal, remove created directory

    err = os.RemoveAll(humanDir)
    if err != nil {
        fmt.Printf("Error removing gm_txt directory %v: %v\n", humanDir, err)
        return
    }
    fmt.Println("Success")
}
