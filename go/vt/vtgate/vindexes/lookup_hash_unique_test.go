// Copyright 2014, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vindexes

import (
	"reflect"
	"testing"

	mproto "github.com/youtube/vitess/go/mysql/proto"
	"github.com/youtube/vitess/go/sqltypes"
	"github.com/youtube/vitess/go/vt/key"
	tproto "github.com/youtube/vitess/go/vt/tabletserver/proto"
	"github.com/youtube/vitess/go/vt/vtgate/planbuilder"
)

var lhu planbuilder.Vindex

func init() {
	h, err := planbuilder.CreateVindex("lookup_hash_unique", map[string]interface{}{"Table": "t", "From": "fromc", "To": "toc"})
	if err != nil {
		panic(err)
	}
	lhu = h
}

func TestLookupHashUniqueCost(t *testing.T) {
	if lhu.Cost() != 10 {
		t.Errorf("Cost(): %d, want 10", lhu.Cost())
	}
}

func TestLookupHashUniqueMap(t *testing.T) {
	vc := &vcursor{numRows: 1}
	got, err := lhu.(planbuilder.Unique).Map(vc, []interface{}{1, int32(2)})
	if err != nil {
		t.Error(err)
	}
	want := []key.KeyspaceId{
		"\x16k@\xb4J\xbaK\xd6",
		"\x16k@\xb4J\xbaK\xd6",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map(): %#v, want %+v", got, want)
	}
}

func TestLookupHashUniqueMapNomatch(t *testing.T) {
	vc := &vcursor{}
	got, err := lhu.(planbuilder.Unique).Map(vc, []interface{}{1, int32(2)})
	if err != nil {
		t.Error(err)
	}
	want := []key.KeyspaceId{"", ""}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map(): %#v, want %+v", got, want)
	}
}

func TestLookupHashUniqueMapFail(t *testing.T) {
	vc := &vcursor{mustFail: true}
	_, err := lhu.(planbuilder.Unique).Map(vc, []interface{}{1, int32(2)})
	want := "lookup.Map: Execute failed"
	if err == nil || err.Error() != want {
		t.Errorf("lhu.Map: %v, want %v", err, want)
	}
}

func TestLookupHashUniqueMapBadData(t *testing.T) {
	result := &mproto.QueryResult{
		Fields: []mproto.Field{{
			Type: mproto.VT_INT24,
		}},
		Rows: [][]sqltypes.Value{
			[]sqltypes.Value{
				sqltypes.MakeFractional([]byte("1.1")),
			},
		},
		RowsAffected: 1,
	}
	vc := &vcursor{result: result}
	_, err := lhu.(planbuilder.Unique).Map(vc, []interface{}{1, int32(2)})
	want := `lookup.Map: strconv.ParseUint: parsing "1.1": invalid syntax`
	if err == nil || err.Error() != want {
		t.Errorf("lhu.Map: %v, want %v", err, want)
	}

	result.Fields = []mproto.Field{{
		Type: mproto.VT_FLOAT,
	}}
	vc = &vcursor{result: result}
	_, err = lhu.(planbuilder.Unique).Map(vc, []interface{}{1, int32(2)})
	want = `lookup.Map: unexpected type for 1.1: float64`
	if err == nil || err.Error() != want {
		t.Errorf("lhu.Map: %v, want %v", err, want)
	}

	vc = &vcursor{numRows: 2}
	_, err = lhu.(planbuilder.Unique).Map(vc, []interface{}{1, int32(2)})
	want = `lookup.Map: unexpected multiple results from vindex t: 1`
	if err == nil || err.Error() != want {
		t.Errorf("lhu.Map: %v, want %v", err, want)
	}
}

func TestLookupHashUniqueVerify(t *testing.T) {
	vc := &vcursor{numRows: 1}
	success, err := lhu.Verify(vc, 1, "\x16k@\xb4J\xbaK\xd6")
	if err != nil {
		t.Error(err)
	}
	if !success {
		t.Errorf("Verify(): %+v, want true", success)
	}
}

func TestLookupHashUniqueVerifyNomatch(t *testing.T) {
	vc := &vcursor{}
	success, err := lhu.Verify(vc, 1, "\x16k@\xb4J\xbaK\xd6")
	if err != nil {
		t.Error(err)
	}
	if success {
		t.Errorf("Verify(): %+v, want false", success)
	}
}

func TestLookupHashUniqueVerifyFail(t *testing.T) {
	vc := &vcursor{mustFail: true}
	_, err := lhu.Verify(vc, 1, "\x16k@\xb4J\xbaK\xd6")
	want := "lookup.Verify: Execute failed"
	if err == nil || err.Error() != want {
		t.Errorf("lhu.Verify: %v, want %v", err, want)
	}
}

func TestLookupHashUniqueCreate(t *testing.T) {
	vc := &vcursor{}
	err := lhu.(planbuilder.Lookup).Create(vc, 1, "\x16k@\xb4J\xbaK\xd6")
	if err != nil {
		t.Error(err)
	}
	wantQuery := &tproto.BoundQuery{
		Sql: "insert into t(fromc, toc) values(:fromc, :toc)",
		BindVariables: map[string]interface{}{
			"fromc": 1,
			"toc":   int64(1),
		},
	}
	if !reflect.DeepEqual(vc.query, wantQuery) {
		t.Errorf("vc.query = %#v, want %#v", vc.query, wantQuery)
	}
}

func TestLookupHashUniqueCreateFail(t *testing.T) {
	vc := &vcursor{mustFail: true}
	err := lhu.(planbuilder.Lookup).Create(vc, 1, "\x16k@\xb4J\xbaK\xd6")
	want := "lookup.Create: Execute failed"
	if err == nil || err.Error() != want {
		t.Errorf("lhu.Create: %v, want %v", err, want)
	}
}

func TestLookupHashUniqueGenerate(t *testing.T) {
	_, ok := lhu.(planbuilder.LookupGenerator)
	if ok {
		t.Errorf("lhu.(planbuilder.LookupGenerator): true, want false")
	}
}

func TestLookupHashUniqueDelete(t *testing.T) {
	vc := &vcursor{}
	err := lhu.(planbuilder.Lookup).Delete(vc, []interface{}{1}, "\x16k@\xb4J\xbaK\xd6")
	if err != nil {
		t.Error(err)
	}
	wantQuery := &tproto.BoundQuery{
		Sql: "delete from t where fromc in ::fromc and toc = :toc",
		BindVariables: map[string]interface{}{
			"fromc": []interface{}{1},
			"toc":   int64(1),
		},
	}
	if !reflect.DeepEqual(vc.query, wantQuery) {
		t.Errorf("vc.query = %#v, want %#v", vc.query, wantQuery)
	}
}

func TestLookupHashUniqueDeleteFail(t *testing.T) {
	vc := &vcursor{mustFail: true}
	err := lhu.(planbuilder.Lookup).Delete(vc, []interface{}{1}, "\x16k@\xb4J\xbaK\xd6")
	want := "lookup.Delete: Execute failed"
	if err == nil || err.Error() != want {
		t.Errorf("lhu.Delete: %v, want %v", err, want)
	}
}
