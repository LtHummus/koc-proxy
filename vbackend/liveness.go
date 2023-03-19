package vbackend

import (
	"context"
	"errors"
	"net"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/spf13/viper"
)

var (
	errTimeout           = errors.New("timeout")
	errUnknownHost       = errors.New("unknown host")
	errConnectionRefused = errors.New("connection refused")
)

func detectError(err error) error {
	if err == nil {
		return nil
	}

	switch t := err.(type) {
	case *net.OpError:
		if t.Op == "dial" {
			return errUnknownHost
		} else if t.Op == "read" {
			return errConnectionRefused
		}
	case syscall.Errno:
		if t == syscall.ECONNREFUSED {
			return errConnectionRefused
		}
	}

	return err
}

// IsAlive checks to see if the given upstream server is accepting connections. We don't actually do anything except
// attempt to open a TCP connection to that server and port.
// (TODO: is there anything better we can do? Maybe some sort of health check endpoint?)
func IsAlive(ctx context.Context) (bool, error) {
	backend := viper.GetString("proxy.backend.upstream")

	conn, err := net.Dial("tcp", backend)
	if err != nil {
		dErr := detectError(err)
		if dErr == errTimeout || dErr == errUnknownHost || dErr == errConnectionRefused {
			log.Warn().Err(dErr).Msg("looks like backend isn't up")
			return false, nil
		}
		return false, err
	}

	// if we got here, we connected, just close the connection and don't worry about it
	err = conn.Close()
	if err != nil {
		return false, err
	}

	return true, nil
}
