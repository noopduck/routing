package main

import "fmt"

func main() {
	gwIp, err := FindLinuxDefaultGW()
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(gwIp)
}
