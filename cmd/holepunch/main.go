package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"holepunch/internal/auth"
	"holepunch/internal/server"
	"holepunch/internal/upnp"
)

var version = "v1.0"

func main() {
	// Flags
	port := flag.Int("port", 8080, "Port to serve on")
	dir := flag.String("dir", ".", "Directory to serve")
	user := flag.String("user", "", "Username for Basic Auth")
	pass := flag.String("pass", "", "Password for Basic Auth")
	token := flag.String("token", "", "Pre-shared token (generated if empty)")
	noUpnp := flag.Bool("no-upnp", false, "Disable UPnP port forwarding")
	showVersion := flag.Bool("version", false, "Show version")
	quiet := flag.Bool("quiet", false, "Suppress non-error output")
	flag.Parse()

	if *showVersion {
		fmt.Println("HolePunch", version)
		os.Exit(0)
	}

	// Validate directory
	absDir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatalf("Invalid directory: %v", err)
	}
	if info, err := os.Stat(absDir); err != nil || !info.IsDir() {
		log.Fatalf("Not a directory: %s", absDir)
	}

	// Setup auth
	authenticator, credentials := auth.Setup(*user, *pass, *token)
	if !*quiet {
		credentials.Print()
	}

	// Setup UPnP
	var externalIP string
	if !*noUpnp {
		externalIP, err = upnp.ForwardPort(*port)
		if err != nil {
			log.Printf("UPnP warning: %v", err)
			log.Println("Server will still start, but may only be accessible locally.")
			log.Println("Use -no-upnp to suppress this warning.")
		}
		defer upnp.Cleanup(*port)
	}

	// Start server
	srv := server.New(absDir, authenticator, credentials.Mode, *quiet)
	srv.Addr = fmt.Sprintf(":%d", *port)

	if !*quiet {
		server.PrintInfo(absDir, *port, externalIP, credentials)
	}

	log.Fatal(srv.ListenAndServe())
}
