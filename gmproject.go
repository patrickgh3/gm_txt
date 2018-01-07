package main

import (
    "encoding/xml"
    "io"
)

// GMProject represents a GameMaker .project.gmx XML file.
type GMProject struct {
    XMLName xml.Name`xml:"assets"`
    ObjectsRoot Node`xml:"objects"`
    ScriptsRoot Node`xml:"scripts"`
}

// Node is either a GameMaker resource (e.g. <object>objects\objBread</object>)
// or a resource folder (e.g. <objects name="folder1">...</objects>)
type Node struct {
    Name string`xml:",chardata"` // text of XML node
    Children []Node`xml:",any"`  // will be empty if no children (leaf)
}

func ReadGMProject (r io.Reader, proj *GMProject) error {
    decoder := xml.NewDecoder(r)
    return decoder.Decode(&proj)
}

// WalkNode executes a function for all children resources of a given
// resource folder Node.
// The function is called with e.g. "objects\objBread"
func WalkNode (node Node, walkFunc func(string) error) error {
    for _, n := range node.Children {

        // Call walkFunc on leaf nodes.
        if len(n.Children) == 0 {
            err := walkFunc(n.Name)
            if err != nil {
                return err
            }

        // Recurse into children of non-leaf nodes.
        } else {
            err := WalkNode(n, walkFunc)
            if err != nil {
                return err
            }
        }
    }

    return nil
}
