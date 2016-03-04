package accessors

import (
	_ "github.com/go-sql-driver/mysql"
)

// Tells whether or not a user is an admin.
func (pa *PermissionAccessor) IsAdmin(netId, areaGuid string) (bool, error) {

	// execute query
	stmt, err := pa.DB.Prepare("SELECT * FROM admin WHERE netId=? AND area=?")
	if err != nil {
		return false, err
	}

	rows, err := stmt.Query(netId, areaGuid)
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	defer rows.Close()

	// If one or more rows was returned then the user is an admin.
	if rows.Next() {
		return true, nil
	}

	// No rows were returned, return false.
	return false, nil
}

// Tells whether or not a user has elevated to superuser rights
func (pa *PermissionAccessor) GetAdmins(area string) ([]string, error) {
	users := make([]string, 0)

	// execute query
	stmt, err := pa.DB.Prepare("SELECT netId FROM admin WHERE area=?")
	if err != nil {
		return users, err
	}

	rows, err := stmt.Query(area)
	if err != nil {
		return users, err
	}
	defer stmt.Close()
	defer rows.Close()

	// If one or more rows was returned then the user has elevated to superuser
	for rows.Next() {
		var user string
		rows.Scan(&user)
		users = append(users, user)
	}

	// User cannot be superuser
	return users, nil
}

// Tells whether or not a user has elevated to superuser rights
func (pa *PermissionAccessor) GetAllSU() ([]string, error) {
	users := make([]string, 0)

	// execute query
	stmt, err := pa.DB.Prepare("SELECT netId FROM superuser")
	if err != nil {
		return users, err
	}

	rows, err := stmt.Query()
	if err != nil {
		return users, err
	}
	defer stmt.Close()
	defer rows.Close()

	// If one or more rows was returned then the user has elevated to superuser
	for rows.Next() {
		var user string
		rows.Scan(&user)
		users = append(users, user)
	}

	// User cannot be superuser
	return users, nil
}

// Tells whether or not a user has elevated to superuser rights
func (pa *PermissionAccessor) IsSuperuser(netId string) (bool, error) {

	// execute query
	stmt, err := pa.DB.Prepare("SELECT active FROM superuser WHERE netId=?")
	if err != nil {
		return false, err
	}

	rows, err := stmt.Query(netId)
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	defer rows.Close()

	// If one or more rows was returned then the user has elevated to superuser
	if rows.Next() {
		var active int
		err = rows.Scan(&active)
		if active == 1 {
			return true, err
		} else {
			return false, err
		}
	}

	// User cannot be superuser
	return false, nil
}

// Tells whether or not a user can be superuser
func (pa *PermissionAccessor) CanSuperuser(netId string) (bool, error) {

	// execute query
	stmt, err := pa.DB.Prepare("SELECT active FROM superuser WHERE netId=?")
	if err != nil {
		return false, err
	}

	rows, err := stmt.Query(netId)
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	defer rows.Close()

	// If a row is returned the user can be superuser
	if rows.Next() {
		return true, nil
	}

	// User cannot be superuser
	return false, nil
}

// Grant admin access.
func (pa *PermissionAccessor) AddAdmin(netId, areaGuid string) error {
	stmt, err := pa.DB.Prepare("INSERT INTO admin (netId, area) VALUES (?,?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(netId, areaGuid)
	return err
}

// Revoke admin access.
func (pa *PermissionAccessor) DeleteAdmin(netId, areaGuid string) error {
	stmt, err := pa.DB.Prepare("DELETE FROM admin WHERE netId=? AND area=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(netId, areaGuid)
	return err
}

// Grant superuser access.
func (pa *PermissionAccessor) AddSU(netId string) error {
	stmt, err := pa.DB.Prepare("INSERT INTO superuser (netId) VALUES (?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(netId)
	return err
}

// Elevate to superuser access.
func (pa *PermissionAccessor) ElevateToSU(netId string) error {
	stmt, err := pa.DB.Prepare("UPDATE superuser SET active=1 WHERE netId=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(netId)
	return err
}

// Return to normal user status.
func (pa *PermissionAccessor) StopSU(netId string) error {
	stmt, err := pa.DB.Prepare("UPDATE superuser SET active=0 WHERE netId=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(netId)
	return err
}

// Revoke superuser access.
func (pa *PermissionAccessor) DeleteSU(netId string) error {
	stmt, err := pa.DB.Prepare("DELETE FROM superuser WHERE netId=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(netId)
	return err
}
