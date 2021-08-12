package bbycrgo

import (
    "bufio"

    "fmt"
    "io"
    "net"
    "sync"

    shellquote "github.com/kballard/go-shellquote"
    log "github.com/sirupsen/logrus"
)

const (
    SOCKET_TYPE string = "tcp"
    SOCKET_ADDR string = ":20201"
)

func SocketHandler(events chan Event, conn net.Conn) {
    log.WithFields(log.Fields{
        "addr": conn.RemoteAddr().String(),
    }).Info("Client connected to the input socket")
    defer conn.Close()

    for {
        event_msg, err := bufio.NewReader(conn).ReadString('\n')
        if err == io.EOF {
            log.WithFields(log.Fields{
                "addr": conn.RemoteAddr().String(),
            }).Warn("Client closed socket connection, received EOF")
            return
        } else if err != nil {
            log.WithFields(log.Fields{
                "err": err,
            }).Warn("Couldn't read socket data")
            return
        }

        event_data, err := shellquote.Split(event_msg)
        if err != nil {
            log.WithFields(log.Fields{
                "data": event_msg,
                "err":  err,
            }).Error("Received invalid msg")
            fmt.Fprintf(conn, "%s\n", InvalidEventData.Error())
            continue
        }

        var argument_data string
        if len(event_data) >= 3 {
            argument_data = shellquote.Join(event_data[2:]...)
        }
        if len(event_data) < 2 {
            log.WithFields(log.Fields{
                "data": event_msg,
                "err":  err,
            }).Error("Received invalid msg")
            fmt.Fprintf(conn, "%s\n", InvalidEventData.Error())
            continue
        }
        ev := Event{Target: event_data[0],
            Command:   event_data[1],
            Arguments: argument_data,
            Response:  make(chan string)}
        events <- ev
        resp := <-ev.Response
        log.Debug(resp)
        fmt.Fprintf(conn, "%s\n", resp)
    }
}

func SocketServer(events chan Event, progress *sync.WaitGroup) {
    defer progress.Done()
    log.WithFields(log.Fields{
        "type": SOCKET_TYPE,
        "addr": SOCKET_ADDR,
    }).Info("Listening on Socket")
    sock, err := net.Listen(SOCKET_TYPE, SOCKET_ADDR)
    if err != nil {
        log.WithFields(log.Fields{
            "err": err,
        }).Error("Couldn't open socket")
        return
    }
    defer sock.Close()

    for {
        conn, err := sock.Accept()
        if err != nil {
            log.WithFields(log.Fields{
                "err": err,
            }).Warn("Couldn't accept socket connection")
            return
        }

        go SocketHandler(events, conn)
    }
}

type SocketClient struct {
    conn net.Conn
}

func SocketConnect() (SocketClient, error) {
    conn, err := net.Dial(SOCKET_TYPE, SOCKET_ADDR)
    if err != nil {
        return SocketClient{nil}, err
    }

    return SocketClient{conn}, nil
}

func (s *SocketClient) Send(ev Event) (string, error) {
    event_data := []string{ev.Target, ev.Command, ev.Arguments}
    msg := shellquote.Join(event_data...)
    return s.Write(msg)
}

func (s *SocketClient) Write(msg string) (string, error) {
    // TODO input handling, sanitizing etc
    var empty_resp string

    msg = fmt.Sprintf("%s\n", msg)
    _, err := fmt.Fprintf(s.conn, msg)
    if err != nil {
        return empty_resp, err
    }
    return s.Read()
}

func (s *SocketClient) Read() (string, error) {
    return bufio.NewReader(s.conn).ReadString('\n')
}

func (s *SocketClient) Close() error {
    return s.conn.Close()
}
