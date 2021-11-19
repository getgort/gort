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

package relay

// Most of the follow exit codes were borrowed from sysexits.h.

const (
	// ExitOK represents a successful termination.
	ExitOK = 0

	// ExitGeneral is catchall for otherwise unspecified errors.
	ExitGeneral = 1

	// ExitNoUser represents a "user unknown" error.
	ExitNoUser = 67

	// ExitNoRelay represents "relay name unknown" error.
	ExitNoRelay = 68

	// ExitUnavailable represents a "relay unavailable" error.
	ExitUnavailable = 69

	// ExitInternalError represents an internal software error. It's returned
	// if a command has a failure that's detectable by the framework.
	ExitInternalError = 70

	// ExitSystemErr represents a system (Gort) error (for example, it can't
	// spawn a worker).
	ExitSystemErr = 71

	// ExitTimeout represents a timeout exceeded.
	ExitTimeout = 72

	// ExitIoErr represents an input/output error.
	ExitIoErr = 74

	// ExitTempFail represents a temporary failure. The user can retry.
	ExitTempFail = 75

	// ExitProtocol represents a remote error in protocol.
	ExitProtocol = 76

	// ExitNoPerm represents a permission denied.
	ExitNoPerm = 77

	// ExitCannotInvoke represents that the invoked command cannot execute.
	// TODO(mtitmus) What does this mean, exactly?
	ExitCannotInvoke = 126

	// ExitCommandNotFound represents that the command can't be found.
	ExitCommandNotFound = 127
)
