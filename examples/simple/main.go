package main

import (
	"encoding/json"
	"fmt"
	agollo "github.com/Shonminh/apollo-client"
	"log"
)

var logger golog

type golog struct {
}

func (l golog) Debugf(template string, args ...interface{}) {
	log.Printf("debug:"+template, args...)
}
func (l golog) Infof(template string, args ...interface{}) {
	log.Printf("info:"+template, args...)
}
func (l golog) Warnf(template string, args ...interface{}) {
	log.Printf("warn:"+template, args...)
}
func (l golog) Errorf(template string, args ...interface{}) {
	log.Printf("error:"+template, args...)
}
func (l golog) Fatalf(template string, args ...interface{}) {
	log.Panicf("fatal:"+template, args...)
}

func main() {
	fmt.Println("Initilize agollo")
	defaultVals := map[string]interface{}{
		"a": 111,
		"b": "this is a default value",
		"c": true,
		"d": 11.22,
	}

	err := agollo.Init(agollo.WithConfFile("./apollo.json"),
		agollo.WithDefaultVals(defaultVals, "application"),
		agollo.WithLogFunc(logger.Debugf, logger.Infof, logger.Errorf),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	agollo.RegChangeEventHandler(HandleAll)

	fmt.Println("Initilization done")

	cr := agollo.GetConfigReader("application")

	temp, _ := cr.GetIntValue("timeout")
	fmt.Printf("\nGetIntValue  timeout = %v\n", temp)
	temp, _ = cr.GetIntValue("a")
	fmt.Printf("GetIntValue  a = %v\n", temp)
	temp, _ = cr.GetIntValue("zero")
	fmt.Printf("GetIntValue  zero = %v\n", temp)

	temp2, _ := cr.GetStringValue("hello")
	fmt.Printf("GetStringValue  hello = %v\n", temp2)
	temp2, _ = cr.GetStringValue("b")
	fmt.Printf("GetStringValue  b = %v\n", temp2)
	temp2, _ = cr.GetStringValue("zero")
	fmt.Printf("GetStringValue  zero = %v\n", temp2)

	temp3, _ := cr.GetBoolValue("ok")
	fmt.Printf("GetBoolValue  ok = %v\n", temp3)
	temp3, _ = cr.GetBoolValue("c")
	fmt.Printf("GetBoolValue  c = %v\n", temp3)
	temp3, _ = cr.GetBoolValue("zero")
	fmt.Printf("GetBoolValue  zero = %v\n", temp3)

	temp4, _ := cr.GetFloatValue("pi")
	fmt.Printf("GetFloatValue pi = %v\n", temp4)
	temp4, _ = cr.GetFloatValue("d")
	fmt.Printf("GetFloatValue d = %v\n", temp4)
	temp4, _ = cr.GetFloatValue("zero")
	fmt.Printf("GetFloatValue zero = %v\n", temp4)

	go agollo.Start()

	fmt.Println("Wait for config change")
	var tmpChan = make(chan string)
	<-tmpChan
}

// HandleAll is a fake handler
func HandleAll(event *agollo.ChangeEvent) error {
	fmt.Println("Hi, this is a Callback!")

	changeEvent := event
	bytes, _ := json.Marshal(changeEvent)
	fmt.Println("event:", string(bytes))

	return nil
}
