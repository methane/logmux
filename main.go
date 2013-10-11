package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func logServer(con net.Conn, c chan []byte) {
	defer func() {
		r := recover()
		if r != nil {
			log.Println(r)
		}
	}()
	defer con.Close()

	reader := bufio.NewReader(con)
	for {
		data, err := reader.ReadBytes('\n')
		if len(data) > 0 {
			select {
			case c <- data:
			default:
				log.Println("Buffer full", data)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
			break
		}
	}
}

func logWriter1(cmd string, ch chan []byte) error {
	proc := exec.Command("sh", "-c", cmd)
	pipe, err := proc.StdinPipe()
	if err != nil {
		log.Println("Can't get pipe", err)
		return err
	}
	defer pipe.Close()

	err = proc.Start()
	if err != nil {
		log.Println("Can't start process", err)
		return err
	}

	stop := make(chan bool)
	go func() {
		for {
			select {
			case <-stop:
				return
			case buf := <-ch:
				pipe.Write(buf)
			}
		}
	}()

	err = proc.Wait()
	stop <- true
	log.Println("Child stopped", err)
	return err
}

func logWriter(cmd string, ch chan []byte) {
	for {
		logWriter1(cmd, ch)
		time.Sleep(time.Second)
	}
}

func parseSock(sock string) (string, string) {
	sockType := "tcp"
	switch {
	case sock[0] == '/':
		sockType = "unix"
	case strings.Contains(sock, "://"):
		pos := strings.Index(sock, "://")
		sockType = sock[:pos]
		sock = sock[pos+3:]
	}
	return sockType, sock
}

func main() {
	sock := os.Args[1]
	cmd := os.Args[2]
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)

	sockType, sock := parseSock(sock)
	log.Println("Listening", sockType, sock)
	l, err := net.Listen(sockType, sock)
	if err != nil {
		log.Panic(err)
	}
	defer l.Close()

	ch := make(chan []byte, 1024)
	go logWriter(cmd, ch)

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			go logServer(c, ch)
		}
	}()

	sig := <-sigc
	log.Println(sig)
}
