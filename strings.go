package main

import (
    "fmt"
    "sort"
)

const helpMessage string = `
Translation supports:
1. both objects and scripts,
2. both modifying and creating files,
3. in both directions (GM <--> NiceObjects)

.gml files are scripts and are simply copied back and forth.
.gmo files are translated objects. Type "objects" for the .gmo format.

Note that creating new .gmo or .gml files adds them to the project.
Also note that physics properties are not preserved.

This window will log each translation, as well as translation errors.
`

const objectsHelpMessage = `
.gmo files are formatted as follows:
Property
Property
---Event Name
//code
---Event Name
//code

All possible Properties:
Sprite [sprite]
Invisible
Solid
Persistent
Depth [depth]
Parent [parent]
Mask [mask]
Omitting a Property assumes the default value for a new object.

Event Names are what you'd expect, such as Create, Alarm 0, Collision objPlayer.
Type "events" for a complete list.
`

func eventsHelpMessage () string {
    s := "\n"
    names := []string{}
    for _, v := range eventCodeToName {
        names = append(names, v)
    }
    sort.Strings(names)
    i := 0
    for _, name := range names {
        if i == 3 {
            i = 0
            s += "\n"
        }
        if name == "Alarm" {
            name += " [num]"
        } else if name == "Collision" {
            name += " [obj]"
        } else if name == "User Defined" || name == "Boundary View" ||
                name == "Outside View" || name == "Keyboard" ||
                name == "Key Press" || name == "Key Release" {
            name += " [num]"
        }
        s += fmt.Sprintf("%-25s", name)
        i++
    }
    s += "\nNote that keyboard events have numbers, not key names, "+
            "for simplicity.\n"
    return s
}

