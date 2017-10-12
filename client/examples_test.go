/*
 * Copyright (C) 2017 Dgraph Labs, Inc. and Contributors
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

package client_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"gopkg.in/adibiarsotp/dgraph.v83/client"
	"gopkg.in/adibiarsotp/dgraph.v83/x"
	"github.com/gogo/protobuf/proto"
	"github.com/twpayne/go-geom/encoding/wkb"
	"google.golang.org/grpc"
)

func Node(val string, c *client.Dgraph) string {
	if uid, err := strconv.ParseUint(val, 0, 64); err == nil {
		return c.NodeUid(uid).String()
	}
	if strings.HasPrefix(val, "_:") {
		n, err := c.NodeBlank(val[2:])
		if err != nil {
			log.Fatalf("Error while converting to node: %v", err)
		}
		return n.String()
	}
	n, err := c.NodeXid(val, false)
	if err != nil {
		log.Fatalf("Error while converting to node: %v", err)
	}
	return n.String()
}

func ExampleReq_Set() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	x.Check(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient(
		[]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)

	// Create new request
	req := client.Req{}

	// Create a node for person1 (the blank node label "person1" exists
	// client-side so the mutation can correctly link nodes.  It is not
	// persisted in the server)
	person1, err := dgraphClient.NodeBlank("person1")
	if err != nil {
		log.Fatal(err)
	}

	// Add edges for name and salary to person1
	e := person1.Edge("name")
	e.SetValueString("Steven Spielberg")
	err = req.Set(e)
	x.Check(err)

	// If the old variable was written over or out of scope we can lookup person1 again,
	// the string->node mapping is remembered by the client for this session.
	p, err := dgraphClient.NodeBlank("person1")
	e = p.Edge("salary")
	e.SetValueFloat(13333.6161)
	err = req.Set(e)
	x.Check(err)

	resp, err := dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error in getting response from server, %s", err)
	}

	// proto.MarshalTextString(resp) can be used to print the raw response as text.  Client
	// programs usually use Umarshal to unpack query responses to a struct (or the protocol
	// buffer can be accessed with resp.N)
	fmt.Printf("%+v\n", proto.MarshalTextString(resp))

	err = dgraphClient.Close()
	x.Check(err)
}

func ExampleReq_Delete() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	x.Check(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient(
		[]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)

	// Create new request
	req := client.Req{}

	// Create a node for person1 (the blank node label "person1" exists
	// client-side so the mutation can correctly link nodes.  It is not
	// persisted in the server)
	person1, err := dgraphClient.NodeBlank("person1")
	if err != nil {
		log.Fatal(err)
	}
	person2, err := dgraphClient.NodeBlank("person2")
	if err != nil {
		log.Fatal(err)
	}

	e := person1.Edge("name")
	e.SetValueString("Steven Spallding")
	err = req.Set(e)
	x.Check(err)

	e = person2.Edge("name")
	e.SetValueString("Steven Stevenson")
	err = req.Set(e)
	x.Check(err)

	e = person1.ConnectTo("friend", person2)

	// Add person1, person2 and friend edge to store
	resp, err := dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error in getting response from server, %s", err)
	}
	fmt.Printf("%+v\n", proto.MarshalTextString(resp))

	// Now remove the friend edge

	// If the old variable was written over or out of scope we can lookup person1 again,
	// the string->node mapping is remembered by the client for this session.
	p1, err := dgraphClient.NodeBlank("person1")
	p2, err := dgraphClient.NodeBlank("person2")

	e = p1.ConnectTo("friend", p2)
	req = client.Req{}
	req.Delete(e)

	// Run the mutation to delete the edge
	resp, err = dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error in getting response from server, %s", err)
	}
	fmt.Printf("%+v\n", proto.MarshalTextString(resp))
	err = dgraphClient.Close()
	x.Check(err)
}

func ExampleDgraph_BatchSet() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	bmOpts := client.BatchMutationOptions{
		Size:          1000,
		Pending:       100,
		PrintCounters: false,
	}
	clientDir, err := ioutil.TempDir("", "client_")
	x.Check(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient([]*grpc.ClientConn{conn}, bmOpts, clientDir)

	// Create a node for person1 (the blank node label "person1" exists
	// client-side so the mutation can correctly link nodes.  It is not
	// persisted in the server)
	person1, err := dgraphClient.NodeBlank("person1")
	if err != nil {
		log.Fatal(err)
	}

	// Add edges for name and salary to the batch mutation
	e := person1.Edge("name")
	e.SetValueString("Steven Spielberg")
	dgraphClient.BatchSet(e)
	e = person1.Edge("salary")
	e.SetValueFloat(13333.6161)
	dgraphClient.BatchSet(e)

	dgraphClient.BatchFlush() // Must be called to flush buffers after all mutations are added.
	err = dgraphClient.Close()
	x.Check(err)
}

func ExampleEdge_AddFacet() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	x.Check(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient(
		[]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)

	req := client.Req{}

	// Create a node for person1 add an edge for name.
	person1, err := dgraphClient.NodeXid("person1", false)
	if err != nil {
		log.Fatal(err)
	}
	e := person1.Edge("name")
	e.SetValueString("Steven Stevenson")

	// Add facets since and alias to the edge.
	e.AddFacet("since", "2006-01-02T15:04:05")
	e.AddFacet("alias", `"Steve"`)

	err = req.Set(e)
	x.Check(err)

	person2, err := dgraphClient.NodeXid("person2", false)
	if err != nil {
		log.Fatal(err)
	}
	e = person2.Edge("name")
	e.SetValueString("William Jones")
	err = req.Set(e)
	x.Check(err)

	e = person1.ConnectTo("friend", person2)

	// Facet on a node-node edge.
	e.AddFacet("close", "true")
	err = req.Set(e)
	x.Check(err)

	req.SetQuery(`mutation { schema { name: string @index(exact) . } }`)

	resp, err := dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error in getting response from server, %s", err)
	}

	req = client.Req{}
	req.SetQuery(`{
		query(func: eq(name,"Steven Stevenson")) {
			name @facets
			friend @facets {
				name
			}
		}
	}`)

	resp, err = dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error in getting response from server, %s", err)
	}

	// Types representing information in the graph.
	type nameFacets struct {
		Since time.Time `json:"since"`
		Alias string    `json:"alias"`
	}

	type friendFacets struct {
		Close bool `json:"close"`
	}

	type Person struct {
		Name         string       `json:"name"`
		NameFacets   nameFacets   `json:"name@facets"`
		Friends      []Person     `json:"friend"`
		FriendFacets friendFacets `json:"@facets"`
	}

	// Helper type to unmarshal query
	type Res struct {
		Root Person `json:"query"`
	}

	var pq Res
	err = client.Unmarshal(resp.N, &pq)
	if err != nil {
		log.Fatal("Couldn't unmarshal response : ", err)
	}

	fmt.Println("Found : ", pq.Root.Name)
	fmt.Println("Who likes to be called : ", pq.Root.NameFacets.Alias,
		" since ", pq.Root.NameFacets.Since)
	fmt.Println("Friends : ")
	for i := range pq.Root.Friends {
		fmt.Print("\t", pq.Root.Friends[i].Name)
		if pq.Root.Friends[i].FriendFacets.Close {
			fmt.Println(" who is a close friend.")
		} else {
			fmt.Println(" who is not a close friend.")
		}
	}
	err = dgraphClient.Close()
	x.Check(err)
}

func ExampleReq_SetQuery() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	x.Check(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient(
		[]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)

	req := client.Req{}
	alice, err := dgraphClient.NodeXid("alice", false)
	if err != nil {
		log.Fatal(err)
	}
	e := alice.Edge("name")
	e.SetValueString("Alice")
	err = req.Set(e)
	x.Check(err)

	e = alice.Edge("falls.in")
	e.SetValueString("Rabbit hole")
	err = req.Set(e)
	x.Check(err)

	req.SetQuery(`mutation { schema { name: string @index(exact) . } }`)
	resp, err := dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error in getting response from server, %s", err)
	}

	req = client.Req{}

	req.SetQuery(`{
		me(func: eq(name, "Alice")) {
			name
			falls.in
		}
	}`)
	resp, err = dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error in getting response from server, %s", err)
	}

	type Alice struct {
		Name         string `json:"name"`
		WhatHappened string `json:"falls.in"`
	}

	type Res struct {
		Root Alice `json:"me"`
	}

	var r Res
	err = client.Unmarshal(resp.N, &r)
	x.Check(err)
	fmt.Printf("Alice: %+v\n\n", r.Root)
	err = dgraphClient.Close()
	x.Check(err)
}

func ExampleReq_SetQueryWithVariables() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	x.Check(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient(
		[]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)

	req := client.Req{}

	alice, err := dgraphClient.NodeXid("alice", false)
	if err != nil {
		log.Fatal(err)
	}
	e := alice.Edge("name")
	e.SetValueString("Alice")
	err = req.Set(e)
	x.Check(err)

	e = alice.Edge("falls.in")
	e.SetValueString("Rabbit hole")
	err = req.Set(e)
	x.Check(err)

	req.SetQuery(`mutation { schema { name: string @index(exact) . } }`)
	resp, err := dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error in getting response from server, %s", err)
	}

	req = client.Req{}
	variables := make(map[string]string)
	variables["$a"] = "Alice"
	req.SetQueryWithVariables(`{
		me(func: eq(name, $a)) {
			name
			falls.in
		}
	}`, variables)

	resp, err = dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error in getting response from server, %s", err)
	}

	type Alice struct {
		Name         string `json:"name"`
		WhatHappened string `json:"falls.in"`
	}

	type Res struct {
		Root Alice `json:"me"`
	}

	var r Res
	err = client.Unmarshal(resp.N, &r)
	x.Check(err)
	fmt.Printf("Alice: %+v\n\n", r.Root)
	err = dgraphClient.Close()
	x.Check(err)
}

func ExampleDgraph_NodeUidVar() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	x.Check(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient(
		[]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)

	req := client.Req{}

	// Add some data
	alice, err := dgraphClient.NodeXid("alice", false)
	if err != nil {
		log.Fatal(err)
	}
	e := alice.Edge("name")
	e.SetValueString("Alice")
	err = req.Set(e)
	x.Check(err)

	req.SetQuery(`mutation { schema { name: string @index(exact) . } }`)

	resp, err := dgraphClient.Run(context.Background(), &req)

	// New request
	req = client.Req{}

	// Now issue a query and mutation using client interface

	req.SetQuery(`{
    a as var(func: eq(name, "Alice"))
    me(func: uid(a)) {
        name
    }
}`)

	// Get a node for the variable a in the query above.
	n, _ := dgraphClient.NodeUidVar("a")
	e = n.Edge("falls.in")
	e.SetValueString("Rabbit hole")
	err = req.Set(e)
	x.Check(err)

	resp, err = dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error in getting response from server, %s", err)
	}
	fmt.Printf("%+v\n", proto.MarshalTextString(resp))

	// This is equivalent to the single query and mutation
	//
	// {
	//		a as var(func: eq(name, "Alice"))
	//		me(func: uid(a)) {
	//			name
	//		}
	// }
	// mutation { set {
	//		var(a) <falls.in> "Rabbit hole" .
	// }}
	//
	// It's often easier to construct such things with client functions that
	// by manipulating raw strings.
	err = dgraphClient.Close()
	x.Check(err)
}

func ExampleDgraph_DropAll() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	x.Checkf(err, "While creating temp dir")
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient(
		[]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)
	x.Checkf(dgraphClient.DropAll(), "While dropping all")
	x.Checkf(dgraphClient.Close(), "While closing client")
}

func ExampleEdge_SetValueBytes() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	x.Check(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient(
		[]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)

	req := client.Req{}

	alice, err := dgraphClient.NodeBlank("alice")
	if err != nil {
		log.Fatal(err)
	}
	e := alice.Edge("name")
	e.SetValueString("Alice")
	err = req.Set(e)
	x.Check(err)

	e = alice.Edge("somestoredbytes")
	err = e.SetValueBytes([]byte(`\xbd\xb2\x3d\xbc\x20\xe2\x8c\x98`))
	x.Check(err)
	err = req.Set(e)
	x.Check(err)

	req.SetQuery(`mutation {
	schema {
		name: string @index(exact) .
	}
}
{
	q(func: eq(name, "Alice")) {
		name
		somestoredbytes
	}
}`)

	resp, err := dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error in getting response from server, %s", err)
	}

	type Alice struct {
		Name      string `json:"name"`
		ByteValue []byte `json:"somestoredbytes"`
	}

	type Res struct {
		Root Alice `json:"q"`
	}

	var r Res
	err = client.Unmarshal(resp.N, &r)
	x.Check(err)
	fmt.Printf("Alice: %+v\n\n", r.Root)
	err = dgraphClient.Close()
	x.Check(err)
}

func ExampleUnmarshal() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	x.Check(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient(
		[]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)

	req := client.Req{}

	// A mutation as a string, see ExampleReq_NodeUidVar, ExampleReq_SetQuery,
	// etc for examples of mutations using client functions.
	req.SetQuery(`
mutation {
	schema {
		name: string @index .
	}
	set {
		_:person1 <name> "Alex" .
		_:person2 <name> "Beatie" .
		_:person3 <name> "Chris" .

		_:person1 <friend> _:person2 .
		_:person1 <friend> _:person3 .
	}
}
{
	friends(func: eq(name, "Alex")) {
		name
		friend {
			name
		}
	}
}`)

	// Run the request in the Dgraph server.  The mutations are added, then
	// the query is exectuted.
	resp, err := dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error in getting response from server, %s", err)
	}

	// Unmarshal the response into a custom struct

	// A type representing information in the graph.
	type person struct {
		Name    string   `json:"name"`
		Friends []person `json:"friend"`
	}

	// A helper type matching the query root.
	type friends struct {
		Root person `json:"friends"`
	}

	var f friends
	err = client.Unmarshal(resp.N, &f)
	if err != nil {
		log.Fatal("Couldn't unmarshal response : ", err)
	}

	fmt.Println("Name : ", f.Root.Name)
	fmt.Print("Friends : ")
	for _, p := range f.Root.Friends {
		fmt.Print(p.Name, " ")
	}
	fmt.Println()

	err = dgraphClient.Close()
	x.Check(err)
}

func ExampleUnmarshal_facetsUpdate() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	x.Check(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient(
		[]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)

	req := client.Req{}

	req.SetQuery(`
mutation {
	schema {
		name: string @index .
	}
	set {
		_:person1 <name> "Alex" .
		_:person2 <name> "Beatie" .
		_:person3 <name> "Chris" .
		_:person4 <name> "David" .

		_:person1 <friend> _:person2 (close=true).
		_:person1 <friend> _:person3 (close=false).
		_:person1 <friend> _:person4 (close=true).
	}
}
{
	friends(func: eq(name, "Alex")) {
		_uid_
		name
		friend @facets {
			_uid_
			name
		}
	}
}`)

	// Run the request in the Dgraph server.  The mutations are added, then
	// the query is exectuted.
	resp, err := dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error in getting response from server, %s", err)
	}

	// Unmarshal the response into a custom struct

	type friendFacets struct {
		Close bool `json:"close"`
	}

	// A type representing information in the graph.
	type person struct {
		ID           uint64        `json:"_uid_"` // record the UID for our update
		Name         string        `json:"name"`
		Friends      []*person     `json:"friend"` // Unmarshal with pointers to structs
		FriendFacets *friendFacets `json:"@facets"`
	}

	// A helper type matching the query root.
	type friends struct {
		Root person `json:"friends"`
	}

	var f friends
	err = client.Unmarshal(resp.N, &f)
	if err != nil {
		log.Fatal("Couldn't unmarshal response : ", err)
	}

	req = client.Req{}

	// Now update the graph.
	// for the close friends, add the reverse edge and note in a facet when we did this.
	for _, p := range f.Root.Friends {
		if p.FriendFacets.Close {
			n := dgraphClient.NodeUid(p.ID)
			e := n.ConnectTo("friend", dgraphClient.NodeUid(f.Root.ID))
			e.AddFacet("since", time.Now().Format(time.RFC3339))
			req.Set(e)
		}
	}

	resp, err = dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error in getting response from server, %s", err)
	}

	err = dgraphClient.Close()
	x.Check(err)
}

func ExampleEdge_SetValueGeoJson() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	x.Check(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient(
		[]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)

	req := client.Req{}

	alice, err := dgraphClient.NodeBlank("alice")
	if err != nil {
		log.Fatal(err)
	}
	e := alice.Edge("name")
	e.SetValueString("Alice")
	err = req.Set(e)
	x.Check(err)

	e = alice.Edge("loc")
	err = e.SetValueGeoJson(`{"Type":"Point", "Coordinates":[1.1,2.0]}`)
	x.Check(err)
	err = req.Set(e)
	x.Check(err)

	e = alice.Edge("city")
	err = e.SetValueGeoJson(`{
		"Type":"Polygon",
		"Coordinates":[[[0.0,0.0], [2.0,0.0], [2.0, 2.0], [0.0, 2.0], [0.0, 0.0]]]
	}`)
	x.Check(err)
	err = req.Set(e)
	x.Check(err)

	req.SetQuery(`mutation {
	schema {
		name: string @index(exact) .
	}
}
{
	q(func: eq(name, "Alice")) {
		name
		loc
		city
	}
}`)

	resp, err := dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error in getting response from server, %s", err)
	}

	type Alice struct {
		Name string `json:"name"`
		Loc  []byte `json:"loc"`
		City []byte `json:"city"`
	}

	type Res struct {
		Root Alice `json:"q"`
	}

	var r Res
	err = client.Unmarshal(resp.N, &r)
	x.Check(err)
	fmt.Printf("Alice: %+v\n\n", r.Root)
	loc, err := wkb.Unmarshal(r.Root.Loc)
	x.Check(err)
	city, err := wkb.Unmarshal(r.Root.City)
	x.Check(err)

	fmt.Printf("Loc: %+v\n\n", loc)
	fmt.Printf("City: %+v\n\n", city)
	err = dgraphClient.Close()
	x.Check(err)
}

func ExampleReq_SetObject() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	x.Check(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient(
		[]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)

	req := client.Req{}

	type School struct {
		Name string `json:"name@en,omitempty"`
	}

	// If omitempty is not set, then edges with empty values (0 for int/float, "" for string, false
	// for bool) would be created for values not specified explicitly.

	type Person struct {
		Uid      uint64   `json:"_uid_,omitempty"`
		Name     string   `json:"name,omitempty"`
		Age      int      `json:"age,omitempty"`
		Married  bool     `json:"married,omitempty"`
		Raw      []byte   `json:"raw_bytes",omitempty`
		Friends  []Person `json:"friend,omitempty"`
		Location string   `json:"loc,omitempty"`
		School   *School  `json:"school,omitempty"`
	}

	// While setting an object if a struct has a Uid then its properties in the graph are updated
	// else a new node is created.
	// In the example below new nodes for Alice and Charlie and school are created (since they dont
	// have a Uid).  Alice is also connected via the friend edge to an existing node with Uid
	// 1000(Bob).  We also set Name and Age values for this node with Uid 1000.

	loc := `{"type":"Point","coordinates":[1.1,2]}`
	p := Person{
		Name:     "Alice",
		Age:      26,
		Married:  true,
		Location: loc,
		Raw:      []byte("raw_bytes"),
		Friends: []Person{{
			Uid:  1000,
			Name: "Bob",
			Age:  24,
		}, {
			Name: "Charlie",
			Age:  29,
		}},
		School: &School{
			Name: "Crown Public School",
		},
	}

	req.SetSchema(`
		age: int .
		married: bool .
	`)

	err = req.SetObject(&p)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}

	// Assigned uids for nodes which were created would be returned in the resp.AssignedUids map.
	puid := resp.AssignedUids["blank-0"]
	q := fmt.Sprintf(`{
		me(func: uid(%d)) {
			_uid_
			name
			age
			loc
			raw_bytes
			married
			friend {
				_uid_
				name
				age
			}
			school {
				name@en
			}
		}
	}`, puid)

	req = client.Req{}
	req.SetQuery(q)
	resp, err = dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}

	type Root struct {
		Me Person `json:"me"`
	}

	var r Root
	err = client.Unmarshal(resp.N, &r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Me: %+v\n", r.Me)
	// R.Me would be same as the person that we set above.
}

func ExampleReq_SetObject_facets(t *testing.T) {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	x.Check(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient(
		[]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)

	req := client.Req{}

	// This example shows example for SetObject using facets.

	type friendFacet struct {
		Since  time.Time `json:"since"`
		Family string    `json:"family"`
		Age    float64   `json:"age"`
		Close  bool      `json:"close"`
	}

	type nameFacets struct {
		Origin string `json:"origin"`
	}

	type schoolFacet struct {
		Since time.Time `json:"since"`
	}

	type School struct {
		Name   string      `json:"name"`
		Facets schoolFacet `json:"@facets"`
	}

	type Person struct {
		Name       string      `json:"name"`
		NameFacets nameFacets  `json:"name@facets"`
		Facets     friendFacet `json:"@facets"`
		Friends    []Person    `json:"friend"`
		School     School      `json:"school"`
	}

	ti := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	p := Person{
		Name: "Alice",
		NameFacets: nameFacets{
			Origin: "Indonesia",
		},
		Friends: []Person{
			Person{
				Name: "Bob",
				Facets: friendFacet{
					Since:  ti,
					Family: "yes",
					Age:    13,
					Close:  true,
				},
			},
			Person{
				Name: "Charlie",
				Facets: friendFacet{
					Family: "maybe",
					Age:    16,
				},
			},
		},
		School: School{
			Name: "Wellington School",
			Facets: schoolFacet{
				Since: ti,
			},
		},
	}

	err = req.SetObject(&p)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}

	auid := resp.AssignedUids["blank-0"]

	q := fmt.Sprintf(`
    {

        me(func: uid(%v)) {
            name @facets
            friend @facets {
                name
            }
            school @facets {
                name
            }

        }
    }`, auid)

	req.SetQuery(q)
	resp, err = dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}
	type Root struct {
		Me Person `json:"me"`
	}

	var r Root
	err = client.Unmarshal(resp.N, &r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Me: %+v\n", r.Me)
}

func ExampleReq_SetObject_list(t *testing.T) {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	x.Check(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient(
		[]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)

	req := client.Req{}

	// This example shows example for SetObject for predicates with list type.
	type Person struct {
		Uid         uint64   `json:"_uid_"`
		Address     []string `json:"address"`
		PhoneNumber []int64  `json:"phone_number"`
	}

	p := Person{
		Address:     []string{"Redfern", "Riley Street"},
		PhoneNumber: []int64{9876, 123},
	}

	req.SetSchema(`
		address: [string] .
		phone_number: [int] .
	`)

	err = req.SetObject(&p)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}

	uid := resp.AssignedUids["blank-0"]

	q := fmt.Sprintf(`
	{
		me(func: uid(%d)) {
			_uid_
			address
			phone_number
		}
	}
	`, uid)

	req.SetQuery(q)
	resp, err = dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}

	type Root struct {
		Me Person `json:"me"`
	}

	var r Root
	err = client.Unmarshal(resp.N, &r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Me: %+v\n", r.Me)
}

func ExampleReq_DeleteObject_edges() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	x.Check(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient(
		[]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)

	req := client.Req{}

	type School struct {
		Uid  uint64 `json:"_uid_"`
		Name string `json:"name@en,omitempty"`
	}

	type Person struct {
		Uid      uint64   `json:"_uid_,omitempty"`
		Name     string   `json:"name,omitempty"`
		Age      int      `json:"age,omitempty"`
		Married  bool     `json:"married,omitempty"`
		Friends  []Person `json:"friend,omitempty"`
		Location string   `json:"loc,omitempty"`
		School   *School  `json:"school,omitempty"`
	}

	// Lets add some data first.
	loc := `{"type":"Point","coordinates":[1.1,2]}`
	p := Person{
		Uid:      1000,
		Name:     "Alice",
		Age:      26,
		Married:  true,
		Location: loc,
		Friends: []Person{{
			Uid:  1001,
			Name: "Bob",
			Age:  24,
		}, {
			Uid:  1002,
			Name: "Charlie",
			Age:  29,
		}},
		School: &School{
			Uid:  1003,
			Name: "Crown Public School",
		},
	}

	req.SetSchema(`
		age: int .
		married: bool .
	`)

	err = req.SetObject(&p)
	if err != nil {
		log.Fatal(err)
	}

	q := fmt.Sprintf(`{
		me(func: uid(1000)) {
			_uid_
			name
			age
			loc
			married
			friend {
				_uid_
				name
				age
			}
			school {
				_uid_
				name@en
			}
		}

		me2(func: uid(1001)) {
			_uid_
			name
			age
		}

		me3(func: uid(1003)) {
			_uid_
			name@en
		}

		me4(func: uid(1002)) {
			_uid_
			name
			age
		}
	}`)
	resp, err := dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}

	// Now lets delete the edge between Alice and Bob.
	// Also lets delete the location for Alice.
	req = client.Req{}
	p2 := Person{
		Uid:      1000,
		Location: "",
		Friends:  []Person{Person{Uid: 1001}},
	}
	err = req.DeleteObject(&p2)
	if err != nil {
		log.Fatal(err)
	}

	req.SetQuery(q)
	resp, err = dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}

	type Root struct {
		Me  Person `json:"me"`
		Me2 Person `json:"me2"`
		Me3 School `json:"me3"`
		Me4 Person `json:"me4"`
	}

	var r Root
	err = client.Unmarshal(resp.N, &r)
	fmt.Printf("Resp: %+v\n", r)
}

func ExampleReq_DeleteObject_node() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	x.Check(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient(
		[]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)

	req := client.Req{}

	// In this test we check S * * deletion.
	type Person struct {
		Uid     uint64    `json:"_uid_,omitempty"`
		Name    string    `json:"name,omitempty"`
		Age     int       `json:"age,omitempty"`
		Married bool      `json:"married,omitempty"`
		Friends []*Person `json:"friend,omitempty"`
	}

	req = client.Req{}

	p := Person{
		Uid:     1000,
		Name:    "Alice",
		Age:     26,
		Married: true,
		Friends: []*Person{&Person{
			Uid:  1001,
			Name: "Bob",
			Age:  24,
		}, &Person{
			Uid:  1002,
			Name: "Charlie",
			Age:  29,
		}},
	}

	req.SetSchema(`
		age: int .
		married: bool .
	`)

	err = req.SetObject(&p)
	if err != nil {
		log.Fatal(err)
	}

	q := fmt.Sprintf(`{
		me(func: uid(1000)) {
			_uid_
			name
			age
			married
			friend {
				_uid_
				name
				age
			}
		}

		me2(func: uid(1001)) {
			_uid_
			name
			age
		}

		me3(func: uid(1002)) {
			_uid_
			name
			age
		}
	}`)
	req.SetQuery(q)

	resp, err := dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}

	type Root struct {
		Me  Person `json:"me"`
		Me2 Person `json:"me2"`
		Me3 Person `json:"me3"`
	}

	var r Root
	err = client.Unmarshal(resp.N, &r)
	fmt.Printf("Resp after SetObject: %+v\n", r)

	// Now lets try to delete Alice. This won't delete Bob and Charlie but just remove the
	// connection between Alice and them.
	p2 := Person{
		Uid: 1000,
	}

	req = client.Req{}
	err = req.DeleteObject(&p2)
	if err != nil {
		log.Fatal(err)
	}

	req.SetQuery(q)
	resp, err = dgraphClient.Run(context.Background(), &req)
	err = client.Unmarshal(resp.N, &r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Resp after deleting node: %+v\n", r)
}

func ExampleReq_DeleteObject_predicate() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	x.Checkf(err, "While trying to dial gRPC")
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	x.Check(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient(
		[]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)

	req := client.Req{}

	type Person struct {
		Uid     uint64   `json:"_uid_,omitempty"`
		Name    string   `json:"name,omitempty"`
		Age     int      `json:"age,omitempty"`
		Married bool     `json:"married,omitempty"`
		Friends []Person `json:"friend,omitempty"`
	}

	p := Person{
		Uid:     1000,
		Name:    "Alice",
		Age:     26,
		Married: true,
		Friends: []Person{Person{
			Uid:  1001,
			Name: "Bob",
			Age:  24,
		}, Person{
			Uid:  1002,
			Name: "Charlie",
			Age:  29,
		}},
	}

	req.SetSchema(`
		age: int .
		married: bool .
	`)

	err = req.SetObject(&p)
	if err != nil {
		log.Fatal(err)
	}

	q := fmt.Sprintf(`{
		me(func: uid(1000)) {
			_uid_
			name
			age
			married
			friend {
				_uid_
				name
				age
			}
		}
	}`)
	req.SetQuery(q)

	resp, err := dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}

	type Root struct {
		Me Person `json:"me"`
	}
	var r Root
	err = client.Unmarshal(resp.N, &r)
	fmt.Printf("Response after SetObject: %+v\n\n", r)

	// Now lets try to delete friend and married predicate.
	type DeletePred struct {
		Friend  interface{} `json:"friend"`
		Married interface{} `json:"married"`
	}
	dp := DeletePred{}
	// Basically we want predicate as JSON keys with value null.
	// After marshalling this would become { "friend" : null, "married": null }

	req = client.Req{}
	err = req.DeleteObject(&dp)
	if err != nil {
		log.Fatal(err)
	}

	// Also lets run the query again to verify that predicate data was deleted.
	req.SetQuery(q)
	resp, err = dgraphClient.Run(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Unmarshal(resp.N, &r)
	// Alice should have no friends and only two attributes now.
	fmt.Printf("Response after deletion: %+v\n", r)
}
