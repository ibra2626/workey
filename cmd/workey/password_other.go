//go:build !windows

package main

// readConsolePassword has no echo-suppressing implementation on non-Windows
// platforms in v1 (the runtime target is local Windows; see DECISIONS). It
// reports ok=false so the caller reads the password as a plain stdin line. A
// Linux deploy (WKY-12) can add a termios-based implementation then.
func readConsolePassword() (pw string, ok bool, err error) {
	return "", false, nil
}
