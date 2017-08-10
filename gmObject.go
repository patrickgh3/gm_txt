package main

import (
    "encoding/xml"
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
