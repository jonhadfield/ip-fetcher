package main_test

import "os"

func testCleanUp(args []string) {
	os.Args = args
	_ = os.Unsetenv("TEST_EXIT")
}
