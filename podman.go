package main

import (
	"bufio"
	"os"
	"strconv"
)

func isPodmanService(pid int) bool {
	file, err := os.Open("/proc/" + strconv.Itoa(pid) + "/cmdline")
	if err != nil {
		return false
	}
	defer file.Close()
	onNul := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		for i := 0; i < len(data); i++ {
			if data[i] == '\x00' {
				return i + 1, data[:i], nil
			}
		}
		return 0, data, bufio.ErrFinalToken
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(onNul)
	scanner.Scan()
	exe := scanner.Text()
	return exe == "/usr/bin/podman"
}
