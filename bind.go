/*
Package pbind provides means of binding protocol buffer message types to sql.Rows output.

A common design challange with Go is scanning sql.Rows results into structs.
Ussualy, one would scan into the indivudual fields or local variables.
But what if the the selected columns are a variable in the application?
In such cases developers have to resort to reflection, an ORM based on reflection or
a code generator like SQLBoiler with tons of adaptor code.

Pbind is for protocol buffer developers which like to avoid the usage of ORMs or reflection.
It uses protoreflect, which uses the embedded protobuf descriptors in the generated Go code.
This should give a potential performance edge over Go reflection.
*/
package pbind

import (
	"database/sql"
	"fmt"

	"google.golang.org/protobuf/proto"
	pr "google.golang.org/protobuf/reflect/protoreflect"
)

// field holds reference to a protoreflect.Message and FieldDescriptor of that Message.
type field struct {
	msg  pr.Message
	desc pr.FieldDescriptor
}

// Scan implements sql.Scanner
func (f *field) Scan(src interface{}) error {
	var (
		v  pr.Value
		ok bool
	)

	kind := f.desc.Kind()

	switch kind {
	case pr.BoolKind:

		var d bool
		d, ok = src.(bool)
		v = pr.ValueOfBool(d)

	case pr.Int32Kind, pr.Sint32Kind, pr.Sfixed32Kind:

		var d int64
		d, ok = src.(int64)
		v = pr.ValueOfInt32(int32(d))

	case pr.Uint32Kind, pr.Fixed32Kind:

		var d int64
		d, ok = src.(int64)
		v = pr.ValueOfUint32(uint32(d))

	case pr.Int64Kind, pr.Sint64Kind, pr.Sfixed64Kind:

		var d int64
		d, ok = src.(int64)
		v = pr.ValueOfInt64(d)

	case pr.Uint64Kind, pr.Fixed64Kind:

		var d int64
		d, ok = src.(int64)
		v = pr.ValueOfUint64(uint64(d))

	case pr.FloatKind:

		var d float64
		d, ok = src.(float64)
		v = pr.ValueOfFloat32(float32(d))

	case pr.DoubleKind:

		var d float64
		d, ok = src.(float64)
		v = pr.ValueOfFloat64(d)

	case pr.StringKind:

		var d string
		d, ok = src.(string)
		v = pr.ValueOfString(d)

	case pr.BytesKind:

		var s []byte
		s, ok = src.([]byte)

		d := make([]byte, len(s))
		copy(d, s)

		v = pr.ValueOfBytes(s)

	default:
		return fmt.Errorf("unsupported type %q for scanning", kind)
	}

	if !ok {
		return fmt.Errorf("cannot scan %T into %s", src, kind)
	}

	f.msg.Set(f.desc, v)

	return nil
}

// Scan the values from a Rows iteration into a protobuf message.
// The caller must call rows.Next() before each call of Scan.
//
// Column (or alias) names are matched with the proto Message Field Names.
// If a column name cannot be matched to a Field Name, Scan returns an error.
//
// All (and only) scalar Field types are supported as defined at:
// https://developers.google.com/protocol-buffers/docs/proto3#scalar.
func Scan(rows *sql.Rows, msg proto.Message) error {
	rm := msg.ProtoReflect()

	fields := rm.Descriptor().Fields()

	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("pbbind Bind: %w", err)
	}

	dest := make([]interface{}, len(cols))

	for i, col := range cols {
		fds := fields.ByName(pr.Name(col))
		if fds == nil {
			return fmt.Errorf("pbbind Bind: %q field not in Message", col)
		}

		dest[i] = &field{
			msg:  rm,
			desc: fields.ByName(pr.Name(col)),
		}
	}

	if err := rows.Scan(dest...); err != nil {
		return fmt.Errorf("pbbind Scan: %w", err)
	}

	return nil
}
