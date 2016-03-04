package accessors

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// Add the log entry with the given type, actor, and data
var Log = func(t, a, d string, logToStdErr bool, db *sql.DB) (bool, error) {
	if logToStdErr {
		log.Println("Type:", t, "Actor:", a, "Data:", d)
	}

	stmt, err := db.Prepare("INSERT INTO log (guid,actor,type,data) VALUES (?,?,?,?)")
	stmt.Exec(NewGuid(), a, t, d)
	stmt.Close()

	if err != nil {
		log.Println("Error inserting into database", "Error:", err.Error(), "Type:", t, "Actor:", a, "Data:", d)
		return false, err
	}

	return true, nil
}
