/*
 *
 * Copyright 2019 tenorbear@github
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

//go:generate protoc -I ../proto --go_out=plugins=grpc:../proto ../proto/injector.proto

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/go-redis/redis"
	pb "github.com/tenorbear/redis-injector/proto"
	"google.golang.org/grpc"
)

var port = flag.Int("port", 50051, "Which port to listen to.")
var redisServer = flag.String("redisServer", "localhost:6379", "Address of the Redis server.")
var redisPass = flag.String("redisPass", "", "Password to the Redis server")
var redisDB = flag.Int("redisDB", 0, "Which Redis DB to connect to")

// server is used to implement injector.KeyInjectionService.
type server struct {
	redisClient *redis.Client
}

func (s *server) Inject(ctx context.Context, in *pb.InjectRequest) (*pb.InjectReply, error) {
	if err := s.redisClient.Set(in.Key, in.Value, 0).Err(); err != nil {
		log.Printf("Error injecting key-val: %v", err)
		return &pb.InjectReply{}, err
	}
	log.Printf("Keyval injected.")

	return &pb.InjectReply{}, nil
}

func main() {
	flag.Parse()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     *redisServer,
		Password: *redisPass,
		DB:       *redisDB,
	})
	defer redisClient.Close()
	if _, err := redisClient.Ping().Result(); err != nil {
		log.Fatalf("failed to get response from Redis server: %v.", err)
	}

	// Start the server.
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterKeyInjectorServer(s, &server{redisClient: redisClient})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
