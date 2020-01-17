package main

import (
	"encoding/json"
	"fmt"
	"github.com/Shonminh/apollo-client"
	"log"
	"sync/atomic"
	"unsafe"
)

type testcfg struct {
	Test  string `json:"test"`
	Test2 int    `json:"test2"`
}
type mysqlcfg struct {
	A int `json:"a"`
}
type rediscfg struct {
	B bool `json:"b"`
}

var curtTestCfg *testcfg

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
	err := agollo.Init(agollo.WithConfFile("./apollo.json"),
		agollo.WithLogFunc(logger.Debugf, logger.Infof, logger.Errorf))
	if err != nil {
		fmt.Println(err)
		return
	}
	agollo.RegChangeEventHandler(HandleAll)

	fmt.Println("Initilization done")
	crt := agollo.GetConfigReader("test.json")
	crm := agollo.GetConfigReader("mysql.json")
	crr := agollo.GetConfigReader("redis.json")

	test, _ := crt.GetBytesValue("content")
	mysql, _ := crm.GetBytesValue("content")
	redis, _ := crr.GetBytesValue("content")

	var first testcfg
	var second mysqlcfg
	var third rediscfg
	json.Unmarshal(test, &first)
	json.Unmarshal(mysql, &second)
	json.Unmarshal(redis, &third)

	fmt.Printf("\n%v\n%v\n%v\n", first, second, third)

	go agollo.Start()

	fmt.Println("Wait for config change")
	var tmpChan = make(chan string)
	<-tmpChan
}

// HandleAll is a fake handler
func HandleAll(event *agollo.ChangeEvent) error {
	fmt.Println("Hi, this is a Callback!")
	bytes, _ := json.Marshal(event)
	fmt.Println("event:", string(bytes))

	switch event.Namespace {
	case "test.json":
		var newVal testcfg
		content := []byte(event.Changes[0].NewValue)

		if err := json.Unmarshal(content, &newVal); err != nil {
			return err
		}
		ptr := (*unsafe.Pointer)(unsafe.Pointer(&curtTestCfg))
		atomic.StorePointer(ptr, unsafe.Pointer(&newVal))
		fmt.Printf("updated config: %v\n", curtTestCfg)

	case "mysql.json":
		// do something
	case "redis.json":
		// do something
	default:
		// maybe do nothing
	}

	return nil
}
