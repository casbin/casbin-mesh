package adapter

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

const testDB = "test.db"

type AdapterTestSuite struct {
	suite.Suite
	db       *BoltStore
	enforcer casbin.IEnforcer
}

func testGetPolicy(t *testing.T, e casbin.IEnforcer, wanted [][]string) {
	t.Helper()
	got := e.GetPolicy()
	if !util.Array2DEquals(wanted, got) {
		t.Error("got policy: ", got, ", wanted policy: ", wanted)
	}
}

func (suite *AdapterTestSuite) SetupTest() {
	t := suite.T()

	db, err := NewBoltStore(testDB)
	if err != nil {
		t.Fatalf("error opening bolt db: %s\n", err.Error())
	}
	suite.db = db

	bts, err := ioutil.ReadFile("../../test/test_data/rbac_policy.csv")
	if err != nil {
		t.Error(err)
	}

	a, err := NewAdapter(db, "casbin", string(bts))
	if err != nil {
		t.Error(err)
	}

	enforcer, err := casbin.NewEnforcer("../../test/test_data/rbac_model.conf", a)
	if err != nil {
		t.Errorf("error creating enforcer: %s\n", err.Error())
	}

	suite.enforcer = enforcer
}

func (suite *AdapterTestSuite) TearDownTest() {
	if _, err := os.Stat(testDB); err == nil {
		os.Remove(testDB)
	}
}

func Test_AdapterTest_Suite(t *testing.T) {
	suite.Run(t, new(AdapterTestSuite))
}

func (suite *AdapterTestSuite) Test_LoadBuiltinPolicy() {
	testGetPolicy(suite.T(), suite.enforcer, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}

func (suite *AdapterTestSuite) Test_SavePolicy_ReturnsErr() {
	e := suite.enforcer
	t := suite.T()

	err := e.SavePolicy()
	assert.EqualError(t, err, "not supported: must use auto-save with this adapter")
}

func (suite *AdapterTestSuite) Test_AutoSavePolicy() {
	e := suite.enforcer
	t := suite.T()

	e.EnableAutoSave(true)

	e.AddPolicy("roger", "data1", "write")
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}, {"roger", "data1", "write"}})

	e.RemovePolicy("roger", "data1", "write")
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	e.AddPolicies([][]string{{"roger", "data1", "read"}, {"roger", "data1", "write"}})
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}, {"roger", "data1", "read"}, {"roger", "data1", "write"}})

	//e.RemoveFilteredPolicy(0, "roger")
	//e.LoadPolicy()
	//testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
	//
	//_, err := e.RemoveFilteredPolicy(1, "data1")
	//assert.EqualError(t, err, "fieldIndex != 0: adapter only supports filter by prefix")

	e.AddPolicies([][]string{{"roger", "data1", "read"}, {"roger", "data1", "write"}})
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}, {"roger", "data1", "read"}, {"roger", "data1", "write"}})

	e.RemovePolicies([][]string{{"roger", "data1", "read"}, {"roger", "data1", "write"}})
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}
