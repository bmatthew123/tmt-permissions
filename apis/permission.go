package apis

import (
	"fmt"
	"net/url"

	eden "github.com/byu-oit-ssengineering/tmt-eden"
	accessors "github.com/byu-oit-ssengineering/tmt-permissions/accessors"
)

// Check whether the user has permission to access a resource.
// GET /permission?object=:objectGUID&verb=:verb&actors[]=:actors
// The GET data passed in will be used to evaluate whether or
//   not the user has access to something. The verb is the
//   action the user is trying to perform, object is the
//   resource being accessed, and actors is a list of the
//   guids of the user, the user's groups, and the current area.
func (a *Api) CheckPermission(c *eden.Context) {
	pa := accessors.NewPermissionAccessor(a.DB)

	// Parse the request
	query, err := url.ParseQuery(c.Request.URL.RawQuery)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on ParseQuery in CheckPermission (GET /permission?object=:objectGUID&verb=:verb&actors[]=:actors): %v", err), true, pa.DB)
		c.Respond(400, eden.Response{"ERROR", false})
		return
	}

	areaGuid, areaGuidOk := query["areaGuid"]
	employeeGuid, employeeGuidOk := query["employeeGuid"]
	verb, verbOk := query["verb"]
	resource, resourceOk := query["resource"]

	if !areaGuidOk || !employeeGuidOk || !verbOk || !resourceOk {
		c.Respond(400, eden.Response{"ERROR", false})
		return
	}

	// Check superuser
	su, err := pa.IsSuperuser(employeeGuid[0])
	if su {
		c.Respond(200, eden.Response{"OK", true})
		return
	}

	// Check admin
	admin, err := pa.IsAdmin(employeeGuid[0], areaGuid[0])
	if admin {
		c.Respond(200, eden.Response{"OK", true})
		return
	}

	// Check permission
	actors, err := pa.Get(areaGuid[0], employeeGuid[0])
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on Get in CheckPermission (GET /permission?object=:objectGUID&verb=:verb&actors[]=:actors): %v", err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", false})
		return
	}

	actorArray := make([]string, 0)
	for i := 0; i < len(actors); i++ {
		actorArray = append(actorArray, actors[i].Guid)
	}

	actorArray = append(actorArray, areaGuid[0])
	permission, err := pa.CheckPermission(actorArray, resource[0], verb[0])
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error in CheckPermission (GET /permission?object=:objectGUID&verb=:verb&actors[]=:actors): %v", err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", false})
		return
	}

	c.Respond(200, eden.Response{"OK", permission})
}

// Get all permission groups that have access to a specified verb
// GET /permission/:resourceGUID/:verb
func (a *Api) GetGroupsByVerb(c *eden.Context) {
	ga := accessors.NewGroupAccessor(a.DB)
	pa := accessors.NewPermissionAccessor(a.DB)

	//Parse input
	resourceGuid := c.Params[0].Value
	verb := c.Params[1].Value

	groups := make([]accessors.Group, 0)
	rawGroups, err := ga.GetByArea(c.User.Area)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on GetByArea in GetGroupsByVerb (GET /permission/:resourceGUID/:verb): %v", err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", false})
		return
	}

	for i := 0; i < len(rawGroups); i++ {
		// Check permission
		permission, err := pa.CheckPermission([]string{rawGroups[i].Guid}, resourceGuid, verb)
		if err != nil {
			accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on CheckPermission in GetGroupsByVerb (GET /permission/:resourceGUID/:verb): %v", err), true, pa.DB)
			c.Respond(500, eden.Response{"ERROR", false})
			return
		}

		if permission {
			groups = append(groups, rawGroups[i])
		}
	}

	c.Respond(200, eden.Response{"OK", groups})
}

// POST /permission actor=:actor verb=:verb resource=:resource
func (a *Api) AddPermission(c *eden.Context) {
	pa := accessors.NewPermissionAccessor(a.DB)

	accessors.Log("notice", c.User.NetId, "Called AddPermission (POST /permission actor=:actor verb=:verb resource=:resource)", true, pa.DB)

	// Parse input
	c.Request.ParseForm()
	actor, actorOk := c.Request.Form["actor"]
	verb, verbOk := c.Request.Form["verb"]
	resource, resourceOk := c.Request.Form["resource"]
	if !actorOk || !verbOk || !resourceOk {
		c.Respond(400, eden.Response{"ERROR", "Not enough information given"})
		return
	}

	// Check superuser
	su, err := pa.IsSuperuser(c.User.NetId)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on IsSuperuser in AddPermission (POST /permission actor=:actor verb=:verb resource=:resource): %v", err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	if su {
		// Insert permission
		err = pa.Add(actor[0], verb[0], resource[0])
		if err != nil {
			accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on Add in AddPermission (POST /permission actor=:actor verb=:verb resource=:resource): %v", err), true, pa.DB)
			c.Respond(500, eden.Response{"ERROR", "An error occurred while granting permission"})
			return
		}

		c.Respond(200, eden.Response{"OK", "success"})
		return
	}

	// Check admin
	admin, err := pa.IsAdmin(c.User.NetId, c.User.Area)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on IsAdmin in AddPermission (POST /permission actor=:actor verb=:verb resource=:resource): %v", err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	if admin {
		// Insert permission
		err = pa.Add(actor[0], verb[0], resource[0])
		if err != nil {
			c.Respond(500, eden.Response{"ERROR", "An error occurred while granting permission"})
			return
		}
		c.Respond(200, eden.Response{"OK", "success"})
		return
	}

	// Check that requestor has permission
	actors, err := pa.Get(c.User.Area, c.User.NetId)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on Get in AddPermission (POST /permission actor=:actor verb=:verb resource=:resource): %v", err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	actorArray := make([]string, 0)
	for i := 0; i < len(actors); i++ {
		actorArray = append(actorArray, actors[i].Guid)
	}
	actorArray = append(actorArray, c.User.Area)
	permission, err := pa.CheckPermission(actorArray, resource[0], verb[0])
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on CheckPermission in AddPermission (POST /permission actor=:actor verb=:verb resource=:resource): %v", err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	if !permission {
		c.Respond(403, eden.Response{"FAILURE", "You need to have this permission in order to grant it"})
		return
	}

	// Insert permission
	err = pa.Add(actor[0], verb[0], resource[0])
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on Add in AddPermission (POST /permission actor=:actor verb=:verb resource=:resource): %v", err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error occurred while granting permission"})
		return
	}

	c.Respond(200, eden.Response{"OK", "success"})
}

// DELETE /permission/:actor/:verb/:resource
func (a *Api) DeletePermission(c *eden.Context) {
	pa := accessors.NewPermissionAccessor(a.DB)

	accessors.Log("notice", c.User.NetId, "Called DeletePermission (DELETE /permission/:actor/:verb/:resource)", true, pa.DB)

	// Parse input
	actor := c.Params[0].Value
	verb := c.Params[1].Value
	object := c.Params[2].Value

	// Check superuser
	su, err := pa.IsSuperuser(c.User.NetId)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on IsSuperuser in DeletePermission (DELETE /permission/:actor/:verb/:resource): %v", err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	if su {
		// Insert permission
		err = pa.Delete(actor, verb, object)
		if err != nil {
			c.Respond(500, eden.Response{"ERROR", "An error occurred while granting permission"})
			return
		}
		c.Respond(200, eden.Response{"OK", "success"})
		return
	}

	// Check admin
	admin, err := pa.IsAdmin(c.User.NetId, c.User.Area)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on IsAdmin in DeletePermission (DELETE /permission/:actor/:verb/:resource): %v", err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}
	if admin {
		// Insert permission
		err = pa.Delete(actor, verb, object)
		if err != nil {
			accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on Delete in DeletePermission (DELETE /permission/:actor/:verb/:resource): %v", err), true, pa.DB)
			c.Respond(500, eden.Response{"ERROR", "An error occurred while granting permission"})
			return
		}

		c.Respond(200, eden.Response{"OK", "success"})
		return
	}

	// Check that requestor has permission
	actors, err := pa.Get(c.User.Area, c.User.NetId)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on Get in DeletePermission (DELETE /permission/:actor/:verb/:resource): %v", err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	actorArray := make([]string, 0)
	for i := 0; i < len(actors); i++ {
		actorArray = append(actorArray, actors[i].Guid)
	}

	actorArray = append(actorArray, c.User.Area)
	permission, err := pa.CheckPermission(actorArray, object, verb)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on CheckPermission in DeletePermission (DELETE /permission/:actor/:verb/:resource): %v", err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	if !permission {
		c.Respond(403, eden.Response{"FAILURE", "You need to have this permission in order to grant it"})
		return
	}

	// Delete permission
	err = pa.Delete(actor, verb, object)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on Delete in DeletePermission (DELETE /permission/:actor/:verb/:resource): %v", err), true, pa.DB)

		c.Respond(500, eden.Response{"ERROR", "An error occurred while revoking permission"})
		return
	}

	c.Respond(200, eden.Response{"OK", "success"})
}

// GET /groups/:employeeGuid/:areaGuid/
func (a *Api) GetGroupsByEmployeeGuid(c *eden.Context) {
	pa := accessors.NewPermissionAccessor(a.DB)

	//Parse input
	areaGuid := c.Params[0].Value
	employeeGuid := c.Params[1].Value

	//Get Groups
	groups, err := pa.Get(areaGuid, employeeGuid)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on Get in GetGroupsByEmployeeGuid (GET /groups/:employeeGuid/:areaGuid/): %v", err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred while getting the groups"})
		return
	}

	c.Respond(200, eden.Response{"OK", groups})
}

// GET /permission/:actor
// Return the list of permissions the group has access to.
func (a *Api) GetGroupPermissions(c *eden.Context) {
	pa := accessors.NewPermissionAccessor(a.DB)

	//Parse input
	actorGuid := c.Params[0].Value

	//Get Groups
	permissions, err := pa.GetGroupPermissions(actorGuid)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error in GetGroupPermissions (GET /permission/:actor): %v", err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred while getting the permissions"})
		return
	}

	c.Respond(200, eden.Response{"OK", permissions})
}

// GET /permission/user/:netId/:areaGuid
func (a *Api) GetUserPermissions(c *eden.Context) {
	pa := accessors.NewPermissionAccessor(a.DB)

	netId := c.Params[0].Value
	area := c.Params[1].Value

	permissions, err := pa.GetUserPermissions(netId, area)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error in GetUserPermissions (GET /permission/user/:netId/:areaGuid): %v", err), true, pa.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred while getting the permissions"})
		return
	}

	c.Respond(200, eden.Response{"OK", permissions})
}
