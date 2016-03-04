package accessors

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	testhelpers "github.com/byu-oit-ssengineering/tmt-test-helpers"
)

func TestCheckPermission(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	pa := NewPermissionAccessor(db)

	columns := []string{"guid"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM policy WHERE verb=. AND resource=. AND actor IN (.+)").
		WithArgs("edit", "11111111-2222-3333-2222-111111111111", "11111111-2222-3333-4444-555555555555", "11111111-1111-2222-2222-333333333333").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("44444444-3333-2222-1111-000000000000\n55555555-6666-7777-8888-999999999999"))
	perm, err := pa.CheckPermission([]string{"11111111-2222-3333-4444-555555555555", "11111111-1111-2222-2222-333333333333"}, "11111111-2222-3333-2222-111111111111", "edit")
	if err != nil {
		t.Error("An unexpected error occurred while getting a group %v", err)
	}

	if !perm {
		t.Error("Expected true but got false")
	}

	if err := pa.DB.Close(); err != nil {
		t.Errorf("An error occurred: %v", err)
	}
}

func TestCheckNoPermission(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	pa := NewPermissionAccessor(db)

	columns := []string{"guid"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM policy WHERE verb=. AND resource=. AND actor IN (.+)").
		WithArgs("edit", "11111111-2222-3333-2222-111111111111", "11111111-2222-3333-4444-555555555555", "11111111-1111-2222-2222-333333333333").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString(""))
	perm, err := pa.CheckPermission([]string{"11111111-2222-3333-4444-555555555555", "11111111-1111-2222-2222-333333333333"}, "11111111-2222-3333-2222-111111111111", "edit")
	if err != nil {
		t.Error("An unexpected error occurred while getting a group %v", err)
	}

	if perm {
		t.Error("Expected false but got true")
	}

	if err := pa.DB.Close(); err != nil {
		t.Errorf("An error occurred: %v", err)
	}
}

func TestAddPermission(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	pa := NewPermissionAccessor(db)

	sqlmock.ExpectPrepare()
	sqlmock.ExpectExec("INSERT INTO policy (.+) VALUES (.+)").
		WithArgs("11111111-2222-3333-2222-111111111111", "edit", "11111111-2222-3333-4444-555555555555").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = pa.Add("11111111-2222-3333-2222-111111111111", "edit", "11111111-2222-3333-4444-555555555555")
	if err != nil {
		t.Error("An unexpected error occurred while getting a resource:\n %s", err.Error())
	}

	if err := pa.DB.Close(); err != nil {
		t.Errorf("An error occurred: %v", err)
	}
}

func TestDeletePermission(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	pa := NewPermissionAccessor(db)

	sqlmock.ExpectPrepare()
	sqlmock.ExpectExec("DELETE FROM policy WHERE actor=. AND verb=. AND resource=.").
		WithArgs("11111111-2222-3333-2222-111111111111", "edit", "88888888-8888-8888-8888-888888888888").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = pa.Delete("11111111-2222-3333-2222-111111111111", "edit", "88888888-8888-8888-8888-888888888888")
	if err != nil {
		t.Error("An unexpected error occurred while getting a resource:\n %s", err.Error())
	}

	if err := pa.DB.Close(); err != nil {
		t.Errorf("An error occurred: %v", err)
	}
}

func TestGetAllGroupsInArea(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	pa := NewPermissionAccessor(db)

	//expected := []string{"11111111-1111-1111-1111-111111111111", "11111111-1111-1111-1111-111111111112"}
	expected := []Group{Group{"guid1", "area1", "name1"}, Group{"guid2", "area2", "name2"}}
	columns := []string{"guid", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT groups.* FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=(.) AND groups.area=(.)").
		WithArgs("00000000-0000-0000-0000-000000000000", "1").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("guid1,area1,name1\nguid2,area2,name2"))

	actors, err := pa.Get("1", "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Errorf("An unexpected error occurred while getting a group %v", err)
	}

	for i := 0; i < len(actors); i++ {
		if actors[i] != expected[i] {
			t.Errorf("Expected %v but got %v", expected, actors)
		}
	}
	if err := pa.DB.Close(); err != nil {
		t.Errorf("An error occurred: %v", err)
	}

}

func TestGetPermissionsByGroup(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	pa := NewPermissionAccessor(db)

	expected := []Permission{Permission{"actor", "verb", "resource"}, Permission{"actor", "verb1", "resource1"}}
	columns := []string{"actor", "verb", "resource"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM policy WHERE actor=.").
		WithArgs("actor").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("actor,verb,resource\nactor,verb1,resource1"))

	permissions, err := pa.GetGroupPermissions("actor")
	if err != nil {
		t.Errorf("An unexpected error occurred while getting a group %v", err)
	}

	for i := 0; i < len(permissions); i++ {
		if permissions[i] != expected[i] {
			t.Errorf("Expected %v but got %v", expected, permissions)
		}
	}
	if err := pa.DB.Close(); err != nil {
		t.Errorf("An error occurred: %v", err)
	}

}

func TestGetUserPermissions(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	pa := NewPermissionAccessor(db)

	// Get groups
	columns := []string{"guid"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT guid FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=(.) AND groups.area=(.)").
		WithArgs("netId", "area").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("g1\ng2"))

	// Get permissions
	columns = []string{"actor", "verb", "resource"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM policy WHERE actor IN .+").
		WithArgs("area", "g1", "g2").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("g1,edit,resource\ng1,read,resource\ng2,edit,resource\ng2,read,resource2"))

	perms, err := pa.GetUserPermissions("netId", "area")
	expected := []Permission{
		Permission{"", "edit", "resource"},
		Permission{"", "read", "resource"},
		Permission{"", "read", "resource2"},
	}
	for i := 0; i < len(perms); i++ {
		if perms[i] != expected[i] {
			t.Errorf("Expected %v but got %v", expected, perms)
		}
	}
}
