package handlers

import (
	"errors"
	"log"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/danielgtaylor/huma/v2"
)

// The helpers below are the only way handlers should build HTTP errors. The
// goal is to consistently:
//
//   - Return 4xx with a plain, user-facing message when the caller did
//     something wrong (missing/invalid params, malformed input, etc.).
//   - Return 404 when the database worked correctly but produced no row(s).
//   - Return a generic 500 when anything unexpected happens (SQL errors,
//     decode failures, ...). The underlying error is logged server-side so we
//     can debug it, but it is NEVER serialised into the response body — that
//     way we don't leak query strings, schema hints, stack traces, etc.
//
// All helpers take a short operation tag (typically the handler method name)
// so the log lines are easy to grep.

// badRequest returns a 400 with just the user-facing message. The underlying
// validation error (if any) is intentionally dropped because it does not add
// anything useful for the client beyond the message.
func badRequest(msg string) error {
	return huma.Error400BadRequest(msg)
}

// notFound returns a 404 with just the user-facing message. It is the right
// response when the database ran fine but returned no data.
func notFound(msg string) error {
	return huma.Error404NotFound(msg)
}

// internalError logs the real error and returns a generic 500 that carries
// no implementation detail. Use it for every unexpected error path (SQL
// failures, marshalling issues, cursor encoding, ...).
func internalError(op string, err error) error {
	log.Printf("api %s: internal error: %v", op, err)
	return huma.Error500InternalServerError("internal server error")
}

// mapDbError turns a database-layer error into an HTTP error: database.ErrNotFound
// becomes a 404 with the supplied user-facing message, everything else is logged
// and masked as a 500.
func mapDbError(op, notFoundMsg string, err error) error {
	if errors.Is(err, database.ErrNotFound) {
		return notFound(notFoundMsg)
	}
	return internalError(op, err)
}
