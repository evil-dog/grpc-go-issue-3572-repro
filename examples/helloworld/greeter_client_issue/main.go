/*
 *
 * Copyright 2015 gRPC authors.
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

// Package main implements a client for Greeter service.
package main

import (
  "context"
  "path"
  "os"
  "time"

  "github.com/gofrs/uuid"
  grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/logging"
  grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
  "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
  "github.com/pkg/errors"
  "github.com/sirupsen/logrus"
  log "github.com/sirupsen/logrus"
  "google.golang.org/grpc"
  "google.golang.org/grpc/codes"
  "google.golang.org/grpc/grpclog"
  "google.golang.org/grpc/metadata"
  "google.golang.org/grpc/balancer/roundrobin"
  "google.golang.org/grpc/resolver"
  pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	address     = "greeter-server:50051"
	defaultName = "world"
)

// Copied from chirpstack-application-server/internal/logging/logging.go (more or less the whole file)
type ContextKey string

const ContextIDKey ContextKey = "ctx_id"

type contextIDGetter interface {
  GetContextId() []byte
}

func UnaryServerCtxIDInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
  // generate unique id
  ctxID, err := uuid.NewV4()
  if err != nil {
    return nil, errors.Wrap(err, "new uuid error")
  }

  // set id to context and add as logrus field
  ctx = context.WithValue(ctx, ContextIDKey, ctxID)
  ctxlogrus.AddFields(ctx, log.Fields{
    "ctx_id": ctxID,
  })

  // set id as response header
  header := metadata.Pairs("ctx-id", ctxID.String())
  grpc.SendHeader(ctx, header)

  // execute the handler
  return handler(ctx, req)
}

// UnaryClientCtxIDInterceptor logs the context id from a RPC response.
func UnaryClientCtxIDInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
  // read reasponse meta-data (set by remote server)
  var header metadata.MD
  opts = append(opts, grpc.Header(&header))

  // set start time and invoke api methd
  startTime := time.Now()
  err := invoker(ctx, method, req, reply, cc, opts...)

  // get error code
  code := grpc_logging.DefaultErrorToCode(err)

  // get log-level for code
  level := grpc_logrus.DefaultCodeToLevel(code)

  // get log fields
  logFields := clientLoggerFields(ctx, method, reply, err, code, startTime, header)

  // log api call
  levelLogf(log.WithFields(logFields), level, "finished client unary call")

  return err
}

func clientLoggerFields(ctx context.Context, fullMethodString string, resp interface{}, err error, code codes.Code, start time.Time, header metadata.MD) logrus.Fields {
  service := path.Dir(fullMethodString)[1:]
  method := path.Base(fullMethodString)

  fields := logrus.Fields{
    "system":        "grpc",
    "span.kind":     "client",
    "grpc.service":  service,
    "grpc.method":   method,
    "grpc.duration": time.Since(start),
    "grpc.code":     code.String(),
    "ctx_id":        ctx.Value(ContextIDKey),
  }

  if err != nil {
    fields[logrus.ErrorKey] = err
  }

  // read context id from meta-data
  if values := header.Get("ctx-id"); len(values) != 0 {
    ctxID, err := uuid.FromString(values[0])
    if err == nil {
      fields["grpc.ctx_id"] = ctxID
    }
  }

  return fields
}

func levelLogf(entry *logrus.Entry, level logrus.Level, format string, args ...interface{}) {
  switch level {
  case logrus.DebugLevel:
    entry.Debugf(format, args...)
  case logrus.InfoLevel:
    entry.Infof(format, args...)
  case logrus.WarnLevel:
    entry.Warningf(format, args...)
  case logrus.ErrorLevel:
    entry.Errorf(format, args...)
  case logrus.FatalLevel:
    entry.Fatalf(format, args...)
  case logrus.PanicLevel:
    entry.Panicf(format, args...)
  }
}

// End - Copied logging.go

// Copied from chirpstack-application-server/cmd/chirpstack-application-server/main.go #15-50
type grpcLogger struct {
  *log.Logger
}

func (gl *grpcLogger) V(l int) bool {
  level, ok := map[log.Level]int{
    log.DebugLevel: 0,
    log.InfoLevel:  1,
    log.WarnLevel:  2,
    log.ErrorLevel: 3,
    log.FatalLevel: 4,
  }[log.GetLevel()]
  if !ok {
    return false
  }

  return l >= level
}

func (gl *grpcLogger) Info(args ...interface{}) {
  if log.GetLevel() == log.DebugLevel {
    log.Debug(args...)
  }
}

func (gl *grpcLogger) Infoln(args ...interface{}) {
  if log.GetLevel() == log.DebugLevel {
    log.Debug(args...)
  }
}

func (gl *grpcLogger) Infof(format string, args ...interface{}) {
  if log.GetLevel() == log.DebugLevel {
    log.Debugf(format, args...)
  }
}
// End - Copied main.go #15-50

// Copied from chirpstack-application-server/cmd/chirpstack-application-server/main.go #52-58
func init() {
  grpclog.SetLoggerV2(&grpcLogger{log.StandardLogger()})
  resolver.SetDefaultScheme("dns")
}
// End - Copied main.go #52-58

func main() {
  // Adapted from chirpstack-application-server/cmd/chirpstack-application-server/cmd/root_run.go #82
  log.SetLevel(log.Level(5))

  // Copied from chirpstack-application-server/internal/backend/networkserver/networkserver.go #99-112
  logrusEntry := log.NewEntry(log.StandardLogger())
  logrusOpts := []grpc_logrus.Option{
    grpc_logrus.WithLevels(grpc_logrus.DefaultCodeToLevel),
  }

  nsOpts := []grpc.DialOption{
    grpc.WithBlock(),
    grpc.WithUnaryInterceptor(
      UnaryClientCtxIDInterceptor,
    ),
    grpc.WithStreamInterceptor(
      grpc_logrus.StreamClientInterceptor(logrusEntry, logrusOpts...),
    ),
    grpc.WithBalancerName(roundrobin.Name),
  // End - Copied networkserver.go #99-112
  // Adapted from chirpstack-application-server/internal/backend/networkserver/networkserver.go #116
    grpc.WithInsecure(),
  }

  // Copied from chirpstack-application-server/internal/backend/networkserver/networkserver.go #136-138
  ctx, cancel := context.WithTimeout(context.Background(), 20000*time.Millisecond)
  defer cancel()
  // End - Copied networkserver.go #136-138

	// Set up a connection to the server.
  // Adapted from chirpstack-application-server/internal/backend/networkserver/networkserver.go #139
	conn, err := grpc.DialContext(ctx, address, nsOpts...)

  //////**** Everything below here is from original example client code ****//////

//	conn, err := grpc.Dial(address, nsOpts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	nctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(nctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}
