package apis

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	eden "github.com/byu-oit-ssengineering/tmt-eden"
	accessors "github.com/byu-oit-ssengineering/tmt-permissions/accessors"
	testhelpers "github.com/byu-oit-ssengineering/tmt-test-helpers"
	"github.com/julienschmidt/httprouter"
)

type testPermission struct {
	Actor    string
	Verb     string
	Resource string
}

type testPermissionResponse struct {
	Status string
	Data   []testPermission
}

func TestCheckPermission(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred instantiating accessor")
		return
	}
	api := &Api{db}

	expected := true

	columns := []string{"guid", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=. AND groups.area=.").
		WithArgs("E", "A").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("x,1,n1\ny,1,n2\nz,1,n3"))

	columns = []string{"actor", "verb", "resource"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM policy WHERE verb=(.) AND resource=(.) AND actor IN (.+)").
		WithArgs("edit", "1", "x", "y", "z", "A").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("x,edit,1\ny,edit,1\nz,edit,1"))

	// Create context, call API
	var result []byte
	var output eden.Response
	c := testhelpers.NewTestingContext("areaGuid=A&employeeGuid=E&verb=edit&resource=1", nil, api.CheckPermission)
	testhelpers.CallAPI(api.CheckPermission, c, &result)

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

func TestGetGroupsByVerb(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred instantiating accessor")
		return
	}
	api := &Api{db}

	expected := []accessors.Group{accessors.Group{"1", "1", "testGroup"}}
	columns := []string{"guid", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=. AND groups.area=.").
		WithArgs("E", "A").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString(""))

	columns = []string{"actor", "verb", "resource"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM policy WHERE verb=. AND resource=. AND actor IN (.+)").
		WithArgs("poot", "3", "1").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("1,poot,3"))

	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM policy WHERE verb=. AND resource=. AND actor IN (.+)").
		WithArgs("poot", "3", "2").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString(""))

	// Create context, call API
	var result []byte
	var output testGroupArrayResponse
	c := testhelpers.NewTestingContext("", httprouter.Params{httprouter.Param{Key: "resourceGuid", Value: "3"}, httprouter.Param{Key: "verb", Value: "poot"}}, api.GetGroupsByVerb)
	testhelpers.CallAPI(api.GetGroupsByVerb, c, &result)

	// Parse output
	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Compare output to expected output
	for i := 0; i < len(output.Data); i++ {
		if output.Data[i] != expected[i] {
			t.Errorf("Expected: %v, but got %v", expected, output.Data)
		}
	}
}

func TestAddPermission(t *testing.T) {
	accessors.Log = func(t, a, d string, logToStdErr bool, db *sql.DB) (bool, error) {
		return true, nil
	}
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred instantiating accessor")
		return
	}
	api := &Api{db}

	columns := []string{"guid", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=. AND groups.area=.").
		WithArgs("guid", "area").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("x,1,n1\ny,1,n2\nz,1,n3"))

	columns = []string{"actor", "verb", "resource"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM policy WHERE verb=(.) AND resource=(.) AND actor IN (.+)").
		WithArgs("edit", "1", "x", "y", "z", "area").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("x,edit,1\ny,edit,1\nz,edit,1"))

	sqlmock.ExpectPrepare()
	sqlmock.ExpectExec("INSERT INTO policy (.+) VALUES (.+)").
		WithArgs("2", "edit", "1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create context, call API
	var result []byte
	var output eden.Response
	c := testhelpers.NewTestingContext("actor=2&verb=edit&resource=1", nil, api.AddPermission)
	c.User = eden.User{"guid", "area"}
	testhelpers.CallAPI(api.AddPermission, c, &result)

	// Parse output
	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Compare output to expected output
	if output.Data != "success" {
		t.Errorf("Expected: %v, but got %v", "success", output)
	}
}

func TestAddWithoutPermission(t *testing.T) {
	accessors.Log = func(t, a, d string, logToStdErr bool, db *sql.DB) (bool, error) {
		return true, nil
	}
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred instantiating accessor")
		return
	}
	api := &Api{db}

	columns := []string{"guid", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=. AND groups.area=.").
		WithArgs("guid", "area").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("x,1,n1\ny,1,n2\nz,1,n3"))

	columns = []string{"actor", "verb", "resource"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM policy WHERE verb=(.) AND resource=(.) AND actor IN (.+)").
		WithArgs("edit", "1", "x", "y", "z", "area").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString(""))

	// Create context, call API
	var result []byte
	var output eden.Response
	c := testhelpers.NewTestingContext("actor=2&verb=edit&resource=1", nil, api.AddPermission)
	c.User = eden.User{"guid", "area"}
	testhelpers.CallAPI(api.AddPermission, c, &result)

	// Parse output
	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Compare output to expected output
	if output.Status != "FAILURE" {
		t.Errorf("Expected: %v, but got %v", "success", output)
	}
}

func TestDeletePermission(t *testing.T) {
	accessors.Log = func(t, a, d string, logToStdErr bool, db *sql.DB) (bool, error) {
		return true, nil
	}
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred instantiating accessor")
		return
	}
	api := &Api{db}

	columns := []string{"guid", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=. AND groups.area=.").
		WithArgs("guid", "area").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("x,1,n1\ny,1,n2\nz,1,n3"))

	columns = []string{"actor", "verb", "resource"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM policy WHERE verb=(.) AND resource=(.) AND actor IN (.+)").
		WithArgs("edit", "1", "x", "y", "z", "area").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("x,edit,1\ny,edit,1\nz,edit,1"))

	sqlmock.ExpectPrepare()
	sqlmock.ExpectExec("DELETE FROM policy WHERE actor=(.) AND verb=(.) AND resource=(.)").
		WithArgs("2", "edit", "1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create context, call API
	var result []byte
	var output eden.Response
	c := testhelpers.NewTestingContext("",
		httprouter.Params{
			httprouter.Param{Key: "actor", Value: "2"},
			httprouter.Param{Key: "verb", Value: "edit"},
			httprouter.Param{Key: "resource", Value: "1"}},
		api.DeletePermission,
	)
	c.User = eden.User{"guid", "area"}
	testhelpers.CallAPI(api.DeletePermission, c, &result)

	// Parse output
	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Compare output to expected output
	if output.Data != "success" {
		t.Errorf("Expected: %v, but got %v", "success", output.Data)
	}
}

func TestDeleteWithoutPermission(t *testing.T) {
	accessors.Log = func(t, a, d string, logToStdErr bool, db *sql.DB) (bool, error) {
		return true, nil
	}
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred instantiating accessor")
		return
	}
	api := &Api{db}

	columns := []string{"guid", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=. AND groups.area=.").
		WithArgs("guid", "area").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("x,1,n1\ny,1,n2\nz,1,n3"))

	columns = []string{"actor", "verb", "resource"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM policy WHERE verb=(.) AND resource=(.) AND actor IN (.+)").
		WithArgs("edit", "1", "x", "y", "z", "area").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString(""))

	// Create context, call API
	var result []byte
	var output eden.Response
	c := testhelpers.NewTestingContext("",
		httprouter.Params{
			httprouter.Param{Key: "actor", Value: "2"},
			httprouter.Param{Key: "verb", Value: "edit"},
			httprouter.Param{Key: "resource", Value: "1"}},
		api.DeletePermission,
	)
	c.User = eden.User{"guid", "area"}
	testhelpers.CallAPI(api.DeletePermission, c, &result)

	// Parse output
	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Compare output to expected output
	if output.Status != "FAILURE" {
		t.Errorf("Expected: %v, but got %v", "success", output.Data)
	}
}

func TestGetPermissionsByGroup(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred instantiating accessor")
		return
	}
	api := &Api{db}

	expected := []testPermission{testPermission{"actor", "verb", "resource"}, testPermission{"actor", "verb1", "resource1"}}
	columns := []string{"actor", "verb", "resource"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM policy WHERE actor=.").
		WithArgs("actor").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("actor,verb,resource\nactor,verb1,resource1"))

	// Create context, call API
	var result []byte
	var output testPermissionResponse
	c := testhelpers.NewTestingContext("", httprouter.Params{httprouter.Param{Key: "actor", Value: "actor"}},
		api.GetGroupPermissions,
	)
	testhelpers.CallAPI(api.GetGroupPermissions, c, &result)

	// Parse output
	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Compare output to expected output
	for i := 0; i < len(output.Data); i++ {
		if output.Data[i] != expected[i] {
			t.Errorf("Expected: %v, but got %v", expected, output.Data)
		}
	}
}

func TestGetUserPermissions(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred instantiating accessor")
		return
	}
	api := &Api{db}

	// Get groups
	columns := []string{"guid", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=(.) AND groups.area=(.)").
		WithArgs("netId", "area").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("g1,1,n1\ng2,1,n2"))

	// Get permissions
	columns = []string{"actor", "verb", "resource"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM policy WHERE actor IN .+").
		WithArgs("area", "g1", "g2").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("g1,edit,resource\ng1,read,resource\ng2,edit,resource\ng2,read,resource2"))

	expected := []testPermission{
		testPermission{"", "edit", "resource"},
		testPermission{"", "read", "resource"},
		testPermission{"", "read", "resource2"},
	}

	// Create context, call API
	var result []byte
	var output testPermissionResponse
	c := testhelpers.NewTestingContext("", httprouter.Params{httprouter.Param{Key: "netId", Value: "netId"}, httprouter.Param{Key: "area", Value: "area"}},
		api.GetUserPermissions,
	)
	testhelpers.CallAPI(api.GetUserPermissions, c, &result)

	// Parse output
	err = json.Unmarshal(result, &output)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Compare output to expected output
	for i := 0; i < len(output.Data); i++ {
		if output.Data[i] != expected[i] {
			t.Errorf("Expected: %v, but got %v", expected, output.Data)
		}
	}
}
