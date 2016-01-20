/*
 * Copyright 2016 Ciena Corporation
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * you may obtain a copy of the License at
 *
 *   http://www.apache.org/license/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, sofware
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Provides a simple configuration service for CORD pods. This server will accept request from
// client devices and return to them a pre-defined configuration file that can be specified
// as a file on the disk identified by a device class, MAC address, or both.
package configserver

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Represents the configuration for the "configuration" server. This is used to specify the parameters
// for the server, IP, port on which it listens as well as where it attempts to locate device configuration
// files.
type Server struct {
	ListenIP               string
	ListenPort             int
	ListenPath             string
	ConfigurationDirectory string
}

// Represents a registration dataum from a client.
type client struct {
	class       string
	mac_address string
	boot_time   string
}

// registers a client with the configuraiton system, can be use to optimize when and what
// is given back to the client as an initialization function
func (s *Server) register_client(c client) {
	log.Printf("REGISTER: {%s, %s, %s}", c.mac_address, c.class, c.boot_time)
}

// handles a call home request from the client. The client is registered and then if an initialization
// file can be located it is returned to the client.
func (s *Server) call_home_handler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
        c := client{
		class:       r.Form["class"][0],
		mac_address: r.Form["mac"][0],
		boot_time:   r.Form["boottime"][0],
	}

	s.register_client(c)

	// Search for a client configuration file in the following order
	//      $DIR/config.$CLASS.$MAC
	//	$DIR/config.$MAC
	//      $DIR/config.$CLASS
	//      $DIR/config
	search := []string{
		fmt.Sprintf("%s/config.%s.%s", s.ConfigurationDirectory, c.class, c.mac_address),
		fmt.Sprintf("%s/config.%s", s.ConfigurationDirectory, c.mac_address),
		fmt.Sprintf("%s/config.%s", s.ConfigurationDirectory, c.class),
		fmt.Sprintf("%s/config", s.ConfigurationDirectory)}

	for _, file := range search {
		config, err := os.Open(file)
		if err == nil {
			defer config.Close()
			cnt, err := io.Copy(w, config)
			if err == nil {
				// Copy complete
				log.Printf("copied %d bytes of configuration file '%s' to client", cnt, file)
				break
			}
		}
		log.Printf("unable to find or copy file '%s' to client as configuration", file)
	}
}

// Start the configuration server and have it listen to and respond to HTTP request
func (s *Server) ListenAndServe() error {
	log.Printf("Listening on: %s:%d/%s", s.ListenIP, s.ListenPort, s.ListenPath)
        http.HandleFunc("/"+s.ListenPath, s.call_home_handler)
	return http.ListenAndServe(fmt.Sprintf("%s:%d", s.ListenIP, s.ListenPort), nil)
}
