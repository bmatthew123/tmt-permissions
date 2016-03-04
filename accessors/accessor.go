package accessors

import (
	"encoding/json"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
)

// Struct to model the response from the guid generator micro-service.
type guid struct {
	Status string
	Data   string
}

var NewGuid = func() string {
	var newGuid guid

	// Request new guid
	res, err := http.Get("http://tmt-guid.byu.edu/guid")
	// Check for errors, if so, create a new guid
	if err != nil {
		guid := uuid.NewV4()
		return guid.String()
	}
	body, err := ioutil.ReadAll(res.Body)
	// Check for errors, if so, create a new guid
	if err != nil {
		guid := uuid.NewV4()
		return guid.String()
	}
	// Parse data, if an error occurred, create a new guid
	if err := json.Unmarshal(body, &newGuid); err != nil || newGuid.Status != "OK" {
		guid := uuid.NewV4()
		return guid.String()
	}

	// Guid retrieved succesfully, return it
	return newGuid.Data
}
