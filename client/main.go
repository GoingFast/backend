package main

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	pb "github.com/GoingFast/backend2/specs"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	svc := newService()
	r := chi.NewRouter()
	r.Use(middleware.Heartbeat("/healthz"))
	r.Get("/message", svc.messageHandler)
	http.ListenAndServe(":8081", r)
}

type service struct {
	conn pb.MessageServiceClient
}

func fallbackEnv(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}

func newService() service {
	var (
		err  error
		conn *grpc.ClientConn
	)
	if os.Getenv("TLS") != "" {
		cert, err := ioutil.ReadFile("/etc/certs/tls.crt")
		if err != nil {
			log.Fatal(err)
		}
		roots := x509.NewCertPool()
		roots.AppendCertsFromPEM(cert)
		credsClient := credentials.NewClientTLSFromCert(roots, "")

		conn, err = grpc.Dial(fallbackEnv("SERVER_ADDR", ":50051"), grpc.WithTransportCredentials(credentials.TransportCredentials(credsClient)))
	} else {
		conn, err = grpc.Dial(fallbackEnv("SERVER_ADDR", ":50051"), grpc.WithInsecure())
	}
	if err != nil {
		log.Fatal(err)
	}
	client := pb.NewMessageServiceClient(conn)
	return service{client}
}

func (s service) messageHandler(w http.ResponseWriter, r *http.Request) {
	res, err := s.conn.Message(r.Context(), &empty.Empty{})
	if err != nil {
		w.Write([]byte(fmt.Sprintf("%s-mm", err.Error())))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	res.ClientHostname = fallbackEnv("HOSTNAME", "clienthostname")
	json.NewEncoder(w).Encode(&res)
}
