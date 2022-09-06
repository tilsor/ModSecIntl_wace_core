package comm

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	pb "wace/waceproto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func contains(array []string, str string) bool {
	for _, v := range array {
		if str == v {
			return true
		}
	}
	return false
}

func TestCommHandlers(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	sendRequestParams := pb.SendRequestParams{
		TransactId: "1",
		Request:    "Whole Request",
		ModelId:    []string{"trivial", "trivial2"},
	}
	sendReqLineAndHeadersParams := pb.SendReqLineAndHeadersParams{
		TransactId: "2",
		ReqLine:    "Request Line",
		ReqHeaders: "Request Headers",
		ModelId:    []string{"trivial", "trivial2"},
	}
	sendRequestBodyParams := pb.SendRequestBodyParams{
		TransactId: "3",
		Body:       "Request Body",
		ModelId:    []string{"trivial", "trivial2"},
	}
	sendResponseParams := pb.SendResponseParams{
		TransactId: "4",
		Response:   "Whole Response",
		ModelId:    []string{"trivial", "trivial2"},
	}
	sendRespLineAndHeadersParams := pb.SendRespLineAndHeadersParams{
		TransactId:  "5",
		StatusLine:  "Status Line",
		RespHeaders: "Response Headers",
		ModelId:     []string{"trivial", "trivial2"},
	}
	sendResponseBodyParams := pb.SendResponseBodyParams{
		TransactId: "6",
		Body:       "Response Body",
		ModelId:    []string{"trivial", "trivial2"},
	}

	handlers := Handlers{
		SendRequest: func(transactionID, request string, models []string) int32 {
			log.Println("SendRequest")
			if transactionID != sendRequestParams.TransactId ||
				request != sendRequestParams.Request {
				return 1
			}
			if len(models) != 2 || !contains(models, "trivial") || !contains(models, "trivial2") {
				return 1
			}
			return 0
		},
		SendReqLineAndHeaders: func(transactionID, reqLine, reqHeaders string, models []string) int32 {
			log.Println("SendReqLineAndHeaders")
			if transactionID != sendReqLineAndHeadersParams.TransactId ||
				reqLine != sendReqLineAndHeadersParams.ReqLine ||
				reqHeaders != sendReqLineAndHeadersParams.ReqHeaders {
				return 1
			}
			if len(models) != 2 || !contains(models, "trivial") || !contains(models, "trivial2") {
				return 1
			}
			return 0
		},
		SendRequestBody: func(transactionID, body string, models []string) int32 {
			log.Println("SendRequestBody")
			if transactionID != sendRequestBodyParams.TransactId ||
				body != sendRequestBodyParams.Body {
				return 1
			}
			if len(models) != 2 || !contains(models, "trivial") || !contains(models, "trivial2") {
				return 1
			}
			return 0
		},
		SendResponse: func(transactionID, request string, models []string) int32 {
			log.Println("SendResponse")
			if transactionID != sendResponseParams.TransactId ||
				request != sendResponseParams.Response {
				return 1
			}
			if len(models) != 2 || !contains(models, "trivial") || !contains(models, "trivial2") {
				return 1
			}
			return 0
		},
		SendRespLineAndHeaders: func(transactionID, statusLine, respHeaders string, models []string) int32 {
			log.Println("SendRespLineAndHeaders")
			if transactionID != sendRespLineAndHeadersParams.TransactId ||
				statusLine != sendRespLineAndHeadersParams.StatusLine ||
				respHeaders != sendRespLineAndHeadersParams.RespHeaders {
				return 1
			}
			if len(models) != 2 || !contains(models, "trivial") || !contains(models, "trivial2") {
				return 1
			}
			return 0
		},
		SendResponseBody: func(transactionID, body string, models []string) int32 {
			log.Println("SendResponseBody")
			if transactionID != sendResponseBodyParams.TransactId ||
				body != sendResponseBodyParams.Body {
				return 1
			}
			if len(models) != 2 || !contains(models, "trivial") || !contains(models, "trivial2") {
				return 1
			}
			return 0
		},
	}

	go func() {
		err := Listen(handlers, "", "50051")
		if err != nil {
			t.Error(err.Error())
		}
	}()
	defer func() {
		if grpcServer != nil {
			grpcServer.Stop()
		}
	}()

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Printf("did not connect: %v", err)
		return
	}
	c := pb.NewWaceProtoClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// SendRequest
	buf.Reset()
	rSendRequest, err := c.SendRequest(ctx, &sendRequestParams)
	if err != nil {
		t.Error(err.Error())
	}
	if !strings.Contains(buf.String(), "SendRequest") {
		t.Errorf("SendRequest handler not executed")
	}
	if rSendRequest.StatusCode != 0 {
		t.Errorf("SendRequest has wrong parameters")
	}
	// SendReqLineAndHeaders
	buf.Reset()
	rSendReqLineAndHeaders, err := c.SendReqLineAndHeaders(ctx, &sendReqLineAndHeadersParams)
	if err != nil {
		t.Error(err.Error())
	}
	if !strings.Contains(buf.String(), "SendReqLineAndHeaders") {
		t.Errorf("SendReqLineAndHeaders handler not executed")
	}
	if rSendReqLineAndHeaders.StatusCode != 0 {
		t.Errorf("SendReqLineAndHeaders has wrong parameters")
	}
	// SendRequestBody
	buf.Reset()
	rSendRequestBody, err := c.SendRequestBody(ctx, &sendRequestBodyParams)
	if err != nil {
		t.Error(err.Error())
	}
	if !strings.Contains(buf.String(), "SendRequestBody") {
		t.Errorf("SendRequestBody handler not executed")
	}
	if rSendRequestBody.StatusCode != 0 {
		t.Errorf("SendRequestBody has wrong parameters")
	}
	// SendResponse
	buf.Reset()
	rSendResponse, err := c.SendResponse(ctx, &sendResponseParams)
	if err != nil {
		t.Error(err.Error())
	}
	if !strings.Contains(buf.String(), "SendResponse") {
		t.Errorf("SendResponse handler not executed")
	}
	if rSendResponse.StatusCode != 0 {
		t.Errorf("SendResponse has wrong parameters")
	}
	// SendRespLineAndHeaders
	buf.Reset()
	rSendRespLineAndHeaders, err := c.SendRespLineAndHeaders(ctx, &sendRespLineAndHeadersParams)
	if err != nil {
		t.Error(err.Error())
	}
	if !strings.Contains(buf.String(), "SendRespLineAndHeaders") {
		t.Errorf("SendRespLineAndHeaders handler not executed")
	}
	if rSendRespLineAndHeaders.StatusCode != 0 {
		t.Errorf("SendRespLineAndHeaders has wrong parameters")
	}
	// SendResponseBody
	buf.Reset()
	rSendResponseBody, err := c.SendResponseBody(ctx, &sendResponseBodyParams)
	if err != nil {
		t.Error(err.Error())
	}
	if !strings.Contains(buf.String(), "SendResponseBody") {
		t.Errorf("SendResponseBody handler not executed")
	}
	if rSendResponseBody.StatusCode != 0 {
		t.Errorf("SendResponseBody has wrong parameters")
	}

}

func TestCheck(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	handlers := Handlers{
		Check: func(transactionID, decisionPlugin string, wafParams map[string]string) (bool, error) {
			log.Println("Check")
			if transactionID != "1" ||
				decisionPlugin != "simple" {
				return false, errors.New("Invalid parameters")
			}
			if len(wafParams) != 2 || wafParams["anomalyscore"] != "50" || wafParams["inboundthreshold"] != "100" {
				return false, errors.New("Invalid WAF parameters")
			}

			return false, nil
		},
	}

	go func() {
		err := Listen(handlers, "", "50051")
		if err != nil {
			t.Error(err.Error())
		}
	}()
	defer func() {
		if grpcServer != nil {
			grpcServer.Stop()
		}
	}()

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Printf("did not connect: %v", err)
		return
	}
	c := pb.NewWaceProtoClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	buf.Reset()
	rCheck, err := c.Check(ctx, &pb.CheckParams{
		TransactId: "1",
		DecisionId: "simple",
		WafParams:  map[string]string{"anomalyscore": "50", "inboundthreshold": "100"},
	})
	if err != nil {
		t.Error(err.Error())
	}
	if !strings.Contains(buf.String(), "Check") {
		t.Errorf("Check handler not executed")
	}
	if rCheck.StatusCode != 0 {
		t.Errorf("Check has wrong parameters")
	}
	if rCheck.BlockTransaction != 0 {
		t.Errorf("Check incorrectly blocks transaction")
	}

}

func TestCheckBlock(t *testing.T) {
	handlers := Handlers{
		Check: func(transactionID, decisionPlugin string, wafParams map[string]string) (bool, error) {
			return true, nil
		},
	}

	go func() {
		err := Listen(handlers, "", "50051")
		if err != nil {
			t.Error(err.Error())
		}
	}()
	defer func() {
		if grpcServer != nil {
			grpcServer.Stop()
		}
	}()

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Printf("did not connect: %v", err)
		return
	}
	c := pb.NewWaceProtoClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	rCheck, err := c.Check(ctx, &pb.CheckParams{
		TransactId: "1",
		DecisionId: "simple",
		WafParams:  map[string]string{},
	})
	if err != nil {
		t.Error(err.Error())
	}
	if rCheck.StatusCode != 0 {
		t.Errorf("Non-zero status code")
	}
	if rCheck.BlockTransaction != 1 {
		t.Errorf("Check incorrectly permits transaction")
	}
}

func TestCheckError(t *testing.T) {
	handlers := Handlers{
		Check: func(transactionID, decisionPlugin string, wafParams map[string]string) (bool, error) {
			return true, errors.New("check error")
		},
	}

	go func() {
		err := Listen(handlers, "", "50051")
		if err != nil {
			t.Error(err.Error())
		}
	}()
	defer func() {
		if grpcServer != nil {
			grpcServer.Stop()
		}
	}()

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Printf("did not connect: %v", err)
		return
	}
	c := pb.NewWaceProtoClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	rCheck, err := c.Check(ctx, &pb.CheckParams{
		TransactId: "1",
		DecisionId: "simple",
		WafParams:  map[string]string{},
	})
	if err != nil { // This should not happen, even if handler rises an error
		t.Errorf("Error: %v", err)
	}
	if rCheck.StatusCode == 0 {
		t.Errorf("Incorrect status code")
	}
	if rCheck.BlockTransaction != 0 {
		t.Errorf("Check blocks transaction on error")
	}
	if !strings.Contains(rCheck.Msg, "check error") {
		t.Errorf("Incorrect check error message")
	}
}

func TestListenInvalidPort(t *testing.T) {
	err := Listen(Handlers{}, "", "invalid port")
	if err == nil {
		t.Error("Listen did not rise an error")
	}
}
