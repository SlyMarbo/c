package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

var (
	execname      = filepath.Join(os.TempDir(), "test")
	filename      = execname + ".c"
	matcher       = `(%s:\d+:\d+: warning:|%s: In function 'main':)`
	matchWarnings = regexp.MustCompile(fmt.Sprintf(matcher, filename, filename))
	decl          = []byte("int main(int argc, char* argv[]) {\n\t")
	end           = []byte(";\n}\n")
)

func main() {
	defer func() {
		if v := recover(); v != nil {
			fmt.Fprintln(os.Stderr, v)
		}
	}()

	f, err := os.Create(filename)
	handle(err)
	defer f.Close()
	defer os.Remove(filename)
	defer os.Remove(execname)

	write(f, decl)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 || (len(line) == 1 && line[0] == 'q') {
			return
		}

		write(f, line, end)
		handle(f.Sync())

		cmd := exec.Command("gcc", "-o", execname, filename)
		cmd.Stderr = StderrNoWarnings{}
		cmd.Stdout = os.Stdout

		handle(cmd.Run())

		cmd = exec.Command(execname)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		err = cmd.Run()
		if _, ok := err.(*exec.ExitError); !ok {
			handle(err)
		}

		handle(f.Truncate(0))
		fmt.Print("> ")
	}

	handle(scanner.Err())
}

type StderrNoWarnings struct{}

func (StderrNoWarnings) Write(b []byte) (int, error) {
	if !matchWarnings.Match(b) {
		return os.Stderr.Write(b)
	}
	return len(b), nil
}

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func write(w io.Writer, b ...[]byte) {
	for _, b := range b {
		_, err := w.Write(b)
		handle(err)
	}
}
