package schema

import (
	"fmt"
	"reflect"
	"testing"
)

// allowedTagKeys is the set of struct-tag keys the DDL generator understands.
// "default" is currently read by nothing (defaults are baked into the dbtype
// string) but is tolerated here; everything else is a typo.
var allowedTagKeys = map[string]bool{
	"db":       true,
	"dbtype":   true,
	"nullable": true,
	"primary":  true,
	"unique":   true,
	"default":  true,
}

// boolTagKeys must hold exactly "true" or "false" when present, so a typo'd value
// (e.g. primary:"ture") can't silently flip a column's DDL semantics.
var boolTagKeys = []string{"nullable", "primary", "unique"}

// TestTableTagsAreValid guards the reflection-driven DDL layer: a mistyped tag key
// or value would otherwise be silently ignored and produce a wrong schema. It runs
// over every table in the canonical AllTables list, so new tables are covered
// automatically.
func TestTableTagsAreValid(t *testing.T) {
	for _, table := range AllTables() {
		rt := reflect.TypeOf(table)
		t.Run(table.TableName(), func(t *testing.T) {
			if rt.Kind() != reflect.Struct {
				t.Fatalf("expected struct, got %s", rt.Kind())
			}

			seenColumns := make(map[string]string) // db name -> field name
			primaryKeys := 0

			for field := range rt.Fields() {
				if !field.IsExported() {
					continue
				}

				// Every persisted field must declare a column name and a SQL type.
				dbName, hasDB := field.Tag.Lookup("db")
				if !hasDB || dbName == "" || dbName == "-" {
					t.Errorf("field %s: missing or empty `db` tag", field.Name)
					continue
				}
				if dbType, hasType := field.Tag.Lookup("dbtype"); !hasType || dbType == "" {
					t.Errorf("field %s (column %q): missing or empty `dbtype` tag", field.Name, dbName)
				}

				// Reject unknown tag keys (catches typos like `dbtyp`/`primry`).
				for _, key := range structTagKeys(string(field.Tag)) {
					if !allowedTagKeys[key] {
						t.Errorf("field %s (column %q): unknown struct-tag key %q", field.Name, dbName, key)
					}
				}

				// Boolean-valued tags, when present, must be exactly true/false.
				for _, key := range boolTagKeys {
					if v, ok := field.Tag.Lookup(key); ok && v != "true" && v != "false" {
						t.Errorf("field %s (column %q): tag %q = %q, expected \"true\" or \"false\"",
							field.Name, dbName, key, v)
					}
				}

				if prev, dup := seenColumns[dbName]; dup {
					t.Errorf("duplicate column %q on fields %s and %s", dbName, prev, field.Name)
				}
				seenColumns[dbName] = field.Name

				if field.Tag.Get("primary") == "true" {
					primaryKeys++
				}
			}

			if primaryKeys == 0 {
				t.Errorf("table %q declares no primary key", table.TableName())
			}

			// GetTableInfo must build cleanly (it panics on a missing db tag).
			if err := safeGetTableInfo(table); err != nil {
				t.Errorf("GetTableInfo failed: %v", err)
			}
		})
	}
}

// TestTableListsAreDisjoint ensures a table is not listed as both a regular table
// and a hypertable, which would try to create it twice during setup.
func TestTableListsAreDisjoint(t *testing.T) {
	seen := make(map[string]string) // table name -> list it came from
	add := func(name, list string) {
		if prev, ok := seen[name]; ok {
			t.Errorf("table %q appears in both %s and %s", name, prev, list)
		}
		seen[name] = list
	}
	for _, tbl := range RegularTables() {
		add(tbl.TableName(), "RegularTables")
	}
	for _, h := range Hypertables() {
		add(h.Table.TableName(), "Hypertables")
	}
}

// safeGetTableInfo calls GetTableInfo, converting a panic into an error so the
// test reports it instead of crashing.
func safeGetTableInfo(table DBTable) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	_, err = table.GetTableInfo()
	return err
}

// structTagKeys returns the keys present in a struct tag. It mirrors the parsing
// loop in reflect.StructTag.Lookup, collecting names rather than a single value.
func structTagKeys(tag string) []string {
	var keys []string
	for tag != "" {
		// Skip leading spaces.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}
		// Scan to the colon that ends the key.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		keys = append(keys, tag[:i])
		tag = tag[i+1:]
		// Scan past the quoted value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		tag = tag[i+1:]
	}
	return keys
}
