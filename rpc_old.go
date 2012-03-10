package mpack

/* a lot of this is borrowed from golang standard rpc package */

import (
	"bytes"
	"errors"
	"log"
	"net"
	"reflect"
	"strings"
	"time"
)

type methodType struct {
	method reflect.Method
}

type service struct {
	name    string
	rcvr    reflect.Value
	kind    reflect.Type
	methods map[string]*methodType
}

type Server struct {
	Host       string
	serviceMap map[string]*service
}

func NewServer(host string) *Server {
	result := new(Server)
	result.Host = host
	return result
}

func (server *Server) Register(rcvr interface{}) error {
	if server.serviceMap == nil {
		server.serviceMap = make(map[string]*service)
	}
	s := new(service)
	s.kind = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()

	// XXX there's some stuff in rpc/server.go that I'm skipping

	if _, present := server.serviceMap[sname]; present {
		return errors.New("rpc service is already defined: " + sname)
	}
	s.name = sname
	s.methods = make(map[string]*methodType)

	for m := 0; m < s.kind.NumMethod(); m++ {
		method := s.kind.Method(m)
		// mtype := method.Type
		mname := method.Name

		s.methods[mname] = &methodType{method: method}
		log.Println("registered method", mname)
	}

	if len(s.methods) == 0 {
		return errors.New("register type " + sname + " has no exported methods")
	}

	server.serviceMap[s.name] = s
	log.Printf("service map: %s", server.serviceMap)
	return nil
}

func (server *Server) HasMethod(serviceName, methodName string) bool {
	if _, present := server.serviceMap[serviceName]; present == false {
		return false
	}
	srvc, _ := server.serviceMap[serviceName]
	if _, present := srvc.methods[methodName]; present == false {
		return false
	}
	return true
}

func (server *Server) CallMethod(serviceName, methodName string, args interface{}) (interface{}, error) {
	srvc, present := server.serviceMap[serviceName]
	if present == false {
		return nil, errors.New("no service found: " + serviceName)
	}
	method, present := srvc.methods[methodName]
	if present == false {
		return nil, errors.New("no method in " + serviceName + " found named " + methodName)
	}
	//	log.Printf("method: %s", method)
	function := method.method.Func
	results := function.Call([]reflect.Value{srvc.rcvr, reflect.ValueOf(args)})
	//	log.Printf("result[0]: %s", results[0].Interface())
	//	log.Printf("result[1]: %s", results[1].Interface())
	err := results[1].Interface()
	if err != nil {
		return nil, err.(error)
	}
	return results[0].Interface(), nil
}

func (server *Server) replySuccess(conn net.Conn, msgid uint32, result interface{}) error {
	response := make([]interface{}, 4)
	response[0] = rpc_response
	response[1] = msgid
	response[2] = nil
	response[3] = result
	b := new(bytes.Buffer)
	_, err := Pack(b, response)
	if err != nil {
		return err
	}
	_, err = conn.Write(b.Bytes())
	return err
}

func (server *Server) replyError(conn net.Conn, msgid uint32, errmsg string) error {
	response := make([]interface{}, 4)
	response[0] = rpc_response
	response[1] = msgid
	response[2] = errmsg
	response[3] = nil
	b := new(bytes.Buffer)
	_, err := Pack(b, response)
	if err != nil {
		return err
	}
	_, err = conn.Write(b.Bytes())
	return err
}

func (server *Server) handleRPC(conn net.Conn) {
	for {
		startTime := time.Now()
		rpc, _, err := Unpack(conn)
		if err != nil {
			log.Printf("read error: %s", err)
			return
		}

		// log.Printf("rpc (%d bytes): %s", bytesRead, rpc)
		args := rpc.([]interface{})
		if args[0] == rpc_request {
			mv := reflect.ValueOf(args[1])
			msgid := uint32(mv.Uint())
			procedure := string(args[2].([]uint8))
			procedureArgs := args[3]
			//			log.Printf("rpc request: msgid=%d, proc=%s, args=%s", msgid, procedure, procedureArgs)
			procElts := strings.Split(procedure, "/")
			if len(procElts) != 2 {
				log.Printf("invalid procedure: %s\n", procedure)
				server.replyError(conn, msgid, "invalid procedure: "+procedure)
				continue
			}

			if server.HasMethod(procElts[0], procElts[1]) == false {
				log.Printf("err:  no procedure: %s\n", procedure)
				server.replyError(conn, msgid, "no procedure: "+procedure)
				continue
			}
			result, err := server.CallMethod(procElts[0], procElts[1], procedureArgs)
			if err != nil {
				log.Printf("error calling method: %s\n", err)
				server.replyError(conn, msgid, "error calling method: "+err.Error())
				continue
			}

			err = server.replySuccess(conn, msgid, result)
			if err != nil {
				log.Printf("error replying: %s\n", err)
			}
			endTime := time.Now()
			msecs := (float64)(endTime.Sub(startTime)) / 1000000
			log.Printf("%s request time: %.3fms", procedure, msecs)
		}
	}
}

func (server *Server) Run() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", server.Host)
	if err != nil {
		log.Printf("error resolving address %s: %s", server.Host, err)
		return
	}

	listener, err := net.Listen(tcpAddr.Network(), tcpAddr.String())
	if err != nil {
		log.Printf("error listening: %s", err)
		return
	}

	defer listener.Close()
	log.Printf("Starting server run loop, listening on %s.", server.Host)

	for {
		log.Printf("waiting for connections")
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("client error: %s", err)
			continue
		}

		log.Printf("connection established: %s", conn)
		go server.handleRPC(conn)
	}

}

type Client struct {
	Host       string
	Connection *net.TCPConn
	Connected  bool
	msgid      int64
}

func NewClient(host string) *Client {
	result := new(Client)
	result.Host = host
	result.Connected = false
	result.msgid = 0
	return result
}

func (c *Client) Connect() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", c.Host)
	if err != nil {
		return err
	}
	c.Connection, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}
	c.Connected = true
	return nil
}

func (c *Client) Close() error {
	err := c.Close()
	if err != nil {
		return err
	}
	c.Connected = false
	return nil
}

// pass in a channel, return the data on a channel?
func (c *Client) Call(method string, args interface{}) (interface{}, error) {
	params := make([]interface{}, 4)
	params[0] = rpc_request
	params[1] = c.msgid
	params[2] = method
	params[3] = args
	if c.Connected == false {
		c.Connect()
	}

	b := new(bytes.Buffer)
	_, err := Pack(b, params)
	if err != nil {
		return nil, err
	}
	_, err = c.Connection.Write(b.Bytes())
	if err != nil {
		return nil, err
	}

	rpc, _, err := Unpack(c.Connection)
	if err != nil {
		log.Printf("read error: %s", err)
		return nil, err
	}
	c.msgid += 1
	return rpc, err
}
