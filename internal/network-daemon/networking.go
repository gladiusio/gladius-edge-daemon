package networkd

import (
	"errors"
	"fmt"
	"gladius-edge-daemon/internal/rpc-manager"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path/filepath"
	"runtime"

	"github.com/powerman/rpc-codec/jsonrpc2"
	"github.com/valyala/fasthttp"
)

// Run - Start a web server
func Run() {
	fmt.Println("Starting...")
	// Create some strucs so we can pass info between goroutines
	rpcOut := &rpcmanager.RPCOut{HTTPState: make(chan bool)}
	httpOut := &rpcmanager.HTTPOut{}

	//  -- Content server stuff below --

	// Listen on 8080
	lnContent, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	// Create a content server
	server := fasthttp.Server{Handler: requestHandler(httpOut)}
	// Serve the content
	defer lnContent.Close()
	go server.Serve(lnContent)

	// -- RPC Stuff below --

	// Register RPC methods
	rpc.Register(&rpcmanager.GladiusEdge{RPCOut: rpcOut})
	// Setup HTTP handling for RPC on port 5000
	http.Handle("/rpc", jsonrpc2.HTTPHandler(nil))
	lnHTTP, err := net.Listen("tcp", ":5000")
	if err != nil {
		panic(err)
	}
	defer lnHTTP.Close()
	go http.Serve(lnHTTP, nil)

	fmt.Println("Started RPC server and HTTP server.")

	// Forever check through the channels on the main thread
	for {
		select {
		case state := <-(*rpcOut).HTTPState: // If it can be assigned to a variable
			if state {
				lnContent, err = net.Listen("tcp", ":8080")
				if err != nil {
					panic(err)
				}
				go server.Serve(lnContent)
				fmt.Println("Started HTTP server (from RPC command)")
			} else {
				lnContent.Close()
				fmt.Println("Stopped HTTP server (from RPC command)")
			}
		}
	}
}

func getContentDir() (string, error) {
	// TODO: Actually get correct filepath
	switch runtime.GOOS {
	case "windows":
		return "/var/lib/gladius/gladius-networkd", nil
	case "linux":
		return "/var/lib/gladius/gladius-networkd", nil
	case "darwin":
		return "/var/lib/gladius/gladius-networkd", nil
	default:
		return "", errors.New("Could not detect operating system")
	}
}

// Return a map of the json bundles on disk
func loadContentFromDisk() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	fmt.Println(exPath)
}

// Return a function like the one fasthttp is expecting
func requestHandler(httpOut *rpcmanager.HTTPOut) func(ctx *fasthttp.RequestCtx) {
	// The actual serving function
	return func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/content":
			contentHandler(ctx)
			// TODO: Write stuff to pass back to httpOut
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}
}

func contentHandler(ctx *fasthttp.RequestCtx) {
	// URL format like /content?website=REQUESTED_SITE
	website := string(ctx.QueryArgs().Peek("website"))

	// // TODO: Make this serve the appropriate JSON
	fmt.Fprintf(ctx, "Hi there! You asked for %q", website)
}