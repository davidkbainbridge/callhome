package configserver

import (
	"fmt"
	"io"
	"os"
	"log"
	"net/http"
)

type Server struct {
	ListenIP               string
	ListenPort             int
	ListenPath             string
	ConfigurationDirectory string
}

type client struct {
	class       string
	mac_address string
	boot_time   string
}

func (s *Server) register_client(c client) {
	log.Printf("REGISTER: {%s, %s, %s}", c.mac_address, c.class, c.boot_time)
}

func (s *Server) respond(c client, w http.ResponseWriter) {
	s.register_client(c)

	/*
	         * Search for a client configuration file in the following order
	         *      $DIR/config.$CLASS.$MAC
		 *	$DIR/config.$MAC
	         *      $DIR/config.$CLASS
	         *      $DIR/config
	*/
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

func (s *Server) Listen() error {
	log.Printf("Listening on: %s:%d/%s", s.ListenIP, s.ListenPort, s.ListenPath)
	http.HandleFunc("/"+s.ListenPath, func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		c := client{
			class:       r.Form["class"][0],
			mac_address: r.Form["mac"][0],
			boot_time:   r.Form["boottime"][0],
		}
		s.respond(c, w)
	})
	return http.ListenAndServe(fmt.Sprintf("%s:%d", s.ListenIP, s.ListenPort), nil)
}
