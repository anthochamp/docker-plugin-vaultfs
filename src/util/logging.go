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

func Printf(format string, a ...any) (n int, err error) {
	if DebugMode || Verbose {
		return fmt.Printf(format, a...)
	}

	return 0, nil
}

func Tracef(format string, a ...any) (n int, err error) {
	if DebugMode {
		return Printf("Trace: "+format, a...)
	}

	return 0, nil
}

func Noticef(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "Notice: "+format, a...)
}

func Errorf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format, a...)
}

func Fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "Fatal: "+format, a...)
	os.Exit(1)
}
