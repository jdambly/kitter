package main

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	HOST  = "localhost"
	PORT  = "8080"
	PROTO = "tcp"
)

type Response struct {
	ServerTime string `json:"serverTime"`
	ClientTime string `json:"clientTime"`
	Client     string `json:"client"`
	Server     string `json:"server"`
	Latency    int64  `json:"latency"`
}

func handleRequest(conn net.Conn) {

	// incoming request
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Error().Err(err).Msg("could not close connection")
		}
	}(conn)

	if err != nil {
		log.Error().Err(err).Msg("could not read buffer")
		return
	}
	recvd := string(buffer[:n])
	// get the time before we do anything
	rStamp := time.Now()

	cStamp, err := time.Parse(time.RFC3339Nano, recvd)
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}
	// Calculate latency
	latency := rStamp.Sub(cStamp)
	// prepare the response
	resp := Response{
		ServerTime: rStamp.Format(time.RFC3339Nano),
		ClientTime: cStamp.Format(time.RFC3339Nano),
		Client:     conn.RemoteAddr().String(),
		Server:     conn.LocalAddr().String(),
		Latency:    latency.Milliseconds(),
	}

	// write data to the response
	respBytes, err := json.Marshal(resp)
	if err != nil {
		log.Error().Err(err).Msg("could not marshal response")
		return
	}

	_, err = conn.Write(respBytes)
	if err != nil {
		log.Error().Err(err).Msg("could not write buffer")
	}

}

func startServer() error {
	listen, err := net.Listen(PROTO, HOST+":"+PORT)
	if err != nil {
		log.Error().Err(err).Msg("could not list on port: " + PORT)
		return err
	}
	//close listener
	defer func(listen net.Listener) {
		err := listen.Close()
		if err != nil {
			log.Error().Err(err).Msg("")
		}
	}(listen)
	// this is what starts the server
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Error().Err(err).Msg("")
			return err
		}
		go handleRequest(conn)
	}

}

func startClient() error {
	var err error = nil
	for {
		myTime := time.Now().Format(time.RFC3339Nano)
		// sleep for 500 ms
		time.Sleep(500 * time.Millisecond)

		tcpServer, err := net.ResolveTCPAddr(PROTO, HOST+":"+PORT)
		if err != nil {
			log.Error().Err(err).Msg("ResolveTCPAddr failed")
			break
		}

		conn, err := net.DialTCP(PROTO, nil, tcpServer)
		if err != nil {
			log.Error().Err(err).Msg("Dail failed")
			return err
		}

		_, err = conn.Write([]byte(myTime))
		if err != nil {
			log.Error().Err(err).Msg("Write data failed")
			break
		}

		//buffer to get data
		resp := make([]byte, 1024)
		_, err = conn.Read(resp)
		if err != nil {
			log.Error().Err(err).Msg("Read data failed")
			break
		}

		log.Debug().Msg(string(resp))
		err = conn.Close()
		if err != nil {
			log.Error().Err(err).Msg("Could not close connection")
			break
		}

		time.Sleep(10 * time.Second)
	}
	return err
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if os.Getenv("LOGGING") == "DEBUG" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	errServer := make(chan error)       // create a channel for the server
	errClient := make(chan error)       // create a channel for the client
	stopChan := make(chan os.Signal, 1) // create a channel to look for ctrl+c
	signal.Notify(stopChan, os.Interrupt)
	signal.Notify(stopChan, syscall.SIGTERM)

	go func() {
		log.Info().Msg("Starting TCP server")
		errServer <- startServer()
	}()
	go func() {
		log.Info().Msg("Starting tcp client")
		errClient <- startClient()
	}()

	exitCode := 0
	select {
	case errS := <-errServer:
		if errS != nil {
			log.Error().Err(errS).Msg("TCP server error")
			exitCode = 1
		}
	case errC := <-errClient:
		if errC != nil {
			log.Error().Err(errC).Msg("TCP client error")
		}
	case s := <-stopChan:
		log.Info().Msg(fmt.Sprintf("signal %s received", s))
		exitCode = 130
	}

	os.Exit(exitCode)
}
