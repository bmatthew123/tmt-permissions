package accessors

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	testhelpers "github.com/byu-oit-ssengineering/tmt-test-helpers"
)

func TestGetGroup(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	ga := NewGroupAccessor(db)

	expected := Group{"1", "1", "testGroup"}
	columns := []string{"guid", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT (.) FROM groups WHERE guid=(.)").
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("1,1,testGroup"))
	group, err := ga.Get("1")
	if err != nil {
		t.Error("An unexpected error occurred while getting a group %v", err)
	}

	if group != expected {
		t.Errorf("Expected %v but got %v", expected, group)
	}

	if err := ga.DB.Close(); err != nil {
		t.Errorf("An error occurred: %v", err)
	}
}

func TestGroupsGetByArea(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	ga := NewGroupAccessor(db)

	expected := []Group{Group{"1", "1", "testGroup"}, Group{"2", "1", "group2"}}
	columns := []string{"guid", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT (.) FROM groups WHERE area=(.)").
		WithArgs().
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("1,1,testGroup\n2,1,group2"))

	groups, err := ga.GetByArea("1")
	if err != nil {
		t.Error("An unexpected error occurred while getting a group %v", err)
	}

	for i := 0; i < len(groups); i++ {
		if groups[i] != expected[i] {
			t.Errorf("Expected %v but got %v", expected, groups)
		}
	}

	if err := ga.DB.Close(); err != nil {
		t.Errorf("An error occurred: %v", err)
	}
}

func TestInsertGroup(t *testing.T) {
	NewGuid = func() string {
		return "123def"
	}

	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	ga := NewGroupAccessor(db)

	sqlmock.ExpectPrepare()
	sqlmock.ExpectExec("INSERT INTO groups .+ VALUES .+").
		WithArgs("123def", "1", "testGroup").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = ga.Insert(Group{"", "1", "testGroup"})
	if err != nil {
		t.Error("An unexpected error occurred while getting a group:\n %s", err.Error())
	}

	if err := ga.DB.Close(); err != nil {
		t.Errorf("An error occurred: %v", err)
	}
}

func TestRenameGroup(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	ga := NewGroupAccessor(db)

	sqlmock.ExpectPrepare()
	sqlmock.ExpectExec("UPDATE groups SET name=(.) WHERE guid=(.)").
		WithArgs("changed", "1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = ga.Rename("1", "changed")
	if err != nil {
		t.Error("An unexpected error occurred while getting a group:\n %s", err.Error())
	}

	if err := ga.DB.Close(); err != nil {
		t.Errorf("An error occurred: %v", err)
	}
}

func TestDeleteGroup(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	ga := NewGroupAccessor(db)

	sqlmock.ExpectPrepare()
	sqlmock.ExpectExec("DELETE FROM groups WHERE guid=(.)").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = ga.Delete("1")
	if err != nil {
		t.Error("An unexpected error occurred while getting a group:\n %s", err.Error())
	}

	if err := ga.DB.Close(); err != nil {
		t.Errorf("An error occurred: %v", err)
	}
}

func TestGetImpliedGroups(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	ga := NewGroupAccessor(db)

	// query expectations
	columns := []string{"guid", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM groups WHERE area=.").
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("guid1,1,group1\nguid2,1,group2\nguid3,1,group3"))
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT groups.* FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=(.) AND groups.area=(.)").
		WithArgs("netId", "1").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("guid1,1,group1\nguid3,1,group3"))
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT groups.* FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=(.) AND groups.area=(.)").
		WithArgs("netId", "1").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("guid1,1,group1\nguid3,1,group3"))
	columns = []string{"actor", "verb", "resource"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM policy WHERE actor IN .+").
		WithArgs("1", "guid1", "guid3").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("group1,edit,res1\ngroup1,view,res1\ngroup3,update,res2"))
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM policy WHERE actor=.").
		WithArgs("guid2").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("group2,edit,res1\ngroup2,view,res1"))
	columns = []string{"guid", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM groups WHERE guid=.").
		WithArgs("guid2").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("guid2,1,group2"))

	implied, err := ga.GetImpliedGroups("netId", "1")
	if err != nil {
		t.Error("An unexpected error occurred while getting a group:\n %s", err.Error())
	}

	if err := ga.DB.Close(); err != nil {
		t.Errorf("An error occurred: %v", err)
	}

	expected := Group{"guid2", "1", "group2"}
	if len(implied) != 1 && implied[0] != expected {
		t.Errorf("Expected group2 but got %v", implied)
	}
}

func TestGetImpliedGroupsNone(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	ga := NewGroupAccessor(db)

	// query expectations
	columns := []string{"guid", "area", "name"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT . FROM groups WHERE area=.").
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("guid1,1,group1\nguid2,1,group2\nguid3,1,group3"))
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT groups.* FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=(.) AND groups.area=(.)").
		WithArgs("netId", "1").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("guid1,1,group1\nguid3,1,group3"))
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT groups.* FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=(.) AND groups.area=(.)").
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
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("group2,update,res1\ngroup2,view,res1"))

	implied, err := ga.GetImpliedGroups("netId", "1")
	if err != nil {
		t.Error("An unexpected error occurred while getting a group:\n %s", err.Error())
	}

	if err := ga.DB.Close(); err != nil {
		t.Errorf("An error occurred: %v", err)
	}

	if len(implied) != 0 {
		t.Errorf("Expected an empty array but got %v", implied)
	}
}
