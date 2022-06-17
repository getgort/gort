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
	"fmt"
	"io"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/getgort/gort/bundles"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess/tests"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	configs = data.DatabaseConfigs{
		Host:       "localhost",
		Password:   "password",
		Port:       10864,
		SSLEnabled: false,
		User:       "gort",
	}

	ctx    context.Context
	cancel context.CancelFunc
	da     PostgresDataAccess
)

// If true, the test database container won't be automatically shut down and
// removed. This is handy for testing.
var DoNotCleanUpDatabase = false

func TestPostgresDataAccessMain(t *testing.T) {
	ctx = context.Background()

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cleanup, err := startDatabaseContainer(ctx, t)
	defer func() {
		if DoNotCleanUpDatabase {
			return
		}
		cleanup()
	}()
	require.NoError(t, err, "failed to start database container")

	ctx, cancel = context.WithTimeout(ctx, time.Minute)
	defer cancel()

	t.Run("testInitialize", testInitialize)

	t.Run("testConnectionLeaks", testConnectionLeaks)

	dat := tests.NewDataAccessTester(ctx, cancel, da)
	t.Run("RunAllTests", dat.RunAllTests)
}

func startDatabaseContainer(ctx context.Context, t *testing.T) (func(), error) {
	ctx2, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return func() {}, err
	}

	reader, err := cli.ImagePull(ctx2, "docker.io/library/postgres:14", types.ImagePullOptions{})
	if err != nil {
		return func() {}, err
	}
	io.Copy(os.Stdout, reader)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	containerName := fmt.Sprintf("gort-test-%x", r.Int())

	resp, err := cli.ContainerCreate(
		ctx2,
		&container.Config{
			Image:        "postgres:14",
			ExposedPorts: nat.PortSet{"5432/tcp": {}},
			Env: []string{
				"POSTGRES_USER=gort",
				"POSTGRES_PASSWORD=password",
			},
			Cmd: []string{"postgres"},
		},
		&container.HostConfig{
			PortBindings: map[nat.Port][]nat.PortBinding{"5432/tcp": {nat.PortBinding{HostPort: "10864/tcp"}}},
		},
		nil, nil, containerName)
	if err != nil {
		return func() {}, err
	}

	cleanup := func() {
		ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
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
	const timeout = 30 * time.Second
	timeoutAt := time.Now().Add(timeout)
	da = NewPostgresDataAccess(configs)

	t.Log("Waiting for database to be ready")

loop:
	for {
		if time.Now().After(timeoutAt) {
			t.Error("timeout waiting for database:", timeout)
			t.FailNow()
		}

		db, err := da.open(ctx, "postgres")
		switch {
		case err != nil:
			t.Logf("connecting to database: %v", err)
		case db == nil:
			t.Log("connecting to database: got nil error but nil db")
		default:
			t.Log("connecting to database: database is ready!")
			break loop
		}

		t.Log("Sleeping 1 second...")
		time.Sleep(time.Second)
	}

	err := da.Initialize(ctx)
	require.NoError(t, err)

	t.Run("testDatabaseExists", testDatabaseExists)
	t.Run("testTablesExist", testTablesExist)
	t.Run("testColumnExists", testColumnExists)
}

func testConnectionLeaks(t *testing.T) {
	const count = 100

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	for i := 0; i < count; i++ {
		bundle, err := getTestBundle()
		require.NoError(t, err)

		err = da.BundleCreate(ctx, bundle)
		require.NoError(t, err)

		err = da.BundleDelete(ctx, bundle.Name, bundle.Version)
		require.NoError(t, err)
	}

	for name, db := range da.dbs {
		require.LessOrEqual(t, db.Stats().OpenConnections, 5, name, " has too many open connections")
		require.LessOrEqual(t, db.Stats().InUse, 2, name, " has too many in-use connections")
	}
}

func testDatabaseExists(t *testing.T) {
	da = NewPostgresDataAccess(configs)

	err := da.Initialize(ctx)
	assert.NoError(t, err)

	// Test database "gort" exists
	db, err := da.open(ctx, DatabaseGort)
	require.NoError(t, err)
	require.NotNil(t, db)

	assert.NoError(t, db.PingContext(ctx))

	// Meta-test: non-existent database should return nil database
	nconn, err := da.open(ctx, "doesntexist")
	assert.Error(t, err)
	assert.Nil(t, nconn)
}

func testTablesExist(t *testing.T) {
	expectedTables := []string{"users", "groups", "groupusers", "tokens", "bundles"}

	conn, err := da.connect(ctx)
	assert.NoError(t, err)
	defer conn.Close()

	// Expects these tables
	for _, table := range expectedTables {
		b, err := da.tableExists(ctx, table, conn)
		assert.NoError(t, err)
		assert.True(t, b)
	}

	// Expect not to find this one.
	b, err := da.tableExists(ctx, "doestexist", conn)
	assert.NoError(t, err)
	assert.False(t, b)
}

func testColumnExists(t *testing.T) {
	tests := []struct {
		Table    string
		Column   string
		Expected bool
	}{
		{
			Table:    "",
			Column:   "",
			Expected: false,
		}, {
			Table:    "users",
			Column:   "",
			Expected: false,
		}, {
			Table:    "users",
			Column:   "foo",
			Expected: false,
		}, {
			Table:    "users",
			Column:   "username",
			Expected: true,
		},
	}

	conn, err := da.connect(ctx)
	assert.NoError(t, err)
	defer conn.Close()

	// Expects these tables
	for i, test := range tests {
		_, err := da.tableExists(ctx, test.Table, conn)
		require.NoError(t, err, "(%d) tableExists error", i)

		c, err := da.columnExists(ctx, test.Table, test.Column, conn)
		require.NoError(t, err, "(%d) columnExists error", i)
		assert.Equal(t, test.Expected, c)
	}

	// Expect not to find this one.
	b, err := da.tableExists(ctx, "doestexist", conn)
	assert.NoError(t, err)
	assert.False(t, b)
}

func getTestBundle() (data.Bundle, error) {
	return bundles.LoadBundleFromFile("../../testing/test-bundle.yml")
}
