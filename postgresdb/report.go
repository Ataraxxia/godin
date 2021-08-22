package postgresdb

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	rep "github.com/Ataraxxia/godin/Report"
)

type MyReport rep.Report

// Make the Attrs struct implement the driver.Valuer interface. This method
// simply returns the JSON-encoded representation of the struct.
func (r MyReport) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Make the Attrs struct implement the sql.Scanner interface. This method
// simply decodes a JSON-encoded value into the struct fields.
func (r *MyReport) Scan(value interface{}) error {
	b, err := value.([]byte)
	if !err {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &r)
}
