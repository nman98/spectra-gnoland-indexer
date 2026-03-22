package date

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Date struct {
	time.Time
}

func (d *Date) UnmarshalText(data []byte) error {
	t, err := time.Parse("2006-01-02", string(data))
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}
	d.Time = t
	return nil
}

func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.Format("2006-01-02")), nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Format("2006-01-02") + `"`), nil
}

func (d *Date) Scan(src any) error {
	if src == nil {
		d.Time = time.Time{}
		return nil
	}
	switch v := src.(type) {
	case time.Time:
		d.Time = v
		return nil
	case string:
		return d.UnmarshalText([]byte(v))
	case []byte:
		return d.UnmarshalText(v)
	default:
		return fmt.Errorf("cannot scan type %T into date.Date", src)
	}
}

// Value implements driver.Valuer so pgx encodes date.Date as YYYY-MM-DD string,
// not as a full timestamp. This ensures date parameters work correctly with
// PostgreSQL date or timestamp columns.
func (d Date) Value() (driver.Value, error) {
	return d.Format("2006-01-02"), nil
}
