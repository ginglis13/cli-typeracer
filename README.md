# cli-typeracer

cli for playing typeracer, and a server to host games on.

![](https://yld.me/raw/dM1J.png)

### Goal of the Project

The goal of this project was to be my first \~real\~ project in Go. I have messed around with Go on and off
for the past year, and made a small go program that was a copy of the tree linux utility. However, this is
the first "real" practical application I've made in Go. It makes use of Websockets, Goroutines, mutexes, etc.
I've learned a lot about websockets and goroutines from doing this project, and my prior knowledge of using
mutexes helped during development. I ended up using this project for  [Hacker in the Bazaar Project 2.](https://www3.nd.edu/~pbui/teaching/cse.40842.sp21/project02.html). Another goal of this project was to simply make a clone of typeracer. When I had first thought of this
idea, I had been brainstorming applications that I could turn into CLI programs, and typeracer just so happened to be one of them.
Go works well for this project due to the simplicity with which you can develop concurrent programs.

### Unique Go Features in this Project

- Goroutines
  * used on the Client side for sending and receiving JSON messages from the Server
  * used on the Server side for sending and receiving JSON messages from the Client
  * used to place keyboard input in the background
### Installing Sauce

First, install dependencies for the client:
```
	go get -u github.com/eiannone/keyboard
	go get -u github.com/fatih/color
	go get -u github.com/ginglis13/cli-typeracer/models
```

Run the client with `go run client.go` and the server with `go run server.go` to view options
