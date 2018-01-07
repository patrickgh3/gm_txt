package main

import (
    "encoding/xml"
    "bytes"
    "io"
    "errors"
)

const gmUndefinedStr string = "<undefined>"

type GMObject struct {
    XMLName xml.Name`xml:"object"`
    SpriteName string`xml:"spriteName"`
    Solid int`xml:"solid"`
    Visible int`xml:"visible"`
    Depth int`xml:"depth"`
    Persistent int`xml:"persistent"`
    ParentName string`xml:"parentName"`
    MaskName string`xml:"maskName"`

    Events Events`xml:"events"`
    
    PhysicsObject int
    PhysicsObjectSensor int 
    PhysicsObjectShape float32
    PhysicsObjectDensity float32
    PhysicsObjectRestitution float32
    PhysicsObjectGroup int
    PhysicsObjectLinearDamping float32
    PhysicsObjectAngularDamping float32
    PhysicsObjectFriction float32
    PhysicsObjectAwake int
    PhysicsObjectKinematic int
    PhysicsShapePoints []PhysicsShapePoint
}

type PhysicsShapePoint struct {

}

type Events struct {
    Events []*Event`xml:"event"`
}

type Event struct {
    Type int`xml:"eventtype,attr"`
    Number int`xml:"enumb,attr"`
    ObjectName string`xml:"ename,attr,omitempty"`
    Actions []Action`xml:"action"`
}

type Action struct {
    Libid int`xml:"libid"`
    Id int`xml:"id"`
    Kind int`xml:"kind"`
    UseRelative int`xml:"userelative"`
    IsQuestion int`xml:"isquestion"`
    UseApplyTo int`xml:"useapplyto"`
    ExeType int`xml:"exetype"`
    FunctionName string`xml:"functionname"`
    CodeString string`xml:"codestring"`
    WhoName string`xml:"whoName"`
    Relative int`xml:"relative"`
    IsNot int`xml:"isnot"`
    Arguments Arguments`xml:"arguments"`
}

type Arguments struct {
    Arguments []Argument`xml:"argument"`
}

type Argument struct {
    Kind int`xml:"kind"`
    String string`xml:"string"`
}

// MyWriter un-escapes XML newlines and converts to windows-style newlines.
// Implements io.Writer
type MyWriter struct {
    InnerWriter io.Writer
}
func (w MyWriter) Write(data []byte) (n int, err error) {
    // Return the provied # of bytes, even though we may not write
    // that exact amount, otherwise things complain.
    // This may be sketchy!!! But it seems to work fine.
    n = len(data)

    // Convert newlines before writing.
    data = bytes.Replace(data, []byte("&#xA;"), []byte("\n"), -1)
    data = bytes.Replace(data, []byte("\n"), []byte("\r\n"), -1)

    // Discard the # of bytes *actually* written.
    _, err = w.InnerWriter.Write(data)

    return
}

func WriteGMObject (obj GMObject, w io.Writer) error {
    // Encode XML using the special newline replacement.
    m := MyWriter{InnerWriter:w}
    encoder := xml.NewEncoder(m)
    encoder.Indent("", "  ")
    return encoder.Encode(obj)
}

func ReadGMObject (r io.Reader, obj *GMObject) error {
    // Decode the XML.
    decoder := xml.NewDecoder(r)
    err := decoder.Decode(&obj)
    if err != nil {
        return err
    }

    // Error if object contains any drag&drop actions.
    for _, event := range obj.Events.Events {
        for _, action := range event.Actions {
            if action.ExeType != 2 {
                return errors.New("Drag&drop in an event")
            }
        }
    }

    return nil
}
