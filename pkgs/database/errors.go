package database

import "errors"

// ErrNotFound is returned by query functions when the database ran
// successfully but produced no matching row(s). Callers (especially the API
// handlers) should treat this as a 404-equivalent condition rather than a
// server-side failure.
//
// Wrap it with fmt.Errorf("...: %w", ErrNotFound) when extra context is
// useful; callers use errors.Is(err, ErrNotFound) to detect it.
var ErrNotFound = errors.New("not found")
