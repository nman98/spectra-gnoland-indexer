package schema

import (
	"reflect"
	"testing"
)

// rowEncoder is satisfied by every struct that is inserted via COPY FROM.
type rowEncoder interface {
	TableName() string
	TableColumns() []string
	CopyRow() []any
}

// allRowEncoders lists every struct inserted through the generic COPY path.
// Add new message/table types here so the alignment guard covers them.
func allRowEncoders() []rowEncoder {
	return []rowEncoder{
		Blocks{},
		ValidatorBlockSigning{},
		TransactionGeneral{},
		AddressTx{},
		MsgSend{},
		MsgMultiSend{},
		MsgCall{},
		MsgAddPackage{},
		MsgRun{},
		MsgAuthCrSession{},
		MsgAuthRvSession{},
		MsgAuthRvAllSessions{},
	}
}

// TestCopyRowMatchesColumns ensures CopyRow stays aligned with the struct's
// fields (and therefore with TableColumns, which is derived from those fields
// in declaration order). A field reorder/insert that is not mirrored in CopyRow
// would otherwise load values into the wrong columns silently during COPY FROM.
func TestCopyRowMatchesColumns(t *testing.T) {
	for _, enc := range allRowEncoders() {
		t.Run(enc.TableName(), func(t *testing.T) {
			testEncoder(t, enc)
		})
	}
}

func testEncoder(
	t *testing.T,
	enc rowEncoder,
) {
	st := reflect.TypeOf(enc)
	cols := enc.TableColumns()
	row := enc.CopyRow()

	if st.NumField() != len(cols) {
		t.Fatalf("field count %d != TableColumns length %d", st.NumField(), len(cols))
	}
	if len(row) != len(cols) {
		t.Fatalf("CopyRow length %d != TableColumns length %d", len(row), len(cols))
	}

	testValidateFields(t, st, cols, row)
}

func testValidateFields(
	t *testing.T,
	st reflect.Type,
	cols []string,
	row []any,
) {
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		got := reflect.TypeOf(row[i])
		if got == nil {
			t.Fatalf("column %q (field %s): CopyRow returned untyped nil", cols[i], field.Name)
		}

		// Slice fields (except []byte) are wrapped in pgtype.Array[T] by pgArray.
		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() != reflect.Uint8 {
			elems, ok := got.FieldByName("Elements")
			if !ok || elems.Type != field.Type {
				t.Fatalf("column %q (field %s %s): expected pgtype.Array wrapping %s, got %s",
					cols[i], field.Name, field.Type, field.Type, got)
			}
			continue
		}

		if got != field.Type {
			t.Fatalf("column %q (field %s): CopyRow value type %s != field type %s",
				cols[i], field.Name, got, field.Type)
		}
	}
}
