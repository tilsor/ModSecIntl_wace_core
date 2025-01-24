/*
Package comm handles all communications between the clients (WAFs) and
WACE.
*/
package comm

import (
	"context"
	"net"

	pb "wace/waceproto"

	lg "github.com/tilsor/ModSecIntl_logging/logging"

	"google.golang.org/grpc"
)

// var port = flag.Int("port", 10000, "The server port")

// The Handlers struct has all the handlers that are needed to
// implement the communication protocol with the WAFs.
type Handlers struct {
	SendRequest            func(string, string, []string) int32
	SendReqLineAndHeaders  func(string, string, string, []string) int32
	SendRequestBody        func(string, string, []string) int32
	SendResponse           func(string, string, []string) int32
	SendRespLineAndHeaders func(string, string, string, []string) int32
	SendResponseBody       func(string, string, []string) int32
	Check                  func(string, string, map[string]string) (bool, error)
	Init				   func(string) int32
	Close				   func(string) int32
}

// type Result struct {
// 	BlockTransaction bool
// 	Message          string
// 	StatusCode       int32
// }

type server struct {
	pb.UnimplementedWaceProtoServer

	handlers Handlers
}

var grpcServer *grpc.Server

func startTransactionLogging(transactionID string) {
	l := lg.Get()
	l.StartTransaction(transactionID)
}

func (s *server) SendRequest(ctx context.Context, in *pb.SendRequestParams) (*pb.SendRequestResult, error) {
	startTransactionLogging(in.GetTransactId())
	res := s.handlers.SendRequest(in.GetTransactId(), in.GetRequest(), in.GetModelId())

	return &pb.SendRequestResult{StatusCode: res}, nil
}

func (s *server) SendReqLineAndHeaders(ctx context.Context, in *pb.SendReqLineAndHeadersParams) (*pb.SendReqLineAndHeadersResult, error) {
	startTransactionLogging(in.GetTransactId())
	res := s.handlers.SendReqLineAndHeaders(in.GetTransactId(), in.GetReqLine(), in.GetReqHeaders(), in.GetModelId())

	return &pb.SendReqLineAndHeadersResult{StatusCode: res}, nil
}

func (s *server) SendRequestBody(ctx context.Context, in *pb.SendRequestBodyParams) (*pb.SendRequestBodyResult, error) {
	startTransactionLogging(in.GetTransactId())
	res := s.handlers.SendRequestBody(in.GetTransactId(), in.GetBody(), in.GetModelId())

	return &pb.SendRequestBodyResult{StatusCode: res}, nil
}

func (s *server) SendResponse(ctx context.Context, in *pb.SendResponseParams) (*pb.SendResponseResult, error) {
	startTransactionLogging(in.GetTransactId())
	res := s.handlers.SendResponse(in.GetTransactId(), in.GetResponse(), in.GetModelId())

	return &pb.SendResponseResult{StatusCode: res}, nil
}

func (s *server) SendRespLineAndHeaders(ctx context.Context, in *pb.SendRespLineAndHeadersParams) (*pb.SendRespLineAndHeadersResult, error) {
	startTransactionLogging(in.GetTransactId())
	res := s.handlers.SendRespLineAndHeaders(in.GetTransactId(), in.GetStatusLine(), in.GetRespHeaders(), in.GetModelId())

	return &pb.SendRespLineAndHeadersResult{StatusCode: res}, nil
}

func (s *server) SendResponseBody(ctx context.Context, in *pb.SendResponseBodyParams) (*pb.SendResponseBodyResult, error) {
	startTransactionLogging(in.GetTransactId())
	res := s.handlers.SendResponseBody(in.GetTransactId(), in.GetBody(), in.GetModelId())

	return &pb.SendResponseBodyResult{StatusCode: res}, nil
}

func (s *server) Check(ctx context.Context, in *pb.CheckParams) (*pb.CheckResult, error) {
	l := lg.Get()
	l.StartTransaction(in.GetTransactId())

	block, err := s.handlers.Check(in.GetTransactId(), in.GetDecisionId(), in.GetWafParams())

	buf := l.EndTransaction(in.GetTransactId())

	if err != nil {
		return &pb.CheckResult{BlockTransaction: 0, Msg: string(buf) + "\nError checking transaction: " + err.Error(), StatusCode: 1}, nil
	}

	var blockTransaction int32
	if block {
		blockTransaction = 1
	} else {
		blockTransaction = 0
	}

	return &pb.CheckResult{BlockTransaction: blockTransaction, Msg: string(buf) + "\nTransaction information analyzed successfully!\n", StatusCode: 0}, nil
}

func (s *server) Init(ctx context.Context, in *pb.InitParams) (*pb.InitResult, error) {
	startTransactionLogging(in.GetTransactId())
	res := s.handlers.Init(in.GetTransactId())

	return &pb.InitResult{StatusCode: res}, nil
}

func (s *server) Close(ctx context.Context, in *pb.CloseParams) (*pb.CloseResult, error) {
	startTransactionLogging(in.GetTransactId())
	res := s.handlers.Close(in.GetTransactId())

	return &pb.CloseResult{StatusCode: res}, nil
}

// Listen implements the main loop of the wace server. It will call
// the appropriate registered handler when a client calls a given
// method.
func Listen(handlers Handlers, address string, port string) error {
	logger := lg.Get()
	lis, err := net.Listen("tcp", address+":"+port)
	if err != nil {
		return err
	}

	grpcServer = grpc.NewServer()
	s := server{handlers: handlers}

	pb.RegisterWaceProtoServer(grpcServer, &s)
	logger.Printf(lg.INFO, "GRPC Server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		return err
	}

	return nil
}
