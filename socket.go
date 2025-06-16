package main

import (
	"context"
	"encoding/json"
	"github.com/getsentry/sentry-go"
	"github.com/logrusorgru/aurora/v4"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	SocketSuccess = `{"success": true}`
	LinkUser      = `INSERT INTO "user" (basket, wii_id) VALUES ('[]', $1) ON CONFLICT(wii_id) DO UPDATE SET basket = '[]', wii_id = $1`
)

func socketListen() {
	// Remove if it didn't gracefully exit for some reason
	os.Remove("/tmp/dominos.sock")

	socket, err := net.Listen("unix", "/tmp/dominos.sock")
	checkError(err)

	defer socket.Close()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove("/tmp/dominos.sock")
		os.Exit(0)
	}()

	pgxConn, err := pool.Acquire(context.Background())
	checkError(err)

	defer pgxConn.Release()
	log.Printf("%s", aurora.Green("UNIX socket connected."))
	log.Printf("%s %s\n", aurora.Green("Listening on UNIX socket:"), socket.Addr())

	// Listen forever
	for {
		conn, err := socket.Accept()
		if err != nil {
			log.Print(aurora.Red("Socket Accept ERROR: "), err.Error(), "\n")
			sentry.CaptureException(err)
		}

		go func(conn net.Conn) {
			defer conn.Close()
			buf := make([]byte, 4096)

			n, err := conn.Read(buf)
			if err != nil {
				log.Print(aurora.Red("Socket Read ERROR: "), err.Error(), "\n")
				sentry.CaptureException(err)
			}

			reply := []byte(SocketSuccess)
			// We only send a Wii Hollywood ID.
			wiiID := strings.Replace(string(buf[:n]), "\n", "", -1)

			_, err = pgxConn.Exec(context.Background(), LinkUser, wiiID)
			if err != nil {
				reply, _ = json.Marshal(map[string]any{
					"success": false,
					"error":   err.Error(),
				})
			}

			_, err = conn.Write(append(reply, []byte("\n")...))
			if err != nil {
				log.Print(aurora.Red("Socket Write ERROR: "), err.Error(), "\n")
				sentry.CaptureException(err)
			}
		}(conn)
	}
}
