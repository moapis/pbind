// Copyright (c) 2021, Mohlmann Solutions SRL. All rights reserved.
// Use of this source code is governed by a License that can be found in the LICENSE file.
// SPDX-License-Identifier: BSD-3-Clause

package pbind

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/moapis/pbind/test/pb"
	"google.golang.org/protobuf/proto"
)

func TestScan(t *testing.T) {
	type args struct {
		columns []string
		values  []driver.Value
		scanErr bool
	}
	tests := []struct {
		name    string
		args    args
		want    *pb.Msg
		wantErr bool
	}{
		{
			"Scan error",
			args{
				[]string{"text", "number", "large"},
				[]driver.Value{"foo", 123, 20000000},
				true,
			},
			&pb.Msg{},
			true,
		},
		{
			"Field not present",
			args{
				[]string{"not", "number", "large"},
				[]driver.Value{"foo", 123, 20000000},
				false,
			},
			&pb.Msg{},
			true,
		},
		{
			"Type mismatch",
			args{
				[]string{"text", "number", "large"},
				[]driver.Value{900, 123, 20000000},
				false,
			},
			&pb.Msg{},
			true,
		},
		{
			"Unsupported type",
			args{
				[]string{"mp"},
				[]driver.Value{900},
				false,
			},
			&pb.Msg{},
			true,
		},
		{
			"All",
			args{
				[]string{
					"toggle",
					"number", "snumber", "unumber",
					"large", "slarge", "ularge",
					"sfloat", "lfloat",
					"text", "bin",
				},
				[]driver.Value{
					true,
					123, 456, 789,
					20000000, 30000000, 40000000,
					33.33, 300.0003,
					"Hello world!", []byte("Foo Bar"),
				},
				false,
			},
			&pb.Msg{
				Toggle:  true,
				Number:  123,
				Snumber: 456,
				Unumber: 789,
				Large:   20000000,
				Slarge:  30000000,
				Ularge:  40000000,
				Sfloat:  33.33,
				Lfloat:  300.0003,
				Text:    "Hello world!",
				Bin:     []byte("Foo Bar"),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sm, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			smrows := sqlmock.NewRows(tt.args.columns).AddRow(tt.args.values...)
			if tt.args.scanErr {
				smrows = smrows.RowError(0, errors.New("Row error"))
			}

			sm.ExpectQuery("select foo").WillReturnRows(smrows).RowsWillBeClosed()
			sm.ExpectClose()

			rows, err := db.Query("select foo")
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()

			var msg pb.Msg

			rows.Next()

			if err := Scan(rows, &msg); (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !proto.Equal(&msg, tt.want) {
				t.Errorf("Scan() =\n%v\nwant\n%v", &msg, tt.want)
			}
		})
	}
}

var db *sql.DB

// Example to scan into the protobuf message as defined in test/pb/bind_test.proto.
func ExampleScan() {
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
