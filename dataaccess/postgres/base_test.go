/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package postgres

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/getgort/gort/data"
	gerr "github.com/getgort/gort/errors"
	"github.com/stretchr/testify/assert"
)

var (
	configs = data.DatabaseConfigs{
		Host:       "localhost",
		Password:   "password",
		Port:       5432,
		SSLEnabled: false,
		User:       "gort",
	}

	da PostgresDataAccess
)

func TestMain(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute/2)
	defer cancel()

	cleanup, err := startDatabaseContainer(ctx, t)
	defer cleanup()
	assert.NoError(t, err, "failed to start database container")

	t.Run("testInitialize", testInitialize)

	t.Run("testUserAccess", testUserAccess)
	t.Run("testGroupAccess", testGroupAccess)
	t.Run("testTokenAccess", testTokenAccess)
	t.Run("testBundleAccess", testBundleAccess)
}

func expectErr(t *testing.T, err error, expected error) {
	if err == nil {
		t.Fatalf("Expected error %q but didn't get one", expected)
	} else if !gerr.Is(err, expected) {
		t.Fatalf("Wrong error:\nExpected: %s\nGot: %s\n", expected, err)
	}
}

func startDatabaseContainer(ctx context.Context, t *testing.T) (func(), error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return func() {}, err
	}

	reader, err := cli.ImagePull(ctx, "docker.io/library/postgres:13", types.ImagePullOptions{})
	if err != nil {
		return func() {}, err
	}
	io.Copy(os.Stdout, reader)

	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:        "postgres:13",
			ExposedPorts: nat.PortSet{"5432/tcp": {}},
			Env: []string{
				"POSTGRES_USER=gort",
				"POSTGRES_PASSWORD=password",
			},
			Cmd: []string{"postgres"},
		},
		&container.HostConfig{
			PortBindings: map[nat.Port][]nat.PortBinding{"5432/tcp": {nat.PortBinding{HostPort: "5432/tcp"}}},
		},
		nil, nil, "gort-test")
	if err != nil {
		return func() {}, err
	}

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		id := resp.ID

		if err := cli.ContainerStop(ctx, id, nil); err != nil {
			t.Log("warning: failed to stop test container: ", err)
		}

		if err := cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{}); err != nil {
			t.Log("warning: failed to remove test container: ", err)
		} else {
			t.Log("container", id[:12], "cleaned up successfully")
		}
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return cleanup, err
	}

	return cleanup, nil
}

func testInitialize(t *testing.T) {
	da = NewPostgresDataAccess(configs)

	t.Log("Waiting for database to be ready")

	for {
		db, err := da.connect("postgres")

		if db != nil && err == nil {
			t.Log("database is ready!")
			break
		}

		t.Log("Sleeping 1 second...")
		time.Sleep(time.Second)
	}

	err := da.Initialize()
	assert.NoError(t, err)

	t.Run("testDatabaseExists", testDatabaseExists)
	t.Run("testTablesExist", testTablesExist)
}

func testDatabaseExists(t *testing.T) {
	da = NewPostgresDataAccess(configs)

	err := da.Initialize()
	assert.NoError(t, err)

	// Test database "gort" exists
	conn, err := da.connect("gort")
	assert.NoError(t, err)
	defer conn.Close()

	assert.NotNil(t, conn)

	if conn != nil {
		assert.NoError(t, conn.Ping())
	}

	// Meta-test: non-existant database should return nil connection
	nconn, err := da.connect("doesntexist")
	assert.Error(t, err)
	assert.Nil(t, nconn)
}

func testTablesExist(t *testing.T) {
	expectedTables := []string{"users", "groups", "groupusers", "tokens", "bundles"}

	db, err := da.connect("gort")
	assert.NoError(t, err)
	defer db.Close()

	// Expects these tables
	for _, table := range expectedTables {
		b, err := da.tableExists(table, db)
		assert.NoError(t, err)
		assert.True(t, b)
	}

	// Expect not to find this one.
	b, err := da.tableExists("doestexist", db)
	assert.NoError(t, err)
	assert.False(t, b)
}
