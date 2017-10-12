/*
 * Copyright 2016 Dgraph Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"testing"

	"gopkg.in/adibiarsotp/dgraph.v83/protos"
	"gopkg.in/adibiarsotp/dgraph.v83/x"
	"github.com/stretchr/testify/assert"
)

func graphValue(x string) *protos.Value {
	return &protos.Value{&protos.Value_StrVal{x}}
}

func TestSetMutation(t *testing.T) {
	req := Req{}

	s := graphValue("Alice")
	err := req.Set(Edge{protos.NQuad{
		Subject:     "alice",
		Predicate:   "name",
		ObjectValue: s,
	}})
	x.Check(err)

	s = graphValue("rabbithole")
	err = req.Set(Edge{protos.NQuad{
		Subject:     "alice",
		Predicate:   "falls.in",
		ObjectValue: s,
	}})
	x.Check(err)

	err = req.Delete(Edge{protos.NQuad{
		Subject:     "alice",
		Predicate:   "falls.in",
		ObjectValue: s,
	}})
	x.Check(err)

	assert.Equal(t, len(req.gr.Mutation.Set), 2, "Set should have 2 entries")
	assert.Equal(t, len(req.gr.Mutation.Del), 1, "Del should have 1 entry")
}

func TestAddSchema(t *testing.T) {
	req := Req{}

	if err := req.AddSchema(protos.SchemaUpdate{Predicate: "name"}); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(req.gr.Mutation.Schema), 1)
}
