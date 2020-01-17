package agollo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_GetConfigReader_ExpectedTestConfigReader_ReturnEqual(t *testing.T) {
	backupGlobals()
	err := Init(WithConfFile("examples/json/apollo.json"))
	assert.Nil(t, err, "should be nil")
	cr := GetConfigReader("test.json")
	var expectedCr configReader = "test.json"
	if cr != expectedCr {
		t.Errorf("configreader expected %v, got %v ", expectedCr, cr)
	}
	recoverGlobals()
}

func Test_WithConfFile_LoadConfFileValue_ReturnEqualValue(t *testing.T) {
	backupGlobals()
	err := Init(WithConfFile("examples/json/apollo.json"))
	assert.Nil(t, err, "should be nil")
	if gOption.confFile != "examples/json/apollo.json" {
		t.Errorf("ConfFile expected %s, got %v ", "examples/json/apollo.json", gOption.confFile)
	}
	if gOption.AppId != "hello3" {
		t.Errorf("AppId expected %s, got %v ", "hello3", gOption.AppId)
	}
	if gOption.Cluster != "default" {
		t.Errorf("Cluster expected %s, got %v ", "default", gOption.Cluster)
	}
	if gOption.NamespaceName != "application,test.json,mysql.json,redis.json" {
		t.Errorf("NamespaceName expected %s, got %v ", "application,test.json,mysql.json,redis.json", gOption.NamespaceName)
	}
	if gOption.ApolloAddr != "127.0.0.1:8000" {
		t.Errorf("ApolloAddr expected %s, got %v ", "127.0.0.1:8000", gOption.ApolloAddr)
	}
	recoverGlobals()
}

func Test_WithQuickInitWithBK_QuickInitWithBK_ReturnTure(t *testing.T) {
	backupGlobals()
	err := Init(WithConfFile("examples/json/apollo.json"), WithQuickInitWithBK())
	assert.Nil(t, err, "should be nil")
	if gOption.quickInitWithBK != true {
		t.Errorf("QuickInitWithBK expected false, got %v ", gOption.quickInitWithBK)
	}
	recoverGlobals()
}

func Test_WithCluster_ClusterSGCoverDefault_ReturnSG(t *testing.T) {
	backupGlobals()
	err := Init(WithConfFile("examples/json/apollo.json"), WithCluster("SG"), WithQuickInitWithBK())
	assert.Nil(t, err, "should be nil")
	if gOption.Cluster != "SG" {
		t.Errorf("Cluster expected SG, got %v ", gOption.Cluster)
	}
	recoverGlobals()
}
