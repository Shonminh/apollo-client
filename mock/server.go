package mock

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sync"
)

type Server struct {
	quit chan int
	wg   sync.WaitGroup
	host string
}

func NewServer() *Server {
	return &Server{
		quit: make(chan int),
	}
}

func (s *Server) SetHost(host string) {
	s.host = host
}

func (s *Server) GetHost() string {
	if len(s.host) == 0 {
		return "127.0.0.1:8000"
	}

	return s.host
}

func (s *Server) Run() {
	s.wg.Add(1)
	defer s.wg.Done()

	router := gin.Default()
	router.GET("notifications/v2", s.handleNotifications)
	router.GET("configs/:appId/:cluster/:namespaceName", s.handleConfigs)

	var addr string
	if len(s.host) == 0 {
		addr = "127.0.0.1:8000"
	} else {
		addr = s.host
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	<-s.quit
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("Server Shutdown:%v", err)
	}
	log.Println("Mock Server exist")

	return

}

func (s *Server) Stop() {
	close(s.quit)
}

func (s *Server) Wait() {
	s.wg.Wait()
}
