package huejack

import (
	"log"
	"net"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func ListenAndServe() error {
	//Alexa is unhappy if we don't have a fixed port TODO: make it configurable?
	laddr := net.TCPAddr{Port: 43312}
	l, err := net.ListenTCP("tcp4", &laddr)
	if err != nil {
		return err
	}

	router := httprouter.New()
	router.GET(upnpPath, upnpSetup)

	router.GET("/api/:userId", getLightsList)
	router.PUT("/api/:userId/lights/:lightId/state", setLightState)
	router.GET("/api/:userId/lights/:lightId", getLightInfo)

	go upnpResponder(l.Addr().(*net.TCPAddr).Port)
	return http.Serve(l, requestLogger(router))
}

// Handler:
// 	state is the state of the "light" after the handler function
//  if error is set to true echo will reply with "sorry the device is not responding"
type Handler func(key, val int)

func requestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("[WEB]", r.RemoteAddr, r.Method, r.URL)
		//		log.Printf("\t%+v\n", r)
		h.ServeHTTP(w, r)
	})
}
