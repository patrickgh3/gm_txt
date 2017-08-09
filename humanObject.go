package main

import (
    "fmt"
    "io"
    "bufio"
    "errors"
    "strings"
)

type EventCode struct {
    Type int
    Number int
}

var eventCodeToName = make(map[EventCode]string)
var eventNameToCode = make(map[string]EventCode)

func addEventCodeTranslation(Name string, Type int, Number int) {
    eventCodeToName[EventCode{Type:Type,Number:Number}] = Name
    eventNameToCode[Name] = EventCode{Type:Type,Number:Number}
}

func InitTranslations () {
    addEventCodeTranslation("Create", 0, 0)
    addEventCodeTranslation("Step", 3, 0)
    addEventCodeTranslation("Collision", 4, 0)
}

func WriteHumanObject (obj GMObject, w io.Writer) error {
    for _, event := range obj.Events.Events {
        name, ok := eventCodeToName[
                EventCode{Type:event.Type,Number:event.Number}]
        if !ok {
            return errors.New(fmt.Sprintf("Unrecognized event code: (%v,%v)",
                    event.Type, event.Number))
        }

        if name == "Collision" {
            fmt.Fprintf(w, "---%v %v\n", name, event.ObjectName)
        } else {
            fmt.Fprintf(w, "---%v\n", name)
        }
        fmt.Fprintf(w, "%v\n", event.Actions[0].Arguments.Arguments[0].String)
    }
    return nil
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

func ReadHumanObject (r io.Reader, obj *GMObject) error {
    // Clear out old events
    obj.Events.Events = make([]*Event, 0)

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
            eventName := tokens[0]
            code, ok := eventNameToCode[eventName]
            if !ok {
                return errors.New(fmt.Sprintf(
                        "Unrecognized event name %v on line %v",
                        eventName, lineNum))
            }

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
            } else {
                if len(tokens) > 1 {
                    return errors.New(fmt.Sprintf(
                            "%v event takes no arguments on line %v",
                            eventName, lineNum))
                }
            }

        // Code lines (all non-event lines are code lines)

        } else {
            if curEvent == nil {
                if len(strings.Trim(line, " ")) != 0 {
                    return errors.New(fmt.Sprintf(
                            "No content allowed above first event: line %v",
                            lineNum))
                }
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
