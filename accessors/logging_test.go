package accessors

import (
	"bytes"
	"io"
	"log"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	testhelpers "github.com/byu-oit-ssengineering/tmt-test-helpers"
)

func TestLogging(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a logging accessor %v", err)
		return
	}

	sqlmock.ExpectExec("INSERT INTO log (guid,actor,type,data) VALUES (.,.,.,.)").
		WithArgs("11111111-2222-3333-2222-111111111111", "11111111-2222-3333-4444-555555555555", "error", "This is a test log message").
		WillReturnResult(sqlmock.NewResult(1, 1))

	old := os.Stderr // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stderr = w

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// Call function that is being tested
	log.SetOutput(os.Stderr)
	Log("error", "system", "Error that does log to stderr", true, db)

	// back to normal state
	w.Close()
	os.Stderr = old // restoring the real stdout
	out := <-outC

	// reading our temp stdout
	if out[20:] != "Type: error Actor: system Data: Error that does log to stderr\n" {
		t.Errorf("Expected %v", out[20:])
		return
	}
}

func TestLoggingNoStderr(t *testing.T) {
	db, err := testhelpers.GetMockDB()
	if err != nil {
		t.Error("An unexpected error occurred when creating a logging accessor %v", err)
		return
	}

	sqlmock.ExpectExec("INSERT INTO log (guid,actor,type,data) VALUES (.,.,.,.)").
		WithArgs("11111111-2222-3333-2222-111111111111", "11111111-2222-3333-4444-555555555555", "error", "This is a test log message").
		WillReturnResult(sqlmock.NewResult(1, 1))

	old := os.Stderr // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stderr = w

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// Call function that is being tested
	log.SetOutput(os.Stderr)
	Log("error", "system", "Error that does log to stderr", false, db)

	// back to normal state
	w.Close()
	os.Stderr = old // restoring the real stdout
	out := <-outC

	// reading our temp stdout
	if out != "" {
		t.Error("Expected an empty string")
		return
	}
}
