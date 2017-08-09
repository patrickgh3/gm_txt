package main

import (
    "fmt"
    "os"
    "bufio"
    "encoding/xml"
    "errors"
)

const humanDir string = "NiceObjects"

func main () {
    InitTranslations()

    // Create human folder

    err := os.MkdirAll(humanDir, os.ModePerm)
    if err != nil {
        fmt.Printf("Error creating NiceObjects directory: %v\n", err)
        return
    }

    // Test translation

    /*err = GMObjectFileToHumanObjectFile(
            "example.gmx/objects/objPlayer.object.gmx",
            humanDir+"/objPlayer")*/
    err = HumanObjectFileToGMObjectFile(
            humanDir+"/objPlayer",
            "example.gmx/objects/objPlayer.object.gmx")
    if err != nil {
        fmt.Printf("Error translating test file: %v\n", err)
    }
}

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
    var obj GMObject
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
    // Read GM object file

    f, err := os.Open(objFilename)

    if err != nil {
        return errors.New(fmt.Sprintf("Error opening file %v: %v",
                objFilename, err))
    }

    defer f.Close()
    decoder := xml.NewDecoder(f)
    var obj GMObject
    err = decoder.Decode(&obj)

    if err != nil {
        return errors.New(fmt.Sprintf("Error decoding XML from %v: %v",
                objFilename, err))
    }

    // Read human object file

    f, err = os.Open(humanFilename)
    if err != nil {
        return errors.New(fmt.Sprintf("Error opening file %v: %v",
                humanFilename, err))
    }
    defer f.Close()

    r := bufio.NewReader(f)
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
