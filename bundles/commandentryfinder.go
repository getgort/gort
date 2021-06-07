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

package bundles

import (
	"context"

	"github.com/getgort/gort/data"
)

// FindCommandEntry is used to find the enabled commands with the provided
// bundle and command names. If either is empty, it is treated as a wildcard.
// Importantly, this must only return ENABLED commands!
type CommandEntryFinder interface {
	FindCommandEntry(ctx context.Context, bundle, command string) ([]data.CommandEntry, error)
}
