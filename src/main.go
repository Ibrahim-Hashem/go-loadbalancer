package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Server interface {
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request) 
}
func (s *simpleServer) Address() string { return s.addr}
func (s *simpleServer) IsAlive() bool {return true}
func (s *simpleServer) Serve(rw http.ResponseWriter, r *http.Request){
	s.proxy.ServeHTTP(rw,r)
} 
type simpleServer struct {
	addr	string
	proxy	*httputil.ReverseProxy
}

func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	checkError(err)
	
	return &simpleServer{
		addr: addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

type LoadBalancer struct {
	port 			string
	servers 		[]Server
	roundRobinIndex int
}


func (lb *LoadBalancer) AddServer(server Server) {
	lb.servers = append(lb.servers, server)
}

func (lb *LoadBalancer) getNextAvailableServer() Server {
	server := lb.servers[lb.roundRobinIndex%len(lb.servers)]
	for !server.IsAlive(){
		lb.roundRobinIndex++
		server = lb.servers[lb.roundRobinIndex%len(lb.servers)]
	}
	lb.roundRobinIndex++
	return server
}

func (lb *LoadBalancer) serverProxy(rw http.ResponseWriter, r *http.Request) {
	targetServer := lb.getNextAvailableServer()
	fmt.Printf("forwarding request to address %q\n",targetServer.Address())
	targetServer.Serve(rw,r)

}

func newLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port: port,
		servers: servers,
		roundRobinIndex: 0,
	}
}

func checkError(err error) {
    if err != nil {
        log.Fatalf("An error occurred: %v", err)
    }
}







func main(){
	servers := []Server{
		newSimpleServer("https://www.facebook.com"),
		newSimpleServer("https://www.google.com"),
		newSimpleServer("https://www.youtube.com"),
	}
	
	lb := newLoadBalancer("8080", servers)
	handleRedirect := func (rw http.ResponseWriter, r *http.Request){
		lb.serverProxy(rw,r)
	}
	http.HandleFunc("/", handleRedirect)
	fmt.Printf("serving requests at `localhost:%s`\n",lb.port)
	log.Fatal(http.ListenAndServe(":" + lb.port, nil))
}