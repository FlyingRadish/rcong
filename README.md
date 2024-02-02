# RCONG

RCON Protocol implementation for Go

## Why again?

I recently started playing `PalWorld`, and I've searched through all the RCON Go libraries, but none of them offer good support for the `ShowPlayers` command. That's why I created this library.

## Command-line tool
If you just want to send RCON commands easily, [rcong-cli](https://github.com/FlyingRadish/rcong-cli) is available.

## Install
```
go get github.com/FlyingRadish/rcong
```

## Useage
```
package main

import (
	"fmt"
	"log"

	"github.com/FlyingRadish/rcong"
)

func main() {
	conn := rcong.NewRCONConnection("127.0.0.1", 25575, "password", 3, 10)
	conn.Connect()
	defer conn.Close()

	response, err := conn.ExecCommand("ShowPlayers")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response)	
}

```
