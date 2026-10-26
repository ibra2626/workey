package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/windows"
)

// readConsolePassword reads a password from the Windows console with input echo
// disabled, prompting on stderr. ok is false when stdin is not a console (e.g.
// a pipe or redirect), so the caller falls back to a plain line read. It uses
// golang.org/x/sys/windows, already a dependency, to avoid adding a new module.
func readConsolePassword() (pw string, ok bool, err error) {
	h := windows.Handle(os.Stdin.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return "", false, nil // not a console; caller reads a line instead
	}
	// Drop echo, keep line editing and Ctrl-C handling.
	newMode := mode &^ uint32(windows.ENABLE_ECHO_INPUT)
	newMode |= windows.ENABLE_LINE_INPUT | windows.ENABLE_PROCESSED_INPUT
	if err := windows.SetConsoleMode(h, newMode); err != nil {
		return "", false, nil
	}
	defer func() { _ = windows.SetConsoleMode(h, mode) }()

	fmt.Fprint(os.Stderr, "Parola: ")
	line, rerr := bufio.NewReader(os.Stdin).ReadString('\n')
	fmt.Fprintln(os.Stderr)
	if rerr != nil && line == "" {
		return "", true, rerr
	}
	return strings.TrimRight(line, "\r\n"), true, nil
}
