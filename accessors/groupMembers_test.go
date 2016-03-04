package accessors

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	testhelpers "github.com/byu-oit-ssengineering/tmt-test-helpers"
)

func TestGetGroupMembers(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	ma := NewMembersAccessor(db)

	expected := []string{"netId", "someone"}
	columns := []string{"netId"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT netId FROM groupMembers WHERE groupGuid=(.)").
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("netId\nsomeone"))
	members, err := ma.GetGroupMembers("1")
	if err != nil {
		t.Error("An unexpected error occurred while getting a group %v", err)
	}

	for i := 0; i < len(members); i++ {
		if members[i] != expected[i] {
			t.Errorf("Expected %v but got %v", expected, members)
			t.FailNow()
		}
	}

	if err := ma.DB.Close(); err != nil {
		t.Errorf("An error occurred: %v", err)
	}
}

func TestGetUserGroups(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	ma := NewMembersAccessor(db)

	expected := []string{"1", "2"}
	columns := []string{"groupId"}
	sqlmock.ExpectPrepare()
	sqlmock.ExpectQuery("SELECT groupGuid FROM groupMembers WHERE netId=(.)").
		WithArgs("netId").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("1\n2"))

	groups, err := ma.GetUserGroups("netId")
	if err != nil {
		t.Error("An unexpected error occurred while getting a group %v", err)
	}

	for i := 0; i < len(groups); i++ {
		if groups[i] != expected[i] {
			t.Errorf("Expected %v but got %v", expected, groups)
			t.FailNow()
		}
	}

	if err := ma.DB.Close(); err != nil {
		t.Errorf("An error occurred: %v", err)
	}
}

func TestAddToGroup(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	ma := NewMembersAccessor(db)

	sqlmock.ExpectPrepare()
	sqlmock.ExpectExec("INSERT INTO groupMembers .+ VALUES .+").
		WithArgs("netId", "1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = ma.AddToGroup("netId", "1")
	if err != nil {
		t.Error("An unexpected error occurred while getting a group:\n %s", err.Error())
	}

	if err := ma.DB.Close(); err != nil {
		t.Errorf("An error occurred: %v", err)
	}
}

func TestRemoveFromGroup(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a group accessor %v", err)
		return
	}

	ma := NewMembersAccessor(db)

	sqlmock.ExpectPrepare()
	sqlmock.ExpectExec("DELETE FROM groupMembers WHERE netId=(.) AND groupGuid=(.)").
		WithArgs("netId", "1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = ma.RemoveFromGroup("netId", "1")
	if err != nil {
		t.Error("An unexpected error occurred while getting a group:\n %s", err.Error())
	}

	if err := ma.DB.Close(); err != nil {
		t.Errorf("An error occurred: %v", err)
	}
}
