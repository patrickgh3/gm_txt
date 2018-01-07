package main

import (
    "fmt"
    "io"
    "bufio"
    "errors"
    "strings"
    "strconv"
)

type EventCode struct {
    Type int
    Number int
}

var eventCodeToName = make(map[EventCode]string)
var eventNameToCode = make(map[string]EventCode)
var eventCodes []EventCode

func addEventCodeTranslation(Name string, Type int, Number int) {
    c := EventCode{Type:Type,Number:Number}
    eventCodeToName[c] = Name
    eventNameToCode[Name] = c
    eventCodes = append(eventCodes, c)
}

func InitTranslations () {
    f := addEventCodeTranslation
    f("Create", 0, 0)
    f("Destroy", 1, 0)
    f("Alarm", 2, 0)
    f("Step", 3, 0)
    f("Begin Step", 3, 1)
    f("End Step", 3, 2)
    f("Collision", 4, 0)
    f("Keyboard", 5, 0)

    f("Mouse Left Button", 6, 0)
    f("Mouse Right Button", 6, 1)
    f("Mouse Middle Button", 6, 2)
    f("Mouse No Button", 6, 3)
    f("Mouse Left Pressed", 6, 4)
    f("Mouse Right Pressed", 6, 5)
    f("Mouse Middle Pressed", 6, 6)
    f("Mouse Left Released", 6, 7)
    f("Mouse Right Released", 6, 8)
    f("Mouse Middle Released", 6, 9)
    f("Mouse Enter", 6, 10)
    f("Mouse Leave", 6, 11)
    f("Mouse Global Left Button", 6, 50)
    f("Mouse Global Right Button", 6, 51)
    f("Mouse Global Middle Button", 6, 52)
    f("Mouse Global Left Pressed", 6, 53)
    f("Mouse Global Right Pressed", 6, 54)
    f("Mouse Global Middle Pressed", 6, 55)
    f("Mouse Global Left Released", 6, 56)
    f("Mouse Global Right Released", 6, 57)
    f("Mouse Global Middle Released", 6, 58)
    f("Mouse Wheel Up", 6, 60)
    f("Mouse Wheel Down", 6, 61)

    f("Outside Room", 7, 0)
    f("Intersect Boundary", 7, 1)
    f("Game Start", 7, 2)
    f("Game End", 7, 3)
    f("Room Start", 7, 4)
    f("Room End", 7, 5)
    f("No More Lives", 7, 6)
    f("Animation End", 7, 7)
    f("End Of Path", 7, 8)
    f("No More Health", 7, 9)
    f("User Defined", 7, 10) // 10-25

    f("Outside View", 7, 40) //40-47
    f("Boundary View", 7, 50) // 50-57
    f("Animation Update", 7, 58)
    f("Image Loaded", 7, 60)
    f("HTTP", 7, 62)
    f("Dialog", 7, 63)
    f("IAP", 7, 66)
    f("Cloud", 7, 67)
    f("Networking", 7, 68)
    f("Steam", 7, 69)
    f("Social", 7, 70)
    f("Push Notification", 7, 71)
    f("Save / Load", 7, 72)
    f("Audio Recording", 7, 73)
    f("Audio Playback", 7, 74)
    f("System Event", 7, 75)

    f("Draw", 8, 0)
    f("Draw GUI", 8, 64)
    f("Resize", 8, 65)
    f("Draw Begin", 8, 72)
    f("Draw End", 8, 73)
    f("Draw GUI Begin", 8, 74)
    f("Draw GUI End", 8, 75)
    f("Pre Draw", 8, 76)
    f("Post Draw", 8, 77)

    f("Key Press", 9, 0)
    f("Key Release", 10, 0)
}

func WriteHumanObject (obj GMObject, w io.Writer, spaceEvents bool) error {
    // Properties

    if obj.SpriteName != gmUndefinedStr {
        fmt.Fprintf(w, "Sprite %v\n", obj.SpriteName)
    }
    if obj.Visible == 0 {
        fmt.Fprintf(w, "Invisible\n")
    }
    if obj.Solid != 0 {
        fmt.Fprintf(w, "Solid\n")
    }
    if obj.Persistent != 0 {
        fmt.Fprintf(w, "Persistent\n")
    }
    if obj.Depth != 0 {
        fmt.Fprintf(w, "Depth %v\n", obj.Depth)
    }
    if obj.ParentName != gmUndefinedStr {
        fmt.Fprintf(w, "Parent %v\n", obj.ParentName)
    }
    if obj.MaskName != gmUndefinedStr {
        fmt.Fprintf(w, "Mask %v\n", obj.MaskName)
    }

    fmt.Fprintf(w, "\n")

    // Events

    for i, event := range obj.Events.Events {
        // Two newlines between events.
        if i != 0 && spaceEvents {
            fmt.Fprintf(w, "\n")
        }

        // Consolidate the event code, e.g. User Defined 0-11 becomes
        // User Defined 0, before looking it up in the map.
        cc := EventCode{Type:event.Type, Number:event.Number}

        // Alarm, Keyboard, Key Press, Key Release
        if cc.Type == 2 || cc.Type == 5 || cc.Type == 9 || cc.Type == 10 {
            cc.Number = 0
        // User Defined
        } else if (cc.Type == 7 && cc.Number >= 10 && cc.Number <= 25) {
            cc.Number = 10
        // Outside View
        } else if (cc.Type == 7 && cc.Number >= 40 && cc.Number <= 47) {
            cc.Number = 40
        // Boundary View
        } else if (cc.Type == 7 && cc.Number >= 50 && cc.Number <= 57) {
            cc.Number = 50
        }

        // Get the event name from the consolidated event code.
        name, ok := eventCodeToName[cc]
        if !ok {
            return errors.New(fmt.Sprintf("Unrecognized event code: (%v,%v)",
                    event.Type, event.Number))
        }

        // Write event name.
        if name == "Collision" {
            fmt.Fprintf(w, "---%v %v\n", name, event.ObjectName)
        } else if name == "Alarm" || name == "Keyboard" ||
                name == "Key Press" || name == "Key Release" {
            fmt.Fprintf(w, "---%v %v\n", name, event.Number)
        } else if name == "User Defined" {
            fmt.Fprintf(w, "---%v %v\n", name, event.Number-10)
        } else if name == "Outside View" {
            fmt.Fprintf(w, "---%v %v\n", name, event.Number-40)
        } else if name == "Boundary View" {
            fmt.Fprintf(w, "---%v %v\n", name, event.Number-50)
        } else {
            fmt.Fprintf(w, "---%v\n", name)
        }

        // Write the event's GML code.
        // If there are multiple actions ("Execute a piece of code" in GM),
        // write them all sequentially.
        for j, action := range event.Actions {
            if j != 0 {
                fmt.Fprintf(w, "\n")
            }
            fmt.Fprintf(w, "%v", action.Arguments.Arguments[0].String)
        }
    }

    return nil
}

func blankObject() *GMObject {
    return &GMObject{SpriteName:gmUndefinedStr, Solid:0, Visible:-1, Depth:0,
            Persistent:0, ParentName:gmUndefinedStr, MaskName:gmUndefinedStr}
}

func blankEvent() *Event {
    arg := Argument{Kind:1, String:""}
    args := Arguments{Arguments:[]Argument{arg}}
    action := Action{Libid:1, Id:603, Kind:7, UseRelative:0, IsQuestion:0,
            UseApplyTo:-1, ExeType:2, FunctionName:"", CodeString:"",
            WhoName:"self", Relative:0, IsNot:0, Arguments:args}
    actions := []Action{action}
    return &Event{Actions:actions}
}

func parseIntParam (param string, boundsCheck bool, min int, max int) (int, error) {
    if param == "" {
        return 0, errors.New("No number parameter provided")
    }
    i, err := strconv.Atoi(param)
    if err != nil {
        return 0, errors.New("Not a valid number")
    }
    if boundsCheck && (i<min || i>max) {
        return 0, errors.New("Number out of range")
    }
    return i, nil
}

func blankEventFromLine (line string) (*Event, error) {
    // Trim prefix and excess spaces
    name := strings.Trim(line[3:], " ")

    // Split name by spaces
    tokens := strings.Split(name, " ")

    // Get the event's code, assuming there's no parameter.
    // (the whole line is treated as the name)
    code, ok := eventNameToCode[name]
    param := ""

    // If that didn't work, try accounting for a trailing parameter.
    // (the trailing param isn't included in the name)
    if !ok {
        if len(tokens) >= 2 {
            param = tokens[len(tokens)-1]
            name = strings.Join(tokens[:len(tokens)-1], " ")

            code, ok = eventNameToCode[name]
        }

        // If that's still not a match, then the name must be invalid.
        if !ok {
            return nil, errors.New("Unrecognized event title")
        }
    }

    // Create the event
    e := blankEvent()
    e.Type   = code.Type
    e.Number = code.Number

    // Parse event parameter, if there is one
    if name == "Collision" {
        if param == "" {
            return nil, errors.New("Missing string parameter")
        }
        e.ObjectName = param
    } else if name == "Alarm" {
        i, err := parseIntParam(param, true, 0, 11)
        if err != nil {
            return nil, err
        }
        e.Number = i
    } else if name == "Keyboard" || name == "Key Press" ||
            name == "Key Release" {
        i, err := parseIntParam(param, true, 0, 1000)
        if err != nil {
            return nil, err
        }
        e.Number = i
    } else if name == "User Defined" {
        i, err := parseIntParam(param, true, 0, 15)
        if err != nil {
            return nil, err
        }
        e.Number = i + 10
    } else if name == "Outside View" {
        i, err := parseIntParam(param, true, 0, 7)
        if err != nil {
            return nil, err
        }
        e.Number = i + 40
    } else if name == "Boundary View" {
        i, err := parseIntParam(param, true, 0, 7)
        if err != nil {
            return nil, err
        }
        e.Number = i + 50
    }

    return e, nil
}

func applyPropertyLine (line string, obj *GMObject) error {
    // Split line by spaces.
    tokens := strings.Split(strings.Trim(line, " "), " ")

    // Grab trailing parameter.
    param := ""
    if len(tokens) >= 2 {
        param = tokens[len(tokens)-1]
    }

    // Apply the property.
    if tokens[0] == "Invisible" {
        obj.Visible = 0
    } else if tokens[0] == "Solid" {
        obj.Solid = -1
    } else if tokens[0] == "Persistent" {
        obj.Persistent = -1

    } else if tokens[0] == "Sprite" {
        if param == "" {
            return errors.New("Missing string parameter")
        }
        obj.SpriteName = param
    } else if tokens[0] == "Parent" {
        if param == "" {
            return errors.New("Missing string parameter")
        }
        obj.ParentName = param
    } else if tokens[0] == "Mask" {
        if param == "" {
            return errors.New("Missing string parameter")
        }
        obj.MaskName = param

    } else if tokens[0] == "Depth" {
        i, err := parseIntParam(param, false, 0, 0)
        if err != nil {
            return err
        }
        obj.Depth = i

    } else {
        return errors.New(fmt.Sprintf("Unrecognized property %v",
                tokens[0]))
    }

    return nil
}

func ReadHumanObject (r io.Reader, obj *GMObject) error {
    // Clear out old events
    obj.Events.Events = make([]*Event, 0)

    // Scan file line by line
    scanner := bufio.NewScanner(r)
    lineNum := 0
    var curEvent *Event

    for scanner.Scan() {
        line := scanner.Text()
        lineNum++

        // Event title lines
        if len(line) >= 3 && line[0:3] == "---" {
            var err error
            curEvent, err = blankEventFromLine(line)
            if err != nil {
                return errors.New(fmt.Sprintf("Line %v: %v", lineNum, err))
            }
            obj.Events.Events = append(obj.Events.Events, curEvent)

        // Property lines
        } else if curEvent == nil {
            if line != "" {
                err := applyPropertyLine(line, obj)
                if err != nil {
                    return errors.New(fmt.Sprintf("Line %v: %v", lineNum, err))
                }
            }
        // Code lines
        } else {
            curEvent.Actions[0].Arguments.Arguments[0].String += line + "\n"
        }
    }

    // Trim trailing empty lines off of each event's code
    for _, event := range obj.Events.Events {
        event.Actions[0].Arguments.Arguments[0].String = 
                strings.Trim(
                event.Actions[0].Arguments.Arguments[0].String, "\n ")
    }
    return nil
}
