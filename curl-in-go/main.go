package main

import (
	"fmt"
	"os/exec"
)

func main() {
	// url := os.Args[1] // URL
	url := "https://www.reddit.com/r/recipes/.rss"
	// cmd := exec.Command("curl", "-O", url)
	cmd := exec.Command("curl", "-A", "Mozilla/5.0 (X11; Linux x86_64; rv:60.0) Gecko/20100101 Firefox/81.0", "-O", url)
	// cmd := exec.Command("ls", "-al")
	// cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}
