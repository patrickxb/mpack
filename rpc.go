package mpack

import (
        "bytes"
        "container/vector"
        "log"
        "net"
        "os"
        "sync"
        "time"
)

type GenericMap map[string]interface{}
type GenericList []interface{}

// the handler type for rpc calls
type handler func(arg interface{}) (interface{}, os.Error)

var handlerMap map[string]handler

func init() {
        handlerMap = make(map[string]handler)
}

// map a name to a handler function
func Handle(name string, function handler) os.Error {
        handlerMap[name] = function
        return nil
}

func ListenAndServe(host string) {
        tcpAddr, err := net.ResolveTCPAddr("tcp", host)
        if err != nil {
                log.Printf("error resolving address %s: %s", host, err)
                return
        }

        listener, err := net.Listen(tcpAddr.Network(), tcpAddr.String())
        if err != nil {
                log.Printf("error listening: %s", err)
                return
        }

        defer listener.Close()
        log.Printf("Starting server run loop, listening on %s.", host)

        for {
                log.Printf("waiting for connections")
                conn, err := listener.Accept()
                if err != nil {
                        log.Fatalf("accept error: %v", err)
                }

                go serve(conn)
                log.Printf("connection established: %s", conn)
        }
}

func serve(conn net.Conn) {
        results := make(chan []byte, 1024)
        quit := make(chan bool)
        go sendResults(results, quit, conn)
        for {
                rpc, _, err := Unpack(conn)
                if err != nil {
                        quit <- true
                        return
                }
                go processRPC(rpc, results)
        }
}

func sendResults(results chan []byte, quit chan bool, conn net.Conn) {
        for {
                select {
                case result := <-results:
                        n, err := conn.Write(result)
                        if err != nil {
                                log.Printf("error writing result: %s", err)
                        }
                        if n != len(result) {
                                log.Printf("didn't fully write result.  wrote %d bytes, not %d bytes", n, len(result))
                        }
                case <-quit:
                        return
                }
        }
}

func processRPC(rpc interface{}, results chan []byte) {
        defer func() {
                if err := recover(); err != nil {
                        log.Println("processRPC failed", err)
                        response, e := errorResponse(0, err.(os.Error).String())
                        if e == nil {
                                results <- response
                        }
                }
        }()
        startTime := time.Nanoseconds()
        args := NewArray(rpc)
        if args.Item(0) != rpc_request {
                log.Printf("did not receive an rpc request")
                return
        }

        msgid := args.Uint32Item(1)
        procedure := args.StringItem(2)
        procedureArgs := args.Item(3)

        log.Printf("rpc request: msgid=%d, proc=%s, args=%s", msgid, procedure, procedureArgs)

        fn, present := handlerMap[procedure]
        if !present {
                log.Printf("error:  no procedure '%s'", procedure)
                response, err := errorResponse(msgid, "no procedure: "+procedure)
                if err != nil {
                        log.Printf("error making err response:", err)
                        return
                }
                results <- response
                return
        }

        result, err := fn(procedureArgs)
        if err != nil {
                log.Printf("error calling procedure '%s': %s", procedure, err)
                response, err := errorResponse(msgid, err.String())
                if err != nil {
                        log.Printf("error making err response:", err)
                        return
                }
                results <- response
                return
        }

        response, err := successResponse(msgid, result)
        if err != nil {
                log.Printf("error making success response:", err)
                return
        }
        results <- response

        log.Printf("rpc execute time: %.3f ms", (float64)(time.Nanoseconds()-startTime)/1e6)
}

func errorResponse(msgid uint32, message string) ([]byte, os.Error) {
        response := makeResponse(msgid)
        response[2] = message
        return packMessage(response)
}

func successResponse(msgid uint32, result interface{}) ([]byte, os.Error) {
        response := makeResponse(msgid)
        response[3] = result
        return packMessage(response)
}

func makeResponse(msgid uint32) []interface{} {
        response := make([]interface{}, 4)
        response[0] = rpc_response
        response[1] = msgid
        response[2] = nil
        response[3] = nil
        return response
}

func packMessage(message interface{}) ([]byte, os.Error) {
        b := new(bytes.Buffer)
        _, err := Pack(b, message)
        if err != nil {
                return nil, err
        }
        return b.Bytes(), nil
}

type RPCClient struct {
        Host           string
        conn           net.Conn
        idCounter      int64
        outputChannels map[int64]chan interface{}
        Connected      bool
}

func NewRPCClient(host string) (*RPCClient, os.Error) {
        result := new(RPCClient)
        result.Host = host

        result.outputChannels = make(map[int64]chan interface{})

        tcpAddr, err := net.ResolveTCPAddr("tcp", host)
        if err != nil {
                return nil, err
        }
        result.conn, err = net.DialTCP("tcp", nil, tcpAddr)
        if err != nil {
                return nil, err
        }
        result.Connected = true
        go result.StartReader()
        return result, nil
}

func (client *RPCClient) StartReader() {
        for {
                generic, _, err := Unpack(client.conn)
                if err != nil {
                        if err == os.EOF {
                                log.Printf("%s: eof", client.Host)
                                log.Printf("any data? %v", generic)
                                client.Connected = false
                                return
                        }

                        log.Printf("%s: unpack error: %s", client.Host, err)
                        continue
                }
                response := NewArray(generic)
                if response.Item(0) != rpc_response {
                        log.Printf("didn't get rpc_response")
                        continue
                }
                if response.Item(2) != nil {
                        log.Printf("error: %s", response.Item(2))
                        continue
                }
                output, present := client.outputChannels[response.IntItem(1)]
                if !present {
                        log.Printf("no output channel found for msgid %d", response.IntItem(1))
                        continue
                }

                /*
                   log.Printf("buffer: %v", response.BufferItem(3))
                   result, _, err := Unpack(response.BufferItem(3))
                   if err != nil {
                           log.Printf("result unpack error: %s", err)
                           continue
                   }

                   log.Printf("unpacked result: %v", result)
                */

                output <- response.Item(3)
                client.outputChannels[response.IntItem(1)] = nil, false
        }
}

// XXX let them call this with multiple params and wrap them in an array
func (client *RPCClient) CallSync(procedure string, params interface{}) (interface{}, os.Error) {
        ch := make(chan interface{})
        err := client.Call(procedure, params, ch)
        if err != nil {
                return nil, err
        }
        result := <-ch
        return result, nil
}

// XXX let them call this with multiple params and wrap them in an array
func (client *RPCClient) Call(procedure string, params interface{}, output chan interface{}) os.Error {
        msgid := client.idCounter
        client.idCounter += 1
        args := make([]interface{}, 1)
        args[0] = params
        request := make([]interface{}, 4)
        request[0] = rpc_request
        request[1] = msgid
        request[2] = procedure
        request[3] = args
        msg, err := packMessage(request)
        if err != nil {
                log.Printf("Error packing message: %s", err)
                return err
        }
        client.outputChannels[msgid] = output
        client.conn.Write(msg)

        return nil
}

func (client *RPCClient) IsConnected() bool {
        return client.Connected
}

func (client *RPCClient) Close() {
        client.Connected = false
        client.conn.Close()
}

type ClientPool struct {
        Host    string
        MaxSize int
        clients *vector.Vector
        lock    *sync.Mutex
}

func NewClientPool(host string) *ClientPool {
        result := new(ClientPool)
        result.Host = host
        result.clients = new(vector.Vector)
        result.lock = new(sync.Mutex)
        result.MaxSize = 10
        return result
}

func (cp *ClientPool) Get() (*RPCClient, os.Error) {
        cp.lock.Lock()
        defer cp.lock.Unlock()

        for cp.clients.Len() > 0 {
                result := cp.clients.Pop().(*RPCClient)
                if result.Connected {
                        return result, nil
                }
                log.Printf("discarding disconnected client")
                // XXX could check for stale connections like Pool does...
        }

        result, err := NewRPCClient(cp.Host)
        if err != nil {
                return nil, err
        }
        return result, nil
}

func (cp *ClientPool) Put(client *RPCClient) {
        cp.lock.Lock()
        defer cp.lock.Unlock()

        if client.Connected == false {
                log.Printf("returning a disconnected client...closing...")
                client.Close()
                log.Printf("close finished")
                return
        }

        if cp.clients.Len() >= cp.MaxSize {
                log.Printf("pool is full, discarding client")
                client.Close()
                return
        }

        log.Printf("putting client back in pool")
        cp.clients.Push(client)
}
