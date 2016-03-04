package apis

import (
	"fmt"
	"strconv"

	eden "github.com/byu-oit-ssengineering/tmt-eden"
	accessors "github.com/byu-oit-ssengineering/tmt-permissions/accessors"
)

// Get all the groups for an area.
// GET /groups?area=:area
func (a *Api) GetGroupsByArea(c *eden.Context) {
	ga := accessors.NewGroupAccessor(a.DB)

	c.Request.ParseForm()
	area, ok := c.Request.Form["area"]
	if !ok || area[0] == "" {
		c.Respond(400, eden.Response{"ERROR", "Invalid area"})
		return
	}

	// Get the area's groups
	result, err := ga.GetByArea(area[0])
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on GetByArea in GetGroupsByArea (GET /groups?area=:area): %v", err), true, ga.DB)
		c.Respond(500, eden.Response{"ERROR", "An error occurred while retrieving groups"})
		return
	}

	// Respond
	c.Respond(200, eden.Response{"OK", result})
}

// Gets a group by guid.
// GET /groups/:guid
func (a *Api) GetGroup(c *eden.Context) {
	// Create new group accessor
	ga := accessors.NewGroupAccessor(a.DB)

	// Parse the group id
	guid := c.Params[0].Value

	// Get the group
	group, err := ga.Get(guid)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on Get in GetGroup (GET /groups/:guid): %v", err), true, ga.DB)
		c.Respond(500, eden.Response{"ERROR", "An error occurred while retrieving group information" + err.Error()})
		return
	}

	// Respond
	c.Respond(200, eden.Response{"OK", group})
}

// Creates a new group
// POST /groups name=:newGroupName, area=:areaGuid
func (a *Api) CreateGroup(c *eden.Context) {
	// Create new group accessor
	ga := accessors.NewGroupAccessor(a.DB)

	accessors.Log("notice", c.User.NetId, "Called CreateGroup (POST /groups name=:newGroupName, area=:areaGuid)", true, ga.DB)

	// Parse area and group name from POST data.
	c.Request.ParseForm()
	if area, ok := c.Request.Form["area"]; !ok || area[0] == "" {
		c.Respond(400, eden.Response{"ERROR", "Invalid area"})
		return
	}

	if name, ok := c.Request.Form["name"]; !ok || name[0] == "" {
		c.Respond(400, eden.Response{"ERROR", "Invalid group name"})
		return
	}

	name := c.Request.Form["name"][0]
	area := c.Request.Form["area"][0]
	group := accessors.Group{Area: area, Name: name}

	// Insert the group and test for errors
	if err := ga.Insert(group); err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on Insert in CreateGroup (POST /groups name=:newGroupName, area=:areaGuid): %v", err), true, ga.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	// Respond
	c.Respond(200, eden.Response{"OK", "success"})
}

// Update a group's name.
// PUT /groups/:guid name=:newName
func (a *Api) RenameGroup(c *eden.Context) {
	// Create new group accessor
	ga := accessors.NewGroupAccessor(a.DB)

	accessors.Log("notice", c.User.NetId, "Called RenameGroup (PUT /groups/:guid name=:newName)", true, ga.DB)

	// Parse group guid
	guid := c.Params[0].Value

	// Parse new group name
	c.Request.ParseForm()
	if group, ok := c.Request.Form["name"]; !ok || group[0] == "" {
		c.Respond(400, eden.Response{"ERROR", "No group name specified"})
		return
	}

	name := c.Request.Form["name"][0]
	if err := ga.Rename(guid, name); err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on Rename in RenameGroup (PUT /groups/:guid name=:newName): %v", err), true, ga.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	// Respond
	c.Respond(200, eden.Response{"OK", "success"})
}

// Delete a group.
// DELETE /groups/:guid
func (a *Api) DeleteGroup(c *eden.Context) {
	// Create new group accessor
	ga := accessors.NewGroupAccessor(a.DB)

	accessors.Log("notice", c.User.NetId, "Called DeleteGroup (DELETE /groups/:guid)", true, ga.DB)

	// Parse group guid
	guid := c.Params[0].Value

	// Delete the group
	if err := ga.Delete(guid); err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on Delete in DeleteGroup (DELETE /groups/:guid): %v", err), true, ga.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	// Respond
	c.Respond(200, eden.Response{"OK", "success"})
}

// GET /groups?areaGuid=:area&netId=:netId&implied=true
func (a *Api) GetUserGroups(c *eden.Context) {
	ga := accessors.NewGroupAccessor(a.DB)
	pa := accessors.NewPermissionAccessor(a.DB)

	c.Request.ParseForm()
	area, ok := c.Request.Form["area"]
	netId, ok1 := c.Request.Form["netId"]
	if !ok || area[0] == "" {
		c.Respond(400, eden.Response{"ERROR", "Invalid area"})
		return
	}
	if !ok1 || netId[0] == "" {
		c.Respond(400, eden.Response{"ERROR", "Invalid netId"})
		return
	}
	imp, ok := c.Request.Form["implied"]
	implied := false
	if ok {
		implied, _ = strconv.ParseBool(imp[0])
	}

	groups, err := pa.Get(area[0], netId[0])
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on Get in GetUserGroups by %s (GET /groups?areaGuid=:area&netId=:netId&implied=true): %v", netId[0], err), true, ga.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred while getting the permissions"})
		return
	}

	if implied {
		impGroups, err := ga.GetImpliedGroups(netId[0], area[0])
		if err != nil {
			accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on GetImpliedGroups in GetUserGroups by %s (GET /groups?areaGuid=:area&netId=:netId&implied=true): %v", netId[0], err), true, ga.DB)
			c.Respond(500, eden.Response{"ERROR", err.Error()})
			return
		}
		groups = append(groups, impGroups...)
	}

	c.Respond(200, eden.Response{"OK", groups})
}
