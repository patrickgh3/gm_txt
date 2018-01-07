package main

import (
    "github.com/fsnotify/fsnotify"

    "fmt"
    "os"
    "time"
    "strings"
    "path/filepath"
)

const timeFormat = "15:04:05"

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
        panic(fmt.Sprintf("Error creating watcher: %v\n", err))
    }
    defer watcher.Close()

    if err := watcher.Add(humanDir); err != nil {
        fmt.Printf("Error adding dir to watcher: %v\n", err)
    }
    if err := watcher.Add(gmObjectsDir); err != nil {
        fmt.Printf("Error adding dir to watcher: %v\n", err)
    }
    if err := watcher.Add(gmScriptsDir); err != nil {
        fmt.Printf("Error adding dir to watcher: %v\n", err)
    }

    fmt.Println(humanDir)
    fmt.Println("Up and running. Read gm_txt/cheatsheet.txt")

    // Watch forever.

    for {
        select {
        case event := <-watcher.Events:
            if event.Op == fsnotify.Write {
                handleFileWritten(event.Name)
            }
        case err := <-watcher.Errors:
            fmt.Printf("Fsnotify watcher error: %v\n", err)
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

func handleFileWritten (path string) {
    ext := filepath.Ext(path)
    inHumanDir := strings.HasPrefix(path, humanDir)

    // Human folder object file written
    if inHumanDir && ext == ".gmo" {
        if humanFileTimingOk(path) {
            humanChanged = time.Now()
            lastHumanFileChanged = path

            translateHumanObject(path)
        }
    // Human folder script file written
    } else if inHumanDir && ext == ".gml" {
        if humanFileTimingOk(path) {
            humanChanged = time.Now()
            lastHumanFileChanged = path

            copyHumanScript(path)
        }

    // GameMaker object file written

    } else if strings.HasPrefix(path, gmObjectsDir) &&
            strings.HasSuffix(path, ".object.gmx") {
        if gmFileTimingOk(path) {
            gmChanged = time.Now()
            lastGMFileChanged = path

            // Compute translated file path.
            resourceName := strings.TrimSuffix(filepath.Base(path),
                    ".object.gmx")
            destPath := filepath.Join(humanDir, resourceName+".gmo")

            // Translate.
            err := GMObjectFileToHumanObjectFile(path, destPath)
            if err != nil {
                fmt.Printf("[%v] (From GM) %v\n",
                        time.Now().Format(timeFormat), resourceName, err)
            } else {
                fmt.Printf("[%v] (From GM) Translated %v\n",
                        time.Now().Format(timeFormat), resourceName)
            }
        }

    // GameMaker script file written

    } else if strings.HasPrefix(path, gmScriptsDir) && ext == ".gml" {
        if gmFileTimingOk(path) {
            gmChanged = time.Now()
            lastGMFileChanged = path

            resourceName := strings.TrimSuffix(filepath.Base(path), ".gml")
            destPath := filepath.Join(humanDir, resourceName+".gml")

            err := cp(destPath, path)
            if err != nil {
                fmt.Printf("[%v] (From GM) %v\n",
                        time.Now().Format(timeFormat), err)
            } else {
                fmt.Printf("[%v] (From GM) Copied %v\n",
                        time.Now().Format(timeFormat), resourceName)
            }
        }
    }
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
        fmt.Printf("[%v] %v\n", time.Now().Format(timeFormat), err)
    } else {
        fmt.Printf("[%v] Copied %v\n",
                time.Now().Format(timeFormat), scriptName)
    }

    // If necessary, add to project file

    if !gmFileExisted {
        err = AppendResourceToGMProject(fn, "script", "scripts")
        if err != nil {
            fmt.Printf("[%v] %v\n", time.Now().Format(timeFormat), err)
        } else {
            fmt.Printf("[%v] Project file updated %v\n",
                    time.Now().Format(timeFormat), scriptName)
        }
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
        fmt.Printf("[%v] %v\n", time.Now().Format(timeFormat), err)
    } else {
        fmt.Printf("[%v] Translated %v\n", time.Now().Format(timeFormat),
                objName)
        // Touching the project file causes GM:Studio to close all
        // folders, which is annoying.
        //touchProjectFile()
    }

    // If necessary, add to project file

    if !gmObjFileExisted {
        err = AppendResourceToGMProject(objName, "object", "objects")
        if err != nil {
            fmt.Printf("[%v] %v\n", time.Now().Format(timeFormat), err)
        } else {
            fmt.Printf("[%v] Project file updated %v\n",
                    time.Now().Format(timeFormat), objName)
        }
    }
}
