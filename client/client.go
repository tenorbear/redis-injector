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

package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/tenorbear/redis-injector/proto"
	"google.golang.org/grpc"
)

var address = flag.String("address", "localhost:50051", "Server to connect to.")
var keyFile = flag.String("keyFile", "", "File containing tab-separated key-val pairs. One pair each line.")

func main() {
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.Dial(*address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()
	c := pb.NewKeyInjectorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Read the key-val file.
	file, err := os.Open(*keyFile)
	if err != nil {
		log.Fatalf("Cannot read from key file %v. Error: %v", *keyFile, err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	// Send key insertion RPCs to server.
	var cnt int
	for scanner.Scan() {
		arr := strings.Split(scanner.Text(), "\t")
		if len(arr) != 2 {
			continue
		}
		key, val := arr[0], arr[1]
		_, err := c.Inject(ctx, &pb.InjectRequest{Key: key, Value: val})
		if err != nil {
			log.Fatalf("Error injecting keyval: %v", err)
		}
		cnt++
	}
	log.Printf("Key injection finished. %d pairs of keyval injected.\n", cnt)
}
