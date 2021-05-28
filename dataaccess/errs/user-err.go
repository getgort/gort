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

package errs

import (
	"errors"
)

// ErrNoSuchUser indicates that the user doesn't exist, or doesn't exist in
// a group.
var ErrNoSuchUser = errors.New("no such user")

// ErrEmptyUserName indicates...
var ErrEmptyUserName = errors.New("user name is empty")

// ErrUserExists TBD
var ErrUserExists = errors.New("user already exists")
