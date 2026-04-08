// SPDX-FileCopyrightText: © 2024 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package util

import (
	"fmt"
	"os"
)

var (
	DebugMode bool = false
	Verbose   bool = false
)

// Printf prints formatted output if DebugMode or Verbose is enabled.
func Printf(format string, a ...any) (n int, err error) {
	if DebugMode || Verbose {
		return fmt.Printf(format, a...)
	}

	return 0, nil
}

// Tracef prints trace output if DebugMode is enabled.
func Tracef(format string, a ...any) (n int, err error) {
	if DebugMode {
		return Printf("Trace: "+format, a...)
	}

	return 0, nil
}

// Noticef prints a notice message to stderr.
func Noticef(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "Notice: "+format, a...)
}

// Errorf prints an error message to stderr.
func Errorf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format, a...)
}

// Fatalf prints a fatal error message to stderr and exits the program.
func Fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "Fatal: "+format, a...)
	os.Exit(1)
}
