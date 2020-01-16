package basetest

import (
	"github.com/stretchr/testify/suite"
)

type BaseSuite struct {
	suite.Suite
	setupSuite, setupTest, tearDownTest, tearDownSuite func()
	hasErr                                             bool
}

func (base *BaseSuite) Register(setupSuite, setupTest, tearDownTest, tearDownSuite func()) {
	base.setupSuite = setupSuite
	base.setupTest = setupTest
	base.tearDownSuite = tearDownSuite
	base.tearDownTest = tearDownTest
}

func (base *BaseSuite) SetupSuite() {
	base.hasErr = false
	if base.setupSuite != nil {
		base.setupSuite()
	}
}

func (base *BaseSuite) SetupTest() {

	if base.setupTest != nil {
		base.setupTest()
	}
}

func (base *BaseSuite) TearDownTest() {
	if base.tearDownTest != nil {
		base.tearDownTest()
	}
}

func (base *BaseSuite) TearDownSuite() {
	if base.tearDownSuite != nil {
		base.tearDownSuite()
	}
}
