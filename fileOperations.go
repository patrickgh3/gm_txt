package main

import (
    "fmt"
    "os"
    "bufio"
    "io"
    "io/ioutil"
    "errors"
    "strings"
    "time"
    "encoding/xml"
    "bytes"
)

// Returns (error, shouldSkip)
func GMObjectFileToHumanObjectFile (objFilename string,
        humanFilename string) (error, bool) {
    // Read GM object file

    f, err := os.Open(objFilename)

    if err != nil {
        return errors.New(fmt.Sprintf("Error opening file %v: %v",
                objFilename, err)), false
    }

    defer f.Close()
    decoder := xml.NewDecoder(f)
    var obj GMObject = *blankObject()
    err = decoder.Decode(&obj)

    if err != nil {
        return errors.New(fmt.Sprintf("Error decoding XML from %v: %v",
                objFilename, err)), false
    }

    // Validate GM object

    for _, event := range obj.Events.Events {
        for _, action := range event.Actions {
            if len(action.Arguments.Arguments) == 0 {
                return errors.New("Drag n Drop in an event"), true
            }
        }
    }

    // Write human object file

    f, err = os.Create(humanFilename)

    if err != nil {
        return errors.New(fmt.Sprintf("Error creating file %v: %v",
                humanFilename, err)), false
    }

    defer f.Close()
    w := bufio.NewWriter(f)
    err = WriteHumanObject(obj, w)

    if err != nil {
        return errors.New(fmt.Sprintf("Error writing human object to %v: %v",
                humanFilename, err)), false
    }

    w.Flush()
    return nil, false
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

    w := MyWriter{File:f}
    encoder := xml.NewEncoder(w)
    encoder.Indent("", "  ")
    err = encoder.Encode(obj)
    if err != nil {
        return errors.New(fmt.Sprintf("Error writing XML to %v: %v",
                objFilename, err))
    }

    if len(obj.PhysicsShapePoints) == 0 {
        err = AddEmptyPhysicsShapePoints(objFilename)
        if err != nil {
            return errors.New(fmt.Sprintf(
                    "Error writing extra PhysicsShapePoints line to %v: %v",
                    objFilename, err))
        }
    }

    return nil
}

// MyWriter un-escapes XML newlines and converts to windows-style newlines.
// Implements io.PipeWriter
type MyWriter struct {
    File *os.File
}
func (w MyWriter) Close() error { return w.File.Close() }
func (w MyWriter) CloseWithError(err error) error { return nil }
func (w MyWriter) Write(data []byte) (n int, err error) {
    n = len(data)
    data = bytes.Replace(data, []byte("&#xA;"), []byte("\n"), -1)
    data = bytes.Replace(data, []byte("\n"), []byte("\r\n"), -1)
    _, err = w.File.Write(data)
    return
}

// https://gist.github.com/elazarl/5507969
func cp(dst, src string) error {
    s, err := os.Open(src)
    if err != nil {
        return err
    }
    // no need to check errors on read only file, we already got everything
    // we need from the filesystem, so nothing can go wrong now.
    defer s.Close()
    d, err := os.Create(dst)
    if err != nil {
        return err
    }
    if _, err := io.Copy(d, s); err != nil {
        d.Close()
        return err
    }
    return d.Close()
}

func AppendResourceToGMProject (objName, resType, resDir string) error {
    searchStr := fmt.Sprintf("  </%vs>", resType)
    toInsert := fmt.Sprintf("    <%v>%v\\%v</%v>",
            resType, resDir, objName, resType)
    return FileInsertLine(searchStr, toInsert, projectPath)
}

func AddEmptyPhysicsShapePoints (gmObjFilename string) error {
    return FileInsertLine("</object>", "  <PhysicsShapePoints/>", gmObjFilename)
}

func FileInsertLine (searchFor, toInsert, filename string) error {
    projData, err := ioutil.ReadFile(filename)
    if err != nil {
        return errors.New(fmt.Sprintf("Error opening file: %v",
                projectPath))
    }

    lines := strings.Split(string(projData), "\r\n")
    var i int
    for ii, line := range lines {
        if line == searchFor {
            i = ii
            break
        }
    }
    lines = append(lines[:i], append([]string{toInsert}, lines[i:]...)...)

    projString := strings.Join(lines, "\r\n")
    err = ioutil.WriteFile(filename, []byte(projString), os.ModePerm)
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
