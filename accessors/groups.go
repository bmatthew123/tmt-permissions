package accessors

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

// Group struct that reflects the groups table.
type Group struct {
	Guid string
	Area string
	Name string
}

type GroupAccessor struct {
	DB *sql.DB // Database connection
}

// Returns a new group accessor.
func NewGroupAccessor(db *sql.DB) *GroupAccessor {
	return &GroupAccessor{db}
}

// Create a new group.
func (ga *GroupAccessor) Insert(group Group) error {
	stmt, err := ga.DB.Prepare("INSERT INTO groups (guid, area, name) VALUES (?,?,?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(NewGuid(), group.Area, group.Name)
	return err
}

// Gets the group with the given id.
func (ga *GroupAccessor) Get(guid string) (Group, error) {
	g := Group{}
	stmt, err := ga.DB.Prepare("SELECT * FROM groups WHERE guid=?")
	if err != nil {
		return g, err
	}

	row := stmt.QueryRow(guid)
	err = row.Scan(&g.Guid, &g.Area, &g.Name)
	return g, err
}

// Gets all the groups in an area.
func (ga *GroupAccessor) GetByArea(area string) ([]Group, error) {
	groups := make([]Group, 0)
	stmt, err := ga.DB.Prepare("SELECT * FROM groups WHERE area=?")
	if err != nil {
		return groups, err
	}

	rows, err := stmt.Query(area)
	if err != nil {
		return groups, err
	}
	defer rows.Close()
	for rows.Next() {
		g := Group{}
		err = rows.Scan(&g.Guid, &g.Area, &g.Name)
		groups = append(groups, g)
	}

	return groups, nil
}

// Renames a group with the given id to have the provided name.
func (ga *GroupAccessor) Rename(guid, name string) error {
	stmt, err := ga.DB.Prepare("UPDATE groups SET name=? WHERE guid=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(name, guid)
	return err
}

// Delete a group.
func (ga *GroupAccessor) Delete(guid string) error {
	stmt, err := ga.DB.Prepare("DELETE FROM groups WHERE guid=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(guid)
	return err
}

func (ga *GroupAccessor) GetImpliedGroups(netId, area string) ([]Group, error) {
	pa := NewPermissionAccessor(ga.DB)

	// Get all groups
	areaGroups, err := ga.GetByArea(area)
	if err != nil {
		return nil, err
	}

	// Get groups the user is a member of
	userGroups, err := pa.Get(area, netId)
	if err != nil {
		return nil, err
	}

	// Get the permissions the user has
	userPerms, err := pa.GetUserPermissions(netId, area)
	if err != nil {
		return nil, err
	}

	// Determine which groups the user is not in
	possible := make([]string, 0) // possible IMPLIED groups
	for i := 0; i < len(areaGroups); i++ {
		found := false
		for j := 0; j < len(userGroups); j++ {
			if userGroups[j].Guid == areaGroups[i].Guid {
				found = true
				break
			}
		}
		if !found {
			possible = append(possible, areaGroups[i].Guid)
		}
	}

	// Determine if membership is implied for each group
	impGroups := make([]string, 0)
	stmt, err := pa.DB.Prepare("SELECT * FROM policy WHERE actor=?")
	for i := 0; i < len(possible); i++ {
		rows, err := stmt.Query(possible[i])
		if err != nil {
			return nil, err
		}
		implied := true
		for rows.Next() {
			p := Permission{}
			rows.Scan(&p.Actor, &p.Verb, &p.Resource)
			found := false
			for j := 0; j < len(userPerms); j++ {
				if p.Verb == userPerms[j].Verb && p.Resource == userPerms[j].Resource {
					found = true
					break
				}
			}
			if !found {
				implied = false
				break
			}
		}
		if implied {
			impGroups = append(impGroups, possible[i])
		}
		rows.Close()
	}

	// Get complete group information
	result := make([]Group, 0)
	for i := 0; i < len(impGroups); i++ {
		group, err := ga.Get(impGroups[i])
		if err != nil {
			return nil, err
		}
		result = append(result, group)
	}

	return result, nil
}
