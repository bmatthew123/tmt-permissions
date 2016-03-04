package accessors

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type Permission struct {
	Actor    string // Actor Guid
	Verb     string
	Resource string // Resource Guid
}

type PermissionAccessor struct {
	DB *sql.DB // Database connection
}

// Returns a new group accessor.
func NewPermissionAccessor(db *sql.DB) *PermissionAccessor {
	return &PermissionAccessor{db}
}

// Tells whether or not a user has permission to access a resource.
func (pa *PermissionAccessor) CheckPermission(permissions []string, obj, verb string) (bool, error) {

	// build query and params
	query := "SELECT * FROM policy WHERE verb=? AND resource=? AND actor IN (?"
	var params []interface{}
	params = append(params, verb)
	params = append(params, obj)
	params = append(params, permissions[0])
	for i := 1; i < len(permissions); i++ {
		query += ",?"
		params = append(params, permissions[i])
	}
	query += ")"

	// execute query
	stmt, err := pa.DB.Prepare(query)
	if err != nil {
		return false, err
	}

	rows, err := stmt.Query(params...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	// If one or more rows was returned then the user has permission.
	if rows.Next() {
		return true, nil
	}

	// No rows were returned, return false.
	return false, nil
}

// Inserts into the policy table which grants permission to a user/group/area
//   to access a certain resource.
func (pa *PermissionAccessor) Add(actor, verb, obj string) error {
	stmt, err := pa.DB.Prepare("INSERT INTO policy (actor, verb, resource) VALUES (?,?,?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(actor, verb, obj)
	return err
}

// Deletes an entry from the policy table, which revokes permission
//   for a user/group/area to access a resource.
func (pa *PermissionAccessor) Delete(actor, verb, resource string) error {
	stmt, err := pa.DB.Prepare("DELETE FROM policy WHERE actor=? AND verb=? AND resource=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(actor, verb, resource)
	return err
}

//Gets all the groups a user is in, but restricted by area. Puts the area in the array so that
//the api can then act on without having to change anything
func (pa *PermissionAccessor) Get(area, netId string) ([]Group, error) {
	groups := make([]Group, 0)
	stmt, err := pa.DB.Prepare("SELECT groups.* FROM groups JOIN groupMembers ON groups.guid = groupMembers.groupGuid WHERE groupMembers.netId=? AND groups.area=?")
	if err != nil {
		return groups, err
	}

	rows, err := stmt.Query(netId, area)
	if err != nil {
		return groups, err
	}
	defer rows.Close()
	for rows.Next() {
		var g Group
		err = rows.Scan(&g.Guid, &g.Area, &g.Name)
		groups = append(groups, g)
	}

	return groups, nil
}

// Return a list of permissions that a group has.
func (pa *PermissionAccessor) GetGroupPermissions(groupGuid string) ([]Permission, error) {
	perms := make([]Permission, 0)
	stmt, err := pa.DB.Prepare("SELECT * FROM policy WHERE actor=?")
	if err != nil {
		return perms, err
	}

	rows, err := stmt.Query(groupGuid)
	if err != nil {
		return perms, err
	}
	defer rows.Close()
	for rows.Next() {
		var p Permission
		err = rows.Scan(&p.Actor, &p.Verb, &p.Resource)
		perms = append(perms, p)
	}

	return perms, nil
}

// Return a list of permissions that a user has.
func (pa *PermissionAccessor) GetUserPermissions(netId, area string) ([]Permission, error) {
	perms := make([]Permission, 0)

	// Pull the user's groups
	groups, err := pa.Get(area, netId)
	if err != nil {
		return nil, err
	}

	// build query and params
	query := "SELECT * FROM policy WHERE actor IN (?"
	var params []interface{}
	params = append(params, area)
	for i := 0; i < len(groups); i++ {
		query += ",?"
		params = append(params, groups[i].Guid)
	}
	query += ")"
	stmt, err := pa.DB.Prepare(query)
	if err != nil {
		return nil, err
	}

	// Get all things the user is allowed to do
	rows, err := stmt.Query(params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p Permission
		err = rows.Scan(&p.Actor, &p.Verb, &p.Resource)
		p.Actor = "" // actor is irrelevant
		perms = append(perms, p)
	}

	// Eliminate duplicate entries (i.e. user is in two groups with access to the same permission)
	uniquePerms := make(map[Permission]bool) // Use this map like a set to get unique permissions, the values don't matter, just the keys
	for i := 0; i < len(perms); i++ {
		uniquePerms[perms[i]] = false
	}
	finalPerms := make([]Permission, 0)
	for p, _ := range uniquePerms {
		finalPerms = append(finalPerms, p)
	}

	return finalPerms, nil
}
