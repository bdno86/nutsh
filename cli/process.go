package cli

import (
	"bufio"
	"github.com/kr/pty"
	"os"
	"os/exec"
)

func init() {
	exec.Command("stty", "-F", "/dev/tty", "-echo", "-icanon", "min", "1").Run()
}

func Quit() {
	exec.Command("stty", "-F", "/dev/tty", "echo").Run()
	os.Exit(0)
}

func startProcess(command string, stdin <-chan rune, stdout chan<- rune) {
	tty, _ := pty.Start(stringToCmd(command))

	input, _ := os.Create("/tmp/nutsh-input")
	output, _ := os.Create("/tmp/nutsh-output")

	go func() {
		for {
			r := <-stdin
			input.Write([]byte(string(r)))
			tty.Write([]byte(string(r)))
		}
	}()

	go func() {
		reader := bufio.NewReader(tty)
		for {
			r, _, err := reader.ReadRune()
			if err != nil {
				Quit()
			}
			output.Write([]byte(string(r)))
			stdout <- r
		}
	}()
}
