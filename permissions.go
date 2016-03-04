package main

import (
	"fmt"

	eden "github.com/byu-oit-ssengineering/tmt-eden"
	apis "github.com/byu-oit-ssengineering/tmt-permissions/apis"
)

// Responds with the allowed HTTP methods for this microservice.
func Options(c *eden.Context) {
	c.Response.Header().Add("Access-Control-Allow-Origin", "*")
	c.Response.Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	c.Response.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	c.Response.WriteHeader(200)
}

func main() {
	r := eden.New()
	r.Use(eden.Authorize)

	a, err := apis.New()
	if err != nil {
		panic(err)
	}

	// Register API paths

	// Permissions
	r.GET("/permission", a.CheckPermission)
	r.GET("/permission/verbs/:resourceGUID/:verb", a.GetGroupsByVerb)
	r.GET("/permission/groups/:group", a.GetGroupPermissions)
	r.GET("/permission/user/:netId/:area", a.GetUserPermissions)
	r.POST("/permission", a.AddPermission)
	r.DELETE("/permission/:actor/:verb/:resource", a.DeletePermission)

	// Admin/Superuser
	r.GET("/admin", a.GetAdmins)
	r.GET("/admin/:netId/:area", a.IsAdmin)
	r.POST("/admin", a.AddAdmin)
	r.DELETE("/admin/:netId/:area", a.DeleteAdmin)
	r.GET("/superuser", a.GetAllSU)
	r.GET("/superuser/is/:netId", a.IsSuperuser)
	r.GET("/superuser/can/:netId", a.CanSuperuser)
	r.POST("/superuser", a.AddSU)
	r.PUT("/superuser/:netId", a.Elevate)
	r.DELETE("/superuser/:netId", a.DeleteSU)

	// Groups
	r.GET("/groups/:guid", a.GetGroup)
	r.GET("/groups", a.GetGroupsByArea)
	r.GET("/permission/groups", a.GetUserGroups)
	r.POST("/groups", a.CreateGroup)
	r.PUT("/groups/:guid", a.RenameGroup)
	r.DELETE("/groups/:guid", a.DeleteGroup)

	// Groups Members
	r.GET("/groupMembers/:groupGuid", a.GetGroupMembers)
	r.GET("/groupMembers", a.GetGroupsByNetId)
	r.POST("/groupMembers", a.AddGroupMember)
	r.DELETE("/groupMembers/:netId/:groupId", a.RemoveGroupMember)
	r.DELETE("/groupMembers/:netId", a.RemoveFromAllGroups)

	// General response for the Cross-Origin OPTIONS preflight request
	r.Register("OPTIONS", "/*path", Options)

	// Run the server
	if err := r.Run(":5000"); err != nil {
		fmt.Println(err)
	}
}
