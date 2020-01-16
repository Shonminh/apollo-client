package agollo

import (
	"fmt"
	"github.com/Shonminh/apollo-client/basetest"
	"github.com/Shonminh/apollo-client/internal/apollo"
	"github.com/Shonminh/apollo-client/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

const (
	TEST_DEFAULT_VAL_A1         string = "11"
	TEST_DEFAULT_VAL_A2         string = "22"
	TEST_DEFAULT_NAMESPACE_NAME string = "application"
	TEST_DEFAULT_APPID          string = "agollo_test"
	TEST_DEFAULT_CLUSTER        string = "default"
	TEST_DEFAULT_APOLLOADDR     string = "127.0.0.1:8000"
)

var gMockServer = mock.NewServer()

type AgolloSuite struct {
	basetest.BaseSuite
}

func TestAgollo(t *testing.T) {
	testSuite := new(AgolloSuite)
	testSuite.Register(
		func() {
			go gMockServer.Run()
		},
		func() {
			initFileConf(DEFAULT_NAMESPACENAME)
		},
		nil,
		func() {
			gMockServer.Stop()
			gMockServer.Wait()
		},
	)

	suite.Run(t, testSuite)
}

func (s *AgolloSuite) testLog(format string, v ...interface{}) {
	s.T().Logf(format, v...)
}

func (s *AgolloSuite) Test_Init_Success() {
	agOpts := []Option{
		WithBackupDir("./mock/tmp"),
		WithBackupSuffix(".json"),
		WithConfFile("./mock/tmp/apollo.json"),
		WithLogFunc(s.testLog, s.testLog, s.testLog),
	}

	err := Init(agOpts...)
	assert.Nil(s.T(), err)
	checkConfigVal(s, TEST_DEFAULT_VAL_A1, TEST_DEFAULT_VAL_A2)
}

func (s *AgolloSuite) Test_SyncConfig_Success() {
	releaseKey := "12345678901"
	service, err := getTestService(-1, releaseKey)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), service)

	err = service.syncConfig(false, nil)
	assert.Nil(s.T(), err)
	checkConfigVal(s, releaseKey, releaseKey)

}

func (s *AgolloSuite) Test_LoadConfigFile_Success() {
	releaseKey := "12345678901"
	service, err := getTestService(-1, releaseKey)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), service)

	err = service.LoadConfigFile(nil)
	assert.Nil(s.T(), err)
	checkConfigVal(s, TEST_DEFAULT_VAL_A1, TEST_DEFAULT_VAL_A2)
}

func (s *AgolloSuite) Test_pullNotify_NotModify() {
	releaseKey := "12345678901"
	service, err := getTestService(-1, releaseKey)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), service)

	service.pullNotify()
	checkEmptyConf(s)
}

func (s *AgolloSuite) Test_pullNotify_ModifySuccess() {
	releaseKey := "12345678901"
	service, err := getTestService(100, releaseKey)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), service)

	service.pullNotify()
	checkConfigVal(s, releaseKey, releaseKey)
}

func initFileConf(namespaceName string) {
	gOption = newDefaultOption()
	gOption.BackupDir = "./mock/tmp"
	gOption.BackupSuffix = ".json"
	path := getConfigPath(namespaceName)

	ac := apollo.Config{
		ConnConfig: apollo.ConnConfig{
			AppId:         TEST_DEFAULT_APPID,
			Cluster:       TEST_DEFAULT_CLUSTER,
			NamespaceName: namespaceName,
			ReleaseKey:    "",
		},
		Configurations: map[string]string{
			"a1": TEST_DEFAULT_VAL_A1,
			"a2": TEST_DEFAULT_VAL_A2,
		},
	}
	err := writeConfigFile(&ac, path)
	if err != nil {
		fmt.Printf("writeConfigFile fail:%s", err.Error())
	}
}

func getTestService(notificationId int64, releaseKey string) (*service, error) {
	option := newDefaultOption()
	option.confFile = "./mock/tmp/apollo.json"
	err := loadJsonConfig(option, option.confFile)
	if err != nil {
		return nil, errors.WithMessage(err, "loadJsonConfig")
	}

	initCache(option.ConfigCacheSize)
	apollo.NofityTimeout = option.notifyTimeout
	apollo.ConnectTimeout = option.connectTimeout
	apollo.RetryInterval = option.retryInterval
	s := &service{
		ConfigCenter: apollo.ConfigCenter{
			Host:     apollo.NewSingleHostResolver(option.ApolloAddr),
			AppId:    option.AppId,
			Cluster:  option.Cluster,
			ClientIp: option.clientIp,
		},
	}
	nl := strings.Split(option.NamespaceName, ",")
	s.namespaceList = make([]*namespace, len(nl))
	for i, v := range nl {
		s.namespaceList[i] = &namespace{
			NamespaceName:  v,
			releaseKey:     releaseKey,
			NotificationId: notificationId,
		}
	}

	return s, nil
}

func checkConfigVal(s *AgolloSuite, expectedA1, expectedA2 string) {
	cr := GetConfigReader("application")
	a1, err := cr.GetStringValue("a1")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), expectedA1, a1)

	a2, err := cr.GetStringValue("a2")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), expectedA2, a2)
}

func checkEmptyConf(s *AgolloSuite) {
	cr := GetConfigReader("application")
	a1, err := cr.GetStringValue("a1")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "", a1)

	a2, err := cr.GetStringValue("a2")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "", a2)
}
