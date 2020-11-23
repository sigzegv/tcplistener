package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// global server config
var config = Config {
	Port: 5555,
	Proxies: []Proxy { // edit this with your servers
		{"127.0.0.1:9222"},
		{"127.0.0.1:9223"},
	},
}

func main() {
	HandleCmdLine()
	TcpListen()
}

// HandleCmdLine manages commandline
func HandleCmdLine() {
	showHelp := false
	flag.BoolVar(&showHelp, "help", false, "Show this help")
	flag.BoolVar(&showHelp, "h", false, "Shortcut for -help")
	flag.IntVar(&config.Port, "port", config.Port, "Set proxy port")
	flag.Parse()

	if showHelp == true {
		fmt.Println("Reverse Proxy Test")
		flag.PrintDefaults()
		os.Exit(0)
	}
}

// TcpListen runs tcp server
func TcpListen() {
    rand.Seed(time.Now().UTC().UnixNano())

	l, err := net.Listen("tcp", ":" + strconv.Itoa(config.Port))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("TCP Proxy started on port", config.Port)

	defer l.Close()

	for {
		cnx, err := l.Accept()

		if err != nil {
			log.Println(err)
		}

		defer cnx.Close()

		go HandleRequest(cnx)
	}
}

// HandleRequest waits for TCP request, forwards it to proxy and prints response
func HandleRequest(srvConn net.Conn) {
	defer srvConn.Close()

	buf, err := ReadPacket(srvConn)

	if err != nil {
		sendInternalError(srvConn)
		return
	}

	// parse http header
	requestPayload := NewHttpPayload()
	requestPayload.ParseBuffer(buf)

	// quick check security
	apikey, ok := requestPayload.Headers["Api-Key"]

	if ok == false || apikey != "MYKEY" {
		sendForbidden(srvConn)
		return
	}
	
	// Proxy forwarding
	p := config.SelectProxy()
	response, err := p.Forward(buf)

	if err != nil {
		log.Println(err)
		sendInternalError(srvConn)
		return
	}

	srvConn.Write(response)
}

// ReadPacket reads a tcp stream from a connection
func ReadPacket(cnx net.Conn) ([]byte, error) {
	buf := []byte{}

	readSize := 256
	tmp := make([]byte, readSize)

	for {
		n, err := cnx.Read(tmp)
		if err != nil {
			return buf, err
		}

		buf = append(buf, tmp[:n]...)

		// End of transmission
		if n < readSize {
			break;
		}
	}

	return buf, nil
}

// ******************
// Manages app config
type Config struct {
	Port int
	Proxies []Proxy
}

// random load balancer
func (cfg *Config) SelectProxy() *Proxy {
	selection := rand.Intn(len(cfg.Proxies))
	return &cfg.Proxies[selection]
}

// ****************************
// Some HTTP Payload management
type HttpPayload struct {
	Method string
	Path string
	Protocol string
	Headers map[string] string
	Body string
}

func NewHttpPayload() HttpPayload {
	return HttpPayload{"", "", "", map[string] string{}, ""}
}

// ParseBuffer parses a http stream to a HttpPayload
func (payload *HttpPayload) ParseBuffer(buf []byte) {
	for i, header := range bytes.Split(buf, []byte("\r\n")) {

		// Method & protocol
		if i == 0 {
			entry := bytes.Split(header, []byte(" "))
			payload.Method = strings.Trim(string(entry[0]), " ")
			payload.Path = strings.Trim(string(entry[1]), " ")
			payload.Protocol = strings.Trim(string(entry[2]), " ")
		}

		// Headers
		if i > 0 && len(header) > 0 {
			entry := bytes.Split(header, []byte(":"))
			payload.Headers[strings.Trim(string(entry[0]), " ")] = strings.Trim(string(entry[1]), " ")
		}
	}
}

// *********************
// Some Proxy management
type Proxy struct {
	Uri string
}

// Forward sends a request to the proxy and returns the response as a raw buffer
func (p *Proxy) Forward(buf []byte) ([]byte, error) {
	// connect to proxy
	pAddr, err := net.ResolveTCPAddr("tcp", p.Uri)
	if err != nil {
		return nil, err
	}

	pCnx, err := net.DialTCP("tcp4", nil, pAddr)
	if err != nil {
		return nil, err
	}

	defer pCnx.Close()

	// forward request
	if _, err := pCnx.Write(buf); err != nil {
		return nil, err
	}

	// grab and return raw response
	data, err := ReadPacket(pCnx)

	return data, err
}

// ************
// Some Helpers
func sendInternalError(cnx net.Conn) {
	msg := "<html><body>Internal Server Error</body></html>"
	fmt.Fprint(cnx, "HTTP/1.1 500 Internal Server Error\r\n")
	fmt.Fprintf(cnx, "Content-Length: %d\r\n", len(msg))
	fmt.Fprint(cnx, "Content-Type: text/html\r\n")
	fmt.Fprint(cnx, "\r\n")
	fmt.Fprint(cnx, msg)
}

func sendForbidden(cnx net.Conn) {
	msg := "<html><body>Forbidden</body></html"
	fmt.Fprint(cnx, "HTTP/1.1 403 Forbidden\r\n")
	fmt.Fprintf(cnx, "Content-Length: %d\r\n", len(msg))
	fmt.Fprint(cnx, "Content-Type: text/html\r\n")
	fmt.Fprint(cnx, "\r\n")
	fmt.Fprint(cnx, msg)
}
