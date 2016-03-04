package apis

import (
	"database/sql"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	eden "github.com/byu-oit-ssengineering/tmt-eden"
	accessors "github.com/byu-oit-ssengineering/tmt-permissions/accessors"
	testhelpers "github.com/byu-oit-ssengineering/tmt-test-helpers"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"testing"
)

type testGroupResponse struct {
	Status string
	Data   accessors.Group
}

type testGroupArrayResponse struct {
	Status string
	Data   []accessors.Group
}

func TestGetGroup(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred instantiating accessor")
		return
	}
	api := &Api{db}

	expected := accessors.Group{"1", "1", "testGroup"}
	columns := []string{"id", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT (.) FROM groups WHERE guid=(.)").
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("1,1,testGroup"))

	// Create context, call API
	var result []byte
	var output testGroupResponse
	c := testhelpers.NewTestingContext("", httprouter.Params{httprouter.Param{Key: "guid", Value: "1"}}, api.GetGroup)
	testhelpers.CallAPI(api.GetGroup, c, &result)

	// Parse output
	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Compare output to expected output
	if output.Data != expected {
		t.Errorf("Expected: %v, but got %v", expected, output.Data)
	}
}

func TestGetGroupsByArea(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred instantiating accessor")
		return
	}
	api := &Api{db}

	expected := []accessors.Group{accessors.Group{"1", "1", "testGroup"}, accessors.Group{"2", "1", "group2"}}
	columns := []string{"guid", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT (.) FROM groups WHERE area=(.)").
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("1,1,testGroup\n2,1,group2"))

	// Create context, call API
	var result []byte
	var output testGroupArrayResponse
	c := testhelpers.NewTestingContext("area=1", nil, api.GetGroup)
	testhelpers.CallAPI(api.GetGroupsByArea, c, &result)

	// Parse output
	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Compare output to expected output
	for i := 0; i < len(expected); i++ {
		if output.Data[i] != expected[i] {
			t.Errorf("Expected: %v, but got %v instead", expected, output)
		}
	}
}

func TestInsertGroup(t *testing.T) {
	accessors.Log = func(t, a, d string, logToStdErr bool, db *sql.DB) (bool, error) {
		return true, nil
	}
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred instantiating accessor")
		return
	}
	accessors.NewGuid = func() string {
		return "123def"
	}
	api := &Api{db}

	sqlmock.ExpectPrepare()
	sqlmock.ExpectExec("INSERT INTO groups .+ VALUES .+").
		WithArgs("123def", "1", "testGroup").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create context and call API
	var result []byte
	var output eden.Response
	c := testhelpers.NewTestingContext("area=1&name=testGroup", nil, api.CreateGroup)
	testhelpers.CallAPI(api.CreateGroup, c, &result)

	// Parse output
	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Ensure correct output
	if output.Data != "success" {
		t.Errorf("expected to get 'success' but got %v instead", output)
	}
}

func TestRenameGroup(t *testing.T) {
	accessors.Log = func(t, a, d string, logToStdErr bool, db *sql.DB) (bool, error) {
		return true, nil
	}
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred instantiating accessor")
		return
	}
	api := &Api{db}

	sqlmock.ExpectPrepare()
	sqlmock.ExpectExec("UPDATE groups SET name=(.) WHERE guid=(.)").
		WithArgs("changed", "1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create context and call API
	var result []byte
	var output eden.Response
	c := testhelpers.NewTestingContext("name=changed", httprouter.Params{httprouter.Param{Key: "id", Value: "1"}}, api.RenameGroup)
	testhelpers.CallAPI(api.RenameGroup, c, &result)

	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Ensure correct output
	if output.Data != "success" {
		t.Errorf("expected to get 'success' but got %v instead", output)
	}
}

func TestDeleteGroup(t *testing.T) {
	accessors.Log = func(t, a, d string, logToStdErr bool, db *sql.DB) (bool, error) {
		return true, nil
	}
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred instantiating accessor")
		return
	}
	api := &Api{db}

	sqlmock.ExpectPrepare()
	sqlmock.ExpectExec("DELETE FROM groups WHERE guid=(.)").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create context and call API
	var result []byte
	var output eden.Response
	c := testhelpers.NewTestingContext("", httprouter.Params{httprouter.Param{Key: "id", Value: "1"}}, api.DeleteGroup)
	testhelpers.CallAPI(api.DeleteGroup, c, &result)

	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Ensure correct output
	if output.Data != "success" {
		t.Errorf("expected to get 'success' but got %v instead", output)
	}
}

func TestGetUserGroups(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred instantiating accessor")
		return
	}
	api := &Api{db}

	expected := accessors.Group{"g1", "area", "n1"}
	columns := []string{"guid", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT groups.. FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=(.) AND groups.area=(.)").
		WithArgs("netId", "area").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("g1,area,n1"))

	// Create context, call API
	var result []byte
	var output testGroupArrayResponse
	c := testhelpers.NewTestingContext("netId=netId&area=area", nil, api.GetUserGroups)
	testhelpers.CallAPI(api.GetUserGroups, c, &result)

	// Parse output
	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Compare output to expected output
	if output.Data[0] != expected {
		t.Errorf("Expected: %v, but got %v", expected, output.Data)
	}
}

func TestGetUserGroupsImplied(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred instantiating accessor")
		return
	}
	api := &Api{db}

	expected := []accessors.Group{accessors.Group{"guid1", "area", "group1"}, accessors.Group{"guid3", "area", "group3"}, accessors.Group{"guid2", "area", "group2"}}
	// query expectations
	columns := []string{"guid", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT groups.. FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=(.) AND groups.area=(.)").
		WithArgs("netId", "1").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("guid1,area,group1\nguid3,area,group3"))

	// for implied
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM groups WHERE area=.").
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("guid1,1,group1\nguid2,1,group2\nguid3,1,group3"))
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT groups.. FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=(.) AND groups.area=(.)").
		WithArgs("netId", "1").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("guid1,1,group1\nguid3,1,group3"))
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT groups.. FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=(.) AND groups.area=(.)").
		WithArgs("netId", "1").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("guid1,1,group1\nguid3,1,group3"))
	columns = []string{"actor", "verb", "resource"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM policy WHERE actor IN .+").
		WithArgs("1", "guid1", "guid3").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("group1,edit,res1\ngroup1,view,res1\ngroup3,update,res2"))
	columns = []string{"actor", "verb", "resource"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM policy WHERE actor=.").
		WithArgs("guid2").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("group2,edit,res1\ngroup2,view,res1"))
	columns = []string{"guid", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM groups WHERE guid=.").
		WithArgs("guid2").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("guid2,area,group2"))

	// Create context, call API
	var result []byte
	var output testGroupArrayResponse
	c := testhelpers.NewTestingContext("implied=true&netId=netId&area=1", nil, api.GetUserGroups)
	testhelpers.CallAPI(api.GetUserGroups, c, &result)

	// Parse output
	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Compare output to expected output
	for i := 0; i < len(expected); i++ {
		if output.Data[i] != expected[i] {
			t.Errorf("Expected: %v, but got %v", expected, output.Data)
		}
	}
}
