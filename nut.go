// Package nut implements the Network UPS Tools network protocol.
//
// This package only implements those functions necessary for the
// nut_exporter; it is therefore not complete.
package nut

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// A Client wraps a connection to a NUT server.
type Client struct {
	conn net.Conn
	br   *bufio.Reader
}

// Dial dials a NUT server using TCP. If the address does not contain
// a port number, it will default to 3493.
func Dial(addr string) (*Client, error) {
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		addr = net.JoinHostPort(addr, "3493")
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), nil
}

// NewClient wraps an existing net.Conn.
func NewClient(conn net.Conn) *Client {
	return &Client{conn, bufio.NewReader(conn)}
}

// Close closes the connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) list(typ string) ([]string, error) {
	cmd := "LIST " + typ
	if err := c.write(cmd); err != nil {
		return nil, err
	}
	l, err := c.read()
	if err != nil {
		return nil, err
	}
	expected := "BEGIN " + cmd
	if l != expected {
		return nil, fmt.Errorf("expected %q, got %q", expected, l)
	}

	var lines []string
	expected = typ + " "
	for {
		l, err := c.read()
		if err != nil {
			return nil, err
		}
		if l == "END "+cmd {
			break
		}
		if !strings.HasPrefix(l, expected) {
			return nil, fmt.Errorf("expected %q, got %q", expected, l)
		}
		l = l[len(expected):]
		lines = append(lines, l)
	}
	return lines, nil
}

// UPSs returns a list of all UPSs on the server.
func (c *Client) UPSs() ([]string, error) {
	lines, err := c.list("UPS")
	if err != nil {
		return nil, err
	}

	var upss []string
	for _, l := range lines {
		idx := strings.IndexByte(l, ' ')
		if idx == -1 {
			return nil, errors.New("protocol error")
		}
		ups := l[:idx]
		upss = append(upss, ups)
	}
	return upss, nil
}

// Variables returns all variables and their values for a UPS.
func (c *Client) Variables(ups string) (map[string]string, error) {
	lines, err := c.list("VAR " + ups)
	if err != nil {
		return nil, err
	}
	vars := map[string]string{}
	for _, l := range lines {
		idx := strings.IndexByte(l, ' ')
		if idx == -1 {
			return nil, errors.New("protocol error")
		}
		k := l[:idx]
		v := l[idx+1:]
		v, err = strconv.Unquote(v)
		if err != nil {
			return nil, err
		}

		vars[k] = v
	}
	return vars, nil
}

func (c *Client) write(s string) error {
	_, err := c.conn.Write([]byte(s + "\n"))
	return err
}

func (c *Client) read() (string, error) {
	l, err := c.br.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(l) > 0 {
		l = l[:len(l)-1]
	}
	return l, nil
}
