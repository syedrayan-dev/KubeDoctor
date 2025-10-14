/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
)

// FatalError handles fatal errors in a testable way
type FatalError struct {
	Message string
	Err     error
}

func (f FatalError) Error() string {
	if f.Err != nil {
		return fmt.Sprintf("%s: %v", f.Message, f.Err)
	}
	return f.Message
}

// ErrorHandler defines an interface for handling fatal errors
type ErrorHandler interface {
	HandleFatalError(logger logr.Logger, err error)
}

// DefaultErrorHandler is the default implementation that calls os.Exit(1)
type DefaultErrorHandler struct{}

func (h DefaultErrorHandler) HandleFatalError(logger logr.Logger, err error) {
	logger.Error(err, "Fatal error occurred")
	os.Exit(1)
}

// TestErrorHandler is for testing - it panics instead of calling os.Exit
type TestErrorHandler struct{}

func (h TestErrorHandler) HandleFatalError(logger logr.Logger, err error) {
	logger.Error(err, "Fatal error occurred")
	panic(err) // Panic allows tests to catch the error
}

// Global error handler - can be replaced for testing
var GlobalErrorHandler ErrorHandler = DefaultErrorHandler{}

// HandleFatalError is a convenience function that uses the global error handler
func HandleFatalError(logger logr.Logger, err error, message string) {
	fatalErr := FatalError{
		Message: message,
		Err:     err,
	}
	GlobalErrorHandler.HandleFatalError(logger, fatalErr)
}

// SetErrorHandlerForTesting sets the error handler to the test version
func SetErrorHandlerForTesting() {
	GlobalErrorHandler = TestErrorHandler{}
}

// ResetErrorHandler resets the error handler to the default version
func ResetErrorHandler() {
	GlobalErrorHandler = DefaultErrorHandler{}
}
