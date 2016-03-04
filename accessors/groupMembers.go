package accessors

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type MembersAccessor struct {
	DB *sql.DB // Database connection
}

// Returns a new group member accessor.
func NewMembersAccessor(db *sql.DB) *MembersAccessor {
	return &MembersAccessor{db}
}

// Gets all members of a group.
func (ga *MembersAccessor) GetGroupMembers(group string) ([]string, error) {
	members := make([]string, 0)
	stmt, err := ga.DB.Prepare("SELECT netId FROM groupMembers WHERE groupGuid=?")
	if err != nil {
		return members, err
	}

	rows, err := stmt.Query(group)
	if err != nil {
		return members, err
	}

	defer rows.Close()
	for rows.Next() {
		var user string
		rows.Scan(&user)
		members = append(members, user)
	}

	return members, nil
}

// Gets a list of all the groups a user belongs to.
func (ga *MembersAccessor) GetUserGroups(netId string) ([]string, error) {
	groups := make([]string, 0)
	stmt, err := ga.DB.Prepare("SELECT groupGuid FROM groupMembers WHERE netId=?")
	if err != nil {
		return groups, err
	}

	rows, err := stmt.Query(netId)
	if err != nil {
		return groups, err
	}

	defer rows.Close()
	for rows.Next() {
		var group string
		rows.Scan(&group)
		groups = append(groups, group)
	}
	return groups, err
}

// Add a user to a group.
func (ga *MembersAccessor) AddToGroup(netId, group string) error {
	stmt, err := ga.DB.Prepare("INSERT INTO groupMembers (netId, groupGuid) VALUES (?,?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(netId, group)
	return err
}

// Remove a user from a group.
func (ga *MembersAccessor) RemoveFromGroup(netId, group string) error {
	stmt, err := ga.DB.Prepare("DELETE FROM groupMembers WHERE netId=? AND groupGuid=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(netId, group)
	return err
}

// Remove a user from all his/her groups.
func (ga *MembersAccessor) RemoveAllGroups(netId string) error {
	stmt, err := ga.DB.Prepare("DELETE FROM groupMembers WHERE netId=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(netId)
	return err
}
