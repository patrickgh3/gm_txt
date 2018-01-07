package main

import (
    "fmt"
    "os"
    "io"
    "io/ioutil"
    "errors"
    "strings"
    "time"
)

func GMObjectFileToHumanObjectFile (path string, destPath string) error {
    // Read GM object file

    f, err := os.Open(path)
    if err != nil {
        return errors.New(fmt.Sprintf("Error opening file %v: %v",
                path, err))
    }
    defer f.Close()

    var obj GMObject = *blankObject()
    err = ReadGMObject(f, &obj)
    if err != nil {
        return errors.New(fmt.Sprintf("Error reading GM object file %v: %v",
                path, err))
    }

    // Write human object file

    f, err = os.Create(destPath)
    if err != nil {
        return errors.New(fmt.Sprintf("Error creating file %v: %v",
                destPath, err))
    }
    defer f.Close()

    err = WriteHumanObject(obj, f, true)
    if err != nil {
        return errors.New(fmt.Sprintf("Error writing human object to %v: %v",
                destPath, err))
    }

    return nil
}

func HumanObjectFileToGMObjectFile (path string,
        destPath string) error {
    // TODO: if gm object exists, load only physics properties from it

    // Read human object file

    f, err := os.Open(path)
    if err != nil {
        return errors.New(fmt.Sprintf("Error opening file %v: %v",
                path, err))
    }
    defer f.Close()

    var obj GMObject = *blankObject()
    err = ReadHumanObject(f, &obj)
    if err != nil {
        return errors.New(fmt.Sprintf("Error reading human object file %v: %v",
                path, err))
    }

    // Write GM object file

    f, err = os.Create(destPath)
    if err != nil {
        return errors.New(fmt.Sprintf("Error creating file %v: %v",
                destPath, err))
    }
    defer f.Close()

    err = WriteGMObject(obj, f)
    if err != nil {
        return errors.New(fmt.Sprintf("Error writing XML to %v: %v",
                destPath, err))
    }

    if len(obj.PhysicsShapePoints) == 0 {
        err = AddEmptyPhysicsShapePoints(destPath)
        if err != nil {
            return errors.New(fmt.Sprintf(
                    "Error writing extra PhysicsShapePoints line to %v: %v",
                    destPath, err))
        }
    }

    return nil
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
    os.Chtimes(projectPath, time.Now(), time.Now())
}
