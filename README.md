[![Build Status](https://travis-ci.org/moapis/pbind.svg?branch=main)](https://travis-ci.org/moapis/pbind)
[![codecov](https://codecov.io/gh/moapis/pbind/branch/main/graph/badge.svg)](https://codecov.io/gh/moapis/pbind)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/moapis/pbind)](https://pkg.go.dev/github.com/moapis/pbind)
[![Go Report Card](https://goreportcard.com/badge/github.com/moapis/pbind)](https://goreportcard.com/report/github.com/moapis/pbind)

# PBIND

Package pbind provides means of binding protocol buffer message types to sql.Rows output.

A common design challange with Go is scanning sql.Rows results into structs. Ussualy, one would scan into the indivudual fields or local variables. But what if the the selected columns are a variable in the application? In such cases developers have to resort to reflection, an ORM based on reflection or a code generator like SQLBoiler with tons of adaptor code.

Pbind is for protocol buffer developers which like to avoid the usage of ORMs or reflection. It uses protoreflect, which uses the embedded protobuf descriptors in the generated Go code. This should give a potential performance edge over Go reflection.

## Example

````
func query() []*pb.Msg {
    // Scanning will match the probuf field names (case sensitive!).
    // If the column name does not match the field name, use an alias.
    const query = "select something as toggle, number, snumber, unumber, large, slarge, ularge, sfloat, lfloat, text, bin from table;"

    rows, err := db.Query(query)
    if err != nil {
        panic(err)
    }
    defer rows.Close()

    var msgs []*pb.Msg

    for rows.Next() {
        msg := new(pb.Msg)

        if err := Scan(rows, msg); err != nil {
            panic(err)
        }

        msgs = append(msgs, msg)
    }
}
````

## Copyright and license

Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
Use of this source code is governed by a License that can be found in the [LICENSE](LICENSE) file.
