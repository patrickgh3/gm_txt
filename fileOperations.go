package main

import (
    "fmt"
    "os"
    "bufio"
    "io/ioutil"
    "errors"
    "strings"
    "time"
    "encoding/xml"
)

func GMObjectFileToHumanObjectFile (objFilename string,
        humanFilename string) error {
    // Read GM object file

    f, err := os.Open(objFilename)

    if err != nil {
        return errors.New(fmt.Sprintf("Error opening file %v: %v",
                objFilename, err))
    }

    defer f.Close()
    decoder := xml.NewDecoder(f)
    var obj GMObject = *blankObject()
    err = decoder.Decode(&obj)

    if err != nil {
        return errors.New(fmt.Sprintf("Error decoding XML from %v: %v",
                objFilename, err))
    }

    // Write human object file

    f, err = os.Create(humanFilename)

    if err != nil {
        return errors.New(fmt.Sprintf("Error creating file %v: %v",
                humanFilename, err))
    }

    defer f.Close()
    w := bufio.NewWriter(f)
    err = WriteHumanObject(obj, w)

    if err != nil {
        return errors.New(fmt.Sprintf("Error writing human object to %v: %v",
                humanFilename, err))
    }

    w.Flush()
    return nil
}

func HumanObjectFileToGMObjectFile (humanFilename string,
        objFilename string) error {
    // TODO: if gm object exists, load only physics properties from it

    // Read human object file

    f, err := os.Open(humanFilename)
    if err != nil {
        return errors.New(fmt.Sprintf("Error opening file %v: %v",
                humanFilename, err))
    }
    defer f.Close()

    r := bufio.NewReader(f)
    var obj GMObject = *blankObject()
    err = ReadHumanObject(r, &obj)
    if err != nil {
        return errors.New(fmt.Sprintf("Error reading human object file %v: %v",
                humanFilename, err))
    }

    // Write GM object file

    f, err = os.Create(objFilename)
    if err != nil {
        return errors.New(fmt.Sprintf("Error opening file %v: %v",
                objFilename, err))
    }
    defer f.Close()

    encoder := xml.NewEncoder(f)
    encoder.Indent("", "  ")
    err = encoder.Encode(obj)
    if err != nil {
        return errors.New(fmt.Sprintf("Error writing XML to %v: %v",
                objFilename, err))
    }

    return nil
}

func AddObjectToGMProject (objName string) error {
    projData, err := ioutil.ReadFile(projectPath)
    if err != nil {
        return errors.New(fmt.Sprintf("Error opening project file: %v",
                projectPath))
    }

    lines := strings.Split(string(projData), "\r\n")
    var i int
    for ii, line := range lines {
        if line == "  </objects>" {
            i = ii
            break
        }
    }
    toInsert := fmt.Sprintf("    <object>objects\\%v</object>", objName)
    lines = append(lines[:i], append([]string{toInsert}, lines[i:]...)...)

    projString := strings.Join(lines, "\r\n")
    err = ioutil.WriteFile(projectPath, []byte(projString), os.ModePerm)
    if err != nil {
        return errors.New(fmt.Sprintf("Error writing project file: %v",
                projectPath))
    }

    return nil
}

func touchProjectFile () {
    if touchProject {
        os.Chtimes(projectPath, time.Now(), time.Now())
    }
}
