package main

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "github.com/fsnotify/fsnotify"
    "time"
    "os/signal"
    "syscall"
    "github.com/sqweek/dialog"
)

var humanDir, _ = filepath.Abs("./NiceObjects")
var projectPath  string
var projectDir   string
var gmObjectsDir string
var gmScriptsDir string

const reverbSpacing time.Duration = 1 * time.Second
var gmChanged       time.Time
var humanChanged    time.Time

const usage string = `Usage:
NiceObjects.exe     (Opens flie picker)
NiceObjects.exe path/to/project.project.gmx`

const touchProject = false

func main () {
    InitTranslations()

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

    // Verify project path ok and compute relative paths

    _, err := os.Stat(projectPath)
    if err != nil {
        fmt.Printf("Error Stat-ing project file %v : %v\n",
            projectPath, err)
        return
    }
    if os.IsNotExist(err) {
        fmt.Printf("Project file %v does not exist\n", projectPath)
        return
    }
    projectDir, _ = filepath.Split(projectPath)
    gmObjectsDir = filepath.Join(projectDir, "objects")
    gmScriptsDir = filepath.Join(projectDir, "scripts")

    // Start listening for SIGINT (Ctrl-C)

    sigchan := make(chan os.Signal, 2)
    signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)

    // Create human folder

    fmt.Println("Initializing NiceObjects directory...")

    err = os.MkdirAll(humanDir, os.ModePerm)
    if err != nil {
        fmt.Printf("Error creating NiceObjects directory: %v\n", err)
        return
    }

    // Translate all objects and copy all scripts into human folder

    err = filepath.Walk(gmObjectsDir, initialTranslateWalkFunc)
    if err != nil {
        fmt.Printf("Error during initial translation of all GM objects "+
                "to human objects: %v\n", err)
        return
    }

    err = filepath.Walk(gmScriptsDir,
            func (path string, info os.FileInfo, err error) error {
                if info.IsDir() || filepath.Ext(path) != ".gml" {
                    return nil
                }
                fn := filepath.Base(path)
                humanScriptPath := filepath.Join(humanDir, fn)
                return cp(humanScriptPath, path)
            })
    if err != nil {
        fmt.Printf("Error during initial copying of scripts: %v\n", err)
        return
    }

    // Start monitoring files for changes

    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        fmt.Printf("Error creating fsnotify watcher: %v\n", err)
        return
    }
    defer watcher.Close()

    // Watcher must be in a separate goroutine
    go func () {
        for {
            select {
            case event := <-watcher.Events:
                processWatcherEvent(event)
            case err := <-watcher.Errors:
                fmt.Printf("Fsnotify watcher error: %v\n", err)
            }
        }
    }()

    if err := watcher.Add(humanDir); err != nil {
        fmt.Printf("Error assigning human dir to fsnotify watcher: %v\n", err)
    }
    if err := watcher.Add(gmObjectsDir); err != nil {
        fmt.Printf("Error assigning GM objects dir to fsnotify watcher: %v\n",
                err)
    }
    if err := watcher.Add(gmScriptsDir); err != nil {
        fmt.Printf("Error assigning GM scripts dir to fsnotify watcher: %v\n",
                err)
    }

    // Surrender this goroutine to the watcher, and wait for control signal

    fmt.Println("Listening (Ctrl-C to quit)")

    <-sigchan

    // Upon control signal, remove created directory

    fmt.Println("Quit signal received, removing NiceObjects directory...")
    err = os.RemoveAll(humanDir)
    if err != nil {
        fmt.Printf("Error removing NiceObjects directory: %v\n", err)
        return
    }
    fmt.Println("Success")
}

func processWatcherEvent (event fsnotify.Event) {
    ext := filepath.Ext(event.Name)
    isHuman := strings.HasPrefix(event.Name, humanDir)
    isHumanObj := isHuman && ext == ".ogml"
    isHumanScript := isHuman && ext == ".gml"
    isGMObj := strings.HasPrefix(event.Name, gmObjectsDir) &&
            ext == ".gmx" // close enough to ".object.gmx"
    isGMScript := strings.HasPrefix(event.Name, gmScriptsDir) &&
            ext == ".gml"

    if isHumanObj {
        if event.Op == fsnotify.Write {
            if time.Since(gmChanged) > reverbSpacing {
                humanChanged = time.Now()
                translateHumanObject(event.Name)
            }
        } else if event.Op == fsnotify.Create {
            humanChanged = time.Now()
            translateAndAddHumanObject(event.Name)
        }
    } else if isHumanScript {
        // TODO: Create
        // also TODO: do we even need to listen to creates?
        // can we just call translateAndAddHumanObject(event.Name) each write?
        if event.Op == fsnotify.Write {
            if time.Since(gmChanged) > reverbSpacing {
                humanChanged = time.Now()
                copyHumanScript(event.Name)
            }
        }
    } else if isGMObj {
        if event.Op == fsnotify.Write {
            if time.Since(humanChanged) > reverbSpacing {
                gmChanged = time.Now()
                translateGMObject(event.Name)
            }
        }
    } else if isGMScript {
        if event.Op == fsnotify.Write {
            if time.Since(humanChanged) > reverbSpacing {
                gmChanged = time.Now()
                copyGMScript(event.Name)
            }
        }
    }
}

func copyHumanScript (humanScriptPath string) {
    fn := filepath.Base(humanScriptPath)
    gmScriptPath := filepath.Join(gmScriptsDir, fn)
    err := cp(gmScriptPath, humanScriptPath)
    if err != nil {
        fmt.Printf("[%v] %v\n", time.Now().Format("15:04:05"), err)
    } else {
        scriptName := strings.Split(fn, ".")[0]
        fmt.Printf("[%v] Copied %v\n",
                time.Now().Format("15:04:05"), scriptName)
    }
}

func copyGMScript (gmScriptPath string) {
    fn := filepath.Base(gmScriptPath)
    humanScriptPath := filepath.Join(humanDir, fn)
    err := cp(humanScriptPath, gmScriptPath)
    if err != nil {
        fmt.Printf("[%v] (From GM) %v\n", time.Now().Format("15:04:05"), err)
    } else {
        scriptName := strings.Split(fn, ".")[0]
        fmt.Printf("[%v] (From GM) Copied %v\n",
                time.Now().Format("15:04:05"), scriptName)
    }
}

func translateHumanObject (humanObjPath string) {
    objName := strings.Split(filepath.Base(humanObjPath), ".")[0]
    gmObjPath := filepath.Join(gmObjectsDir, objName + ".object.gmx")

    err := HumanObjectFileToGMObjectFile(humanObjPath, gmObjPath)
    if err != nil {
        fmt.Printf("[%v] %v\n", time.Now().Format("15:04:05"), err)
    } else {
        fmt.Printf("[%v] Translated %v\n",
                time.Now().Format("15:04:05"), objName)
        touchProjectFile()
    }
}

func translateAndAddHumanObject (humanObjPath string) {
    objName := strings.Split(filepath.Base(humanObjPath), ".")[0]
    gmObjPath := filepath.Join(gmObjectsDir, objName + ".object.gmx")

    // GM file not existing before translation meanse we have to add it to the
    // project file

    _, err := os.Stat(gmObjPath)
    gmObjFileExisted := !os.IsNotExist(err) 

    // Translate object

    err = HumanObjectFileToGMObjectFile(humanObjPath, gmObjPath)
    if err != nil {
        fmt.Printf("[%v] %v\n", time.Now().Format("15:04:05"), err)
        return
    }
    fmt.Printf("[%v] Translated %v\n", time.Now().Format("15:04:05"),
            objName)
    touchProjectFile()

    // If necessary, add to project file

    if !gmObjFileExisted {
        err = AddObjectToGMProject(objName)
        if err != nil {
            fmt.Printf("[%v] %v\n", time.Now().Format("15:04:05"), err)
        } else {
            fmt.Printf("[%v] Project file updated %v\n",
                    time.Now().Format("15:04:05"), objName)
        }
    }
}

func translateGMObject (gmObjPath string) {
    objName := strings.Split(filepath.Base(gmObjPath), ".")[0]
    humanObjPath := filepath.Join(humanDir, objName + ".ogml")

    err := GMObjectFileToHumanObjectFile(gmObjPath, humanObjPath)
    if err != nil {
        fmt.Printf("[%v] (From GM) %v\n", time.Now().Format("15:04:05"), err)
    } else {
        fmt.Printf("[%v] (From GM) Translated %v\n",
                time.Now().Format("15:04:05"), objName)
    }
}

func initialTranslateWalkFunc (gmObjPath string,
        info os.FileInfo, err error) error {
    if info.IsDir() {
        return nil
    }
    dotSep := strings.SplitN(filepath.Base(gmObjPath), ".", 2)
    if !(len(dotSep) == 2 && dotSep[1] == "object.gmx") {
        return nil
    }
    objName := dotSep[0] + ".ogml"
    humanObjPath := filepath.Join(humanDir, objName)
    return GMObjectFileToHumanObjectFile(gmObjPath, humanObjPath)
}

