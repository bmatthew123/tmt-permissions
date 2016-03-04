package apis

import (
	"fmt"

	eden "github.com/byu-oit-ssengineering/tmt-eden"
	accessors "github.com/byu-oit-ssengineering/tmt-permissions/accessors"
)

// Check whether the user is an admin in the given area
// GET /admin/:netId/:user
func (a *Api) IsAdmin(c *eden.Context) {
	pa := accessors.NewPermissionAccessor(a.DB)

	netId := c.Params[0].Value
	areaGuid := c.Params[1].Value

	admin, err := pa.IsAdmin(netId, areaGuid)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error in IsAdmin by %s (GET /admin/:netId/:user): %v", netId, err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", false})
		return
	}

	c.Respond(200, eden.Response{"OK", admin})
}

// Check whether the user is an admin in the given area
// GET /admin?area=:areaGuid
func (a *Api) GetAdmins(c *eden.Context) {
	pa := accessors.NewPermissionAccessor(a.DB)

	c.Request.ParseForm()
	area, areaOk := c.Request.Form["area"]

	if !areaOk || area[0] == "" {
		c.Respond(400, eden.Response{"ERROR", "Invalid input"})
		return
	}

	admins, err := pa.GetAdmins(area[0])
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error in GetAdmins (GET /admin?area=:areaGuid): v", err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", admins})
		return
	}

	c.Respond(200, eden.Response{"OK", admins})
}

// Check whether the user is an admin in the given area
// GET /superuser
func (a *Api) GetAllSU(c *eden.Context) {
	pa := accessors.NewPermissionAccessor(a.DB)

	sus, err := pa.GetAllSU()
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error in GetAllSU (GET /superuser): %v", err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", sus})
		return
	}

	c.Respond(200, eden.Response{"OK", sus})
}

// Grant a user admin access in an area
// POST /admin?area=:areaGuid&netId=:netId
func (a *Api) AddAdmin(c *eden.Context) {
	pa := accessors.NewPermissionAccessor(a.DB)

	// Parse input
	c.Request.ParseForm()
	area, areaOk := c.Request.Form["area"]
	netId, netIdOk := c.Request.Form["netId"]

	accessors.Log("notice", c.User.NetId, fmt.Sprintf("%s called AddAdmin (POST /admin?area=:areaGuid&netId=:netId)", netId[0]), true, pa.DB)

	if !areaOk || !netIdOk || netId[0] == "" || area[0] == "" {
		c.Respond(400, eden.Response{"ERROR", "Invalid input"})
		return
	}

	// Check that the user is an admin first
	isAdmin, err := pa.IsAdmin(c.User.NetId, c.User.Area)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on IsAdmin in AddAdmin by %s (POST /admin?area=:areaGuid&netId=:netId): %v", netId[0], err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	if !isAdmin {
		isSU, err := pa.IsSuperuser(c.User.NetId)
		if err != nil {
			accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on IsSuperuser in AddAdmin by %s (POST /admin?area=:areaGuid&netId=:netId): %v", netId[0], err), true, pa.DB)
			c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
			return
		}

		if !isSU {
			c.Respond(403, eden.Response{"ERROR", "You need to be an admin to grant admin rights"})
			return
		}
	}

	err = pa.AddAdmin(netId[0], area[0])
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on AddAdmin by %s (POST /admin?area=:areaGuid&netId=:netId): ", netId[0], err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	c.Respond(200, eden.Response{"OK", "success"})
}

// Revoke admin access
// DELETE /admin/:netId/:areaGuid
func (a *Api) DeleteAdmin(c *eden.Context) {
	pa := accessors.NewPermissionAccessor(a.DB)

	netId := c.Params[0].Value
	areaGuid := c.Params[1].Value

	accessors.Log("notice", c.User.NetId, fmt.Sprintf("%s called DeleteAdmin (DELETE /admin/:netId/:areaGuid)", netId), true, pa.DB)

	// Check that the user is an admin first
	isAdmin, err := pa.IsAdmin(c.User.NetId, areaGuid)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on IsAdmin in DeleteAdmin by %s (DELETE /admin/:netId/:areaGuid): %v", netId, err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}
	if !isAdmin {
		isSU, err := pa.IsSuperuser(c.User.NetId)
		if err != nil {
			accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on IsSuperuser in DeleteAdmin by %s (DELETE /admin/:netId/:areaGuid): %v", netId, err), true, pa.DB)
			c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
			return
		}

		if !isSU {
			c.Respond(403, eden.Response{"ERROR", "You need to be an admin to grant admin rights"})
			return
		}
	}

	err = pa.DeleteAdmin(netId, areaGuid)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error in DeleteAdmin by %s (DELETE /admin/:netId/:areaGuid): %v", netId, err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	c.Respond(200, eden.Response{"OK", "success"})
}

// Check whether the user has elevated to superuser
// GET /superuser/is/:netId
func (a *Api) IsSuperuser(c *eden.Context) {
	pa := accessors.NewPermissionAccessor(a.DB)

	netId := c.Params[0].Value

	su, err := pa.IsSuperuser(netId)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error in IsSuperuser by %s (GET /superuser/is/:netId): %v", netId, err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", false})
		return
	}

	c.Respond(200, eden.Response{"OK", su})
}

// Check whether the user can elevate to superuser privileges
// GET /superuser/can/:netId
func (a *Api) CanSuperuser(c *eden.Context) {
	pa := accessors.NewPermissionAccessor(a.DB)

	netId := c.Params[0].Value

	su, err := pa.CanSuperuser(netId)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error in CanSuperuser by %s (GET /superuser/can/:netId): %v", netId, err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", false})
		return
	}

	c.Respond(200, eden.Response{"OK", su})
}

// Grant a user superuser access
// POST /superuser?netId=:netId
func (a *Api) AddSU(c *eden.Context) {
	pa := accessors.NewPermissionAccessor(a.DB)

	// Parse input
	c.Request.ParseForm()
	netId, netIdOk := c.Request.Form["netId"]

	accessors.Log("notice", c.User.NetId, fmt.Sprintf("%s called AddSU (POST /superuser?netId=:netId)", netId[0]), true, pa.DB)

	if !netIdOk || netId[0] == "" {
		c.Respond(400, eden.Response{"ERROR", "Invalid input"})
		return
	}

	// Check that the user is superuser
	isSU, err := pa.IsSuperuser(c.User.NetId)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on IsSuperuser in AddSU by %s (POST /superuser?netId=:netId): %v", netId[0], err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}
	if !isSU {
		c.Respond(403, eden.Response{"ERROR", "You need to be superuser to add superuser rights"})
		return
	}

	err = pa.AddSU(netId[0])
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error in AddSU by %s (POST /superuser?netId=:netId): %v", netId[0], err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	c.Respond(200, eden.Response{"OK", "success"})
}

// Elevate to or stop superuser access.
// PUT /superuser/:netId?elevate=true
func (a *Api) Elevate(c *eden.Context) {
	pa := accessors.NewPermissionAccessor(a.DB)

	netId := c.Params[0].Value

	accessors.Log("notice", c.User.NetId, fmt.Sprintf("%s called Elevate (PUT /superuser/:netId?elevate=true)", netId), true, pa.DB)

	su, err := pa.CanSuperuser(netId)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on CanSuperuser in Elevate by %s (PUT /superuser/:netId?elevate=true): %v", netId, err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", false})
		return
	}
	if !su {
		c.Respond(403, eden.Response{"ERROR", "You do not have the right to become superuser"})
		return
	}

	// Parse input
	c.Request.ParseForm()
	elevate, elevateOk := c.Request.Form["elevate"]
	if elevateOk && elevate[0] == "true" {
		err = pa.ElevateToSU(netId)
		if err != nil {
			accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on ElevateToSU in Elevate by %s (PUT /superuser/:netId?elevate=true): %v", netId, err), true, pa.DB)
			c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
			return
		}
	} else {
		err = pa.StopSU(netId)
		if err != nil {
			accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on StopSU in Elevate by %s (PUT /superuser/:netId?elevate=true): %v", netId, err), true, pa.DB)
			c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
			return
		}
	}

	c.Respond(200, eden.Response{"OK", "success"})
}

// Revoke superuser access.
// DELETE /superuser/:netId
func (a *Api) DeleteSU(c *eden.Context) {
	pa := accessors.NewPermissionAccessor(a.DB)

	netId := c.Params[0].Value

	accessors.Log("notice", c.User.NetId, fmt.Sprintf("%s called DeleteSU (DELETE /superuser/:netId)", netId), true, pa.DB)

	// Check that the user is superuser
	isSU, err := pa.IsSuperuser(c.User.NetId)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on IsSuperuser in DeleteSU by %s (DELETE /superuser/:netId): %v", netId, err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	if !isSU {
		c.Respond(403, eden.Response{"ERROR", "You need to be superuser to add superuser rights"})
		return
	}

	err = pa.DeleteSU(netId)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error in DeleteSU by %s (DELETE /superuser/:netId): %v", netId, err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	c.Respond(200, eden.Response{"OK", "success"})
}
