package cmd

import (
	"log"
	"net"
	"os"

	"github.com/JakWai01/airdrip/pkg/signaling"
	"github.com/spf13/cobra"
)

const (
	laddrKey = "laddr"
)

var signalCmd = &cobra.Command{
	Use:   "signal",
	Short: "Start a signaling server.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle lifecycle
		fatal := make(chan error)
		done := make(chan struct{})

		go func() {

			l, err := net.Listen("tcp", "localhost:8080")
			if err != nil {
				panic(err)
			}

			for {

				// breaker := make(chan error)

				// go func() {

				// 	// Parse subsystem-specific flags
				// 	addr, err := net.ResolveTCPAddr("tcp", "localhost:8080")
				// 	if err != nil {
				// 		fatal <- fmt.Errorf("could not resolve address: %v", err)

				// 		return
				// 	}

				// 	// Parse PORT env variable for Heroku compatibility
				// 	if portEnv := os.Getenv("PORT"); portEnv != "" {
				// 		port, err := strconv.Atoi(portEnv)
				// 		if err != nil {
				// 			fatal <- fmt.Errorf("could not parse port: %v", port)

				// 			return
				// 		}

				// 		addr.Port = port
				// 	}

				// 	signaler := signaling.NewSignalingServer()

				// 	defer func() {
				// 		signaler.Close() // Best offer
				// 	}()

				// 	// Start
				// 	log.Printf("signaling server listening on %v", addr.String())

				// 	// Register interrupt handler
				// 	go func() {
				// 		s := make(chan os.Signal, 1)
				// 		signal.Notify(s, os.Interrupt)
				// 		<-s

				// 		log.Println("gracefully shutting down signaling server")

				// 		// Register secondary interrupt handler (which hard-exists)
				// 		go func() {
				// 			s := make(chan os.Signal, 1)
				// 			signal.Notify(s, os.Interrupt)
				// 			<-s

				// 			log.Fatal("cancelled graceful signaling server shutdown, existing immediately")
				// 		}()

				// 		breaker <- nil

				// 		_ = signaler.Close() // Best effort

				// 		done <- struct{}{}
				// 	}()

				// This won't work! We can't use http
				// handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				// 	l, err := net.Listen("tcp", "localhost:8080")
				// 	conn, err := l.Accept()
				// 	if err != nil {
				// 		log.Println("could not accept on Socket:", err)

				// 		return
				// 	}

				// 	log.Println("client connected")

				// 	go func() {
				// 		if err != nil {
				// 			panic(err)
				// 		}

				// 		signaler.HandleConn(conn)
				// 	}()
				// })
				// fatal <- http.ListenAndServe(addr.String(), handler)

				// }()

				// err := <-breaker

				// Interrupting
				// if err != nil {
				// 	break
				// }

				// log.Println("signaling server crashed, restarting in 1s:", err)

				// time.Sleep(time.Second)

				c, err := l.Accept()
				if err != nil {
					panic(err)
				}
				signaler := signaling.NewSignalingServer()

				go signaler.HandleConn(c)
			}
		}()

		for {
			select {
			case err := <-fatal:
				log.Fatal(err)
			case <-done:
				os.Exit(0)
			}
		}
	},
}
