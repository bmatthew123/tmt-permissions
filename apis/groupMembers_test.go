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

type testStringArrayResponse struct {
	Status string
	Data   []string
}

func TestGetGroupsByNetId(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred instantiating accessor")
		return
	}
	api := &Api{db}

	expected := []string{"1", "2"}
	columns := []string{"groupGuid"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT groupGuid FROM groupMembers WHERE netId=(.)").
		WithArgs("someone").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("1\n2"))

	// Create context, call API
	var result []byte
	var output testStringArrayResponse
	c := testhelpers.NewTestingContext("netId=someone", nil, api.GetGroupsByNetId)
	testhelpers.CallAPI(api.GetGroupsByNetId, c, &result)

	// Parse output
	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Compare output to expected output
	for i := 0; i < len(output.Data); i++ {
		if output.Data[i] != expected[i] {
			t.Errorf("Expected: %v, but got %v", expected, output.Data)
			t.FailNow()
		}
	}
}

func TestGetGroupMembers(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred instantiating accessor")
		return
	}
	api := &Api{db}

	expected := []string{"netId", "someone"}
	columns := []string{"netId"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT netId FROM groupMembers WHERE groupGuid=(.)").
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("netId\nsomeone"))

	// Create context, call API
	var result []byte
	var output testStringArrayResponse
	c := testhelpers.NewTestingContext("", httprouter.Params{httprouter.Param{Key: "groupId", Value: "1"}}, api.GetGroupMembers)
	testhelpers.CallAPI(api.GetGroupMembers, c, &result)

	// Parse output
	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}

	// Compare output to expected output
	for i := 0; i < len(output.Data); i++ {
		if output.Data[i] != expected[i] {
			t.Errorf("Expected: %v, but got %v instead", expected, output)
		}
	}
}

func TestAddGroupMember(t *testing.T) {
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
	sqlmock.ExpectExec("INSERT INTO groupMembers .+ VALUES .+").
		WithArgs("netId", "1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create context and call API
	var result []byte
	var output eden.Response
	c := testhelpers.NewTestingContext("netId=netId&group=1", nil, api.AddGroupMember)
	testhelpers.CallAPI(api.AddGroupMember, c, &result)

	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}

	// Ensure correct output
	if output.Data != "success" {
		t.Errorf("expected to get 'success' but got %v instead", output.Data)
	}
}

func TestRemoveGroupMember(t *testing.T) {
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
	sqlmock.ExpectExec("DELETE FROM groupMembers WHERE netId=(.) AND groupGuid=(.)").
		WithArgs("netId", "1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create context and call API
	var result []byte
	var output eden.Response
	c := testhelpers.NewTestingContext("", httprouter.Params{httprouter.Param{Key: "netId", Value: "netId"}, httprouter.Param{Key: "groupGuid", Value: "1"}}, api.RemoveGroupMember)
	testhelpers.CallAPI(api.RemoveGroupMember, c, &result)

	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}

	// Ensure correct output
	if output.Data != "success" {
		t.Errorf("expected to get 'success' but got %v instead", output.Data)
	}
}
