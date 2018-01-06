package main

import (
    "github.com/fsnotify/fsnotify"

    "fmt"
    "os"
    "time"
    "strings"
    "path/filepath"
)

// "Reverb" refers to translations causing file Write events back and forth
// between the GM and human folders.

const reverbSpacing time.Duration = 1 * time.Second
const dedupSpacing  time.Duration = 100 * time.Millisecond
var gmChanged       time.Time
var humanChanged    time.Time
var lastGMFileChanged    string
var lastHumanFileChanged string

func watch () {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        fmt.Printf("Error creating fsnotify watcher: %v\n", err)
        return
    }
    defer watcher.Close()

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

    fmt.Println("Up and running. Read gm_txt/cheatsheet.txt")

    for {
        select {
        case event := <-watcher.Events:
            handleEvent(event)
        case err := <-watcher.Errors:
            fmt.Printf("Fsnotify watcher error: %v\n", err)
        }
    }
}

func handleEvent (event fsnotify.Event) {
    if event.Op == fsnotify.Write {
        ext := filepath.Ext(event.Name)
        isHuman := strings.HasPrefix(event.Name, humanDir)
        isHumanObj := isHuman && ext == ".gmo"
        isHumanScript := isHuman && ext == ".gml"
        isGMObj := strings.HasPrefix(event.Name, gmObjectsDir) &&
                ext == ".gmx" // close enough to ".object.gmx"
        isGMScript := strings.HasPrefix(event.Name, gmScriptsDir) &&
                ext == ".gml"

        if isHumanObj {
            if humanFileTimingOk(event.Name) {
                humanChanged = time.Now()
                lastHumanFileChanged = event.Name
                translateHumanObject(event.Name)
            }
        } else if isHumanScript {
            if humanFileTimingOk(event.Name) {
                humanChanged = time.Now()
                lastHumanFileChanged = event.Name
                copyHumanScript(event.Name)
            }
        } else if isGMObj {
            if gmFileTimingOk(event.Name) {
                gmChanged = time.Now()
                lastGMFileChanged = event.Name
                translateGMObject(event.Name)
            }
        } else if isGMScript {
            if gmFileTimingOk(event.Name) {
                gmChanged = time.Now()
                lastGMFileChanged = event.Name
                copyGMScript(event.Name)
            }
        }
    }
}

func humanFileTimingOk (humanFile string) bool {
    return time.Since(gmChanged) > reverbSpacing &&
            (lastHumanFileChanged != humanFile ||
            time.Since(humanChanged) > dedupSpacing)
}

func gmFileTimingOk (gmFile string) bool {
    return time.Since(humanChanged) > reverbSpacing &&
            (lastGMFileChanged != gmFile ||
            time.Since(gmChanged) > dedupSpacing)
}

func copyHumanScript (humanScriptPath string) {
    fn := filepath.Base(humanScriptPath)
    gmScriptPath := filepath.Join(gmScriptsDir, fn)
    scriptName := strings.Split(fn, ".")[0]

    // GM file not existing before translation meanse we have to add it to the
    // project file

    _, err := os.Stat(gmScriptPath)
    gmFileExisted := !os.IsNotExist(err) 

    // Copy script

    err = cp(gmScriptPath, humanScriptPath)
    if err != nil {
        fmt.Printf("[%v] %v\n", time.Now().Format("15:04:05"), err)
    } else {
        fmt.Printf("[%v] Copied %v\n",
                time.Now().Format("15:04:05"), scriptName)
    }

    // If necessary, add to project file

    if !gmFileExisted {
        err = AppendResourceToGMProject(fn, "script", "scripts")
        if err != nil {
            fmt.Printf("[%v] %v\n", time.Now().Format("15:04:05"), err)
        } else {
            fmt.Printf("[%v] Project file updated %v\n",
                    time.Now().Format("15:04:05"), scriptName)
        }
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

    // GM file not existing before translation meanse we have to add it to the
    // project file

    _, err := os.Stat(gmObjPath)
    gmObjFileExisted := !os.IsNotExist(err) 

    // Translate object

    err = HumanObjectFileToGMObjectFile(humanObjPath, gmObjPath)
    if err != nil {
        fmt.Printf("[%v] %v\n", time.Now().Format("15:04:05"), err)
    } else {
        fmt.Printf("[%v] Translated %v\n", time.Now().Format("15:04:05"),
                objName)
        // Touching the project file causes GM:Studio to close all
        // folders, which is annoying.
        //touchProjectFile()
    }

    // If necessary, add to project file

    if !gmObjFileExisted {
        err = AppendResourceToGMProject(objName, "object", "objects")
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
    humanObjPath := filepath.Join(humanDir, objName + ".gmo")

    err := GMObjectFileToHumanObjectFile(gmObjPath, humanObjPath)
    if err != nil {
        fmt.Printf("[%v] (From GM) %v\n", time.Now().Format("15:04:05"), err)
    } else {
        fmt.Printf("[%v] (From GM) Translated %v\n",
                time.Now().Format("15:04:05"), objName)
    }
}
