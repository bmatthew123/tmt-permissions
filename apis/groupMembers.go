package apis

import (
	"fmt"
	"net/url"

	eden "github.com/byu-oit-ssengineering/tmt-eden"
	accessors "github.com/byu-oit-ssengineering/tmt-permissions/accessors"
)

// Get a list of all the groups a user is a member of.
// GET /groupMembers?netId=:netId
func (a *Api) GetGroupsByNetId(c *eden.Context) {
	ma := accessors.NewMembersAccessor(a.DB)

	// Parse the request query and get area
	query, err := url.ParseQuery(c.Request.URL.RawQuery)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on ParseQuery in GetGroupsByNetId (GET /groupMembers?netId=:netId): %v", err), true, ma.DB)
		c.Respond(400, eden.Response{"ERROR", "Unable to process request"})
		return
	}

	netId, ok := query["netId"]
	if !ok {
		c.Respond(400, eden.Response{"ERROR", "No netId specified"})
		return
	}

	// Get the areas groups
	result, err := ma.GetUserGroups(netId[0])
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on GetUserGroups in GetGroupsByNetId by %s (GET /groupMembers?netId=:netId): %v", netId[0], err), true, ma.DB)
		c.Respond(500, eden.Response{"ERROR", "An error occurred while retrieving groups"})
		return
	}

	// Respond
	c.Respond(200, eden.Response{"OK", result})
}

// Get a list of all the members of a certain group.
// GET /groupMembers/:groupId
func (a *Api) GetGroupMembers(c *eden.Context) {
	// Create new group accessor
	ma := accessors.NewMembersAccessor(a.DB)

	// Parse the group id
	id := c.Params[0].Value

	// Get the group
	members, err := ma.GetGroupMembers(id)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error in GetGroupMembers (GET /groupMembers/:groupId): %v", err), true, ma.DB)
		c.Respond(500, eden.Response{"ERROR", "An error occurred while retrieving group information"})
		return
	}

	// Respond
	c.Respond(200, eden.Response{"OK", members})
}

// Add a user to a group.
// POST /groupMembers netId=:netId, group=:groupId
func (a *Api) AddGroupMember(c *eden.Context) {
	// Create new group accessor
	ma := accessors.NewMembersAccessor(a.DB)

	accessors.Log("notice", c.User.NetId, "Called AddGroupMember (POST /groupMembers netId=:netId, group=:groupId)", true, ma.DB)

	// Parse group id and netId from POST data.
	c.Request.ParseForm()
	netId, netIdOk := c.Request.Form["netId"]
	group, groupOk := c.Request.Form["group"]
	if !netIdOk || !groupOk {
		c.Respond(400, eden.Response{"ERROR", "Invalid netId or groupId"})
		return
	}

	// Insert the group and test for errors
	if err := ma.AddToGroup(netId[0], group[0]); err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on AddToGroup in AddGroupMember by %s (POST /groupMembers netId=:netId, group=:groupId): %v", netId[0], err), true, ma.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	// Respond
	c.Respond(200, eden.Response{"OK", "success"})
}

// Remove a user from a group.
// DELETE /groupMembers/:netId/:groupGuid
func (a *Api) RemoveGroupMember(c *eden.Context) {
	// Create new group accessor
	ma := accessors.NewMembersAccessor(a.DB)

	// Parse group id
	netId := c.Params[0].Value
	groupId := c.Params[1].Value

	accessors.Log("notice", c.User.NetId, fmt.Sprintf("%s called RemoveGroupMember (DELETE /groupMembers/:netId/:groupGuid)", netId), true, ma.DB)

	// Delete the group
	if err := ma.RemoveFromGroup(netId, groupId); err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on RemoveFromGroup in RemoveGroupMember by %s (DELETE /groupMembers/:netId/:groupGuid): %v", netId, err), true, ma.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	// Respond
	c.Respond(200, eden.Response{"OK", "success"})
}

// Removes a user from all his/her groups
// DELETE /groupMembers/:netId
func (a *Api) RemoveFromAllGroups(c *eden.Context) {
	ma := accessors.NewMembersAccessor(a.DB)

	// Parse netId
	netId := c.Params[0].Value

	err := ma.RemoveAllGroups(netId)
	if err != nil {
		accessors.Log("error", c.User.NetId, fmt.Sprintf("Error on RemoveAllGroups in RemoveFromAllGroups by %s (DELETE /groupMembers/:netId): %v", netId, err), true, ma.DB)
		c.Respond(500, eden.Response{"ERROR", "An error has occurred"})
		return
	}

	// Respond
	c.Respond(200, eden.Response{"OK", "success"})
}
