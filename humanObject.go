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

    for _, event := range obj.Events.Events {
        ec := EventCode{Type:event.Type,Number:event.Number}
        if ec.Type == 2 {
            ec.Number = 0
        } else if (ec.Type == 7 && ec.Number >= 10 && ec.Number <= 25) {
            ec.Number = 10
        } else if (ec.Type == 7 && ec.Number >= 40 && ec.Number <= 47) {
            ec.Number = 40
        } else if (ec.Type == 7 && ec.Number >= 50 && ec.Number <= 57) {
            ec.Number = 50
        } else if (ec.Type == 5 || ec.Type == 9 || ec.Type == 10) {
            ec.Number = 0
        }
        name, ok := eventCodeToName[ec]
        if !ok {
            return errors.New(fmt.Sprintf("Unrecognized event code: (%v,%v)",
                    event.Type, event.Number))
        }

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

        for _, action := range event.Actions {
            fmt.Fprintf(w, "%v", action.Arguments.Arguments[0].String)
        }

        if spaceEvents {
            fmt.Fprintf(w, "\n\n")
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

func parseIntArg(tokens []string, evName string, lineNum int,
        intPos int, min int, max int) (int, error) {
    if len(tokens) < intPos+1 {
        return 0, errors.New(fmt.Sprintf(
                "%v event needs number on line %v",
                evName, lineNum))
    }
    i, err := strconv.Atoi(tokens[intPos])
    if err != nil || i<min || i>max {
        return 0, errors.New(fmt.Sprintf(
                "%v event has invalid number on line %v",
                evName, lineNum))
    }
    return i, nil
}

func ReadHumanObject (r io.Reader, obj *GMObject) error {
    // Clear out old events
    obj.Events.Events = make([]*Event, 0)

    // Scan file line by line

    scanner := bufio.NewScanner(r)
    lineNum := 0
    var curEvent *Event

    for scanner.Scan() {
        lineNum++
        line := scanner.Text()

        // Event title lines

        if len(line) >= 3 && line[0:3] == "---" {
            eventTitle := strings.Trim(line[3:], " ")
            if len(eventTitle) == 0 {
                return errors.New(fmt.Sprintf("Empty event title on line %v",
                        lineNum))
            }
            tokens := strings.Split(eventTitle, " ")

            // Try entire title as name

            eventName := eventTitle
            code, ok := eventNameToCode[eventName]
            if !ok {
                // Try first token as name

                eventName = tokens[0]
                code, ok = eventNameToCode[eventName]

                if !ok {
                    // Try first two tokens as name

                    if len(tokens) >= 2 {
                        eventName = tokens[0] + " " + tokens[1]
                    }
                    code, ok = eventNameToCode[eventName]

                    if !ok {
                        return errors.New(fmt.Sprintf(
                                "Unrecognized event title %v on line %v",
                                eventName, lineNum))
                    }
                }
            }


            // Create the event

            curEvent = blankEvent()
            obj.Events.Events = append(obj.Events.Events, curEvent)

            curEvent.Type = code.Type
            curEvent.Number = code.Number
            if eventName == "Collision" {
                if len(tokens) < 2 {
                    return errors.New(fmt.Sprintf(
                            "Collision event needs object name on line %v",
                            lineNum))
                } else {
                    curEvent.ObjectName = tokens[1]
                }
            } else if eventName == "Alarm" {
                i, err := parseIntArg(
                        tokens, eventName, lineNum, 1, 0, 11)
                if err != nil {
                    return err
                }
                curEvent.Number = i
            } else if eventName == "Keyboard" || eventName == "Key Press" ||
                    eventName == "Key Release" {
                i, err := parseIntArg(
                        tokens, eventName, lineNum,
                        len(strings.Split(eventName, " ")), 0, 1000)
                if err != nil {
                    return err
                }
                curEvent.Number = i
            } else if eventName == "User Defined" {
                i, err := parseIntArg(
                        tokens, eventName, lineNum, 2, 0, 15)
                if err != nil {
                    return err
                }
                curEvent.Number = i + 10
            } else if eventName == "Outside View" {
                i, err := parseIntArg(
                        tokens, eventName, lineNum, 2, 0, 7)
                if err != nil {
                    return err
                }
                curEvent.Number = i + 40
            } else if eventName == "Boundary View" {
                i, err := parseIntArg(
                        tokens, eventName, lineNum, 2, 0, 7)
                if err != nil {
                    return err
                }
                curEvent.Number = i + 50
            }

        // Non-event lines are property or code lines

        } else {
            // Property
            if curEvent == nil {
                tokens := strings.Split(strings.Trim(line, " "), " ")
                if tokens[0] != "" {
                    if tokens[0] == "Sprite" {
                        if len(tokens) < 2 {
                            return errors.New(fmt.Sprintf(
                                    "Sprite property needs name on line %v",
                                    lineNum))
                        }
                        obj.SpriteName = tokens[1]
                    } else if tokens[0] == "Invisible" {
                        obj.Visible = 0
                    } else if tokens[0] == "Solid" {
                        obj.Solid = -1
                    } else if tokens[0] == "Persistent" {
                        obj.Persistent = -1
                    } else if tokens[0] == "Depth" {
                        if len(tokens) < 2 {
                            return errors.New(fmt.Sprintf(
                                    "Sprite property needs name on line %v",
                                    lineNum))
                        }
                        i, err := strconv.Atoi(tokens[1])
                        if err != nil {
                            return errors.New(fmt.Sprintf(
                                    "Invalid depth number %v on line %v\n",
                                    tokens[1], lineNum))
                        }
                        obj.Depth = i
                    } else if tokens[0] == "Parent" {
                        if len(tokens) < 2 {
                            return errors.New(fmt.Sprintf(
                                    "Parent property needs name on line %v",
                                    lineNum))
                        }
                        obj.ParentName = tokens[1]
                    } else if tokens[0] == "Mask" {
                        if len(tokens) < 2 {
                            return errors.New(fmt.Sprintf(
                                    "Mask property needs name on line %v",
                                    lineNum))
                        }
                        obj.MaskName = tokens[1]
                    } else {
                        fmt.Printf("len tokens: %v\n", len(tokens))
                        return errors.New(fmt.Sprintf(
                                "Unrecognized object property %v on line %v\n",
                                tokens[0], lineNum))
                    }
                }
            // Code
            } else {
                curEvent.Actions[0].Arguments.Arguments[0].String += line + "\n"
            }
        }
    }

    // Trim trailing empty lines off of each event's code
    // (keep the last newline though)
    for _, event := range obj.Events.Events {
        event.Actions[0].Arguments.Arguments[0].String = 
                strings.Trim(
                event.Actions[0].Arguments.Arguments[0].String, "\n ") + "\n"
    }
    return nil
}
