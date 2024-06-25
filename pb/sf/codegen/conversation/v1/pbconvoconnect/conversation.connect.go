// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: sf/codegen/conversation/v1/conversation.proto

package pbconvoconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// ConversationServiceName is the fully-qualified name of the ConversationService service.
	ConversationServiceName = "sf.codegen.conversation.v1.ConversationService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ConversationServiceConverseProcedure is the fully-qualified name of the ConversationService's
	// Converse RPC.
	ConversationServiceConverseProcedure = "/sf.codegen.conversation.v1.ConversationService/Converse"
	// ConversationServiceDiscoverProcedure is the fully-qualified name of the ConversationService's
	// Discover RPC.
	ConversationServiceDiscoverProcedure = "/sf.codegen.conversation.v1.ConversationService/Discover"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	conversationServiceServiceDescriptor        = v1.File_sf_codegen_conversation_v1_conversation_proto.Services().ByName("ConversationService")
	conversationServiceConverseMethodDescriptor = conversationServiceServiceDescriptor.Methods().ByName("Converse")
	conversationServiceDiscoverMethodDescriptor = conversationServiceServiceDescriptor.Methods().ByName("Discover")
)

// ConversationServiceClient is a client for the sf.codegen.conversation.v1.ConversationService
// service.
type ConversationServiceClient interface {
	Converse(context.Context) *connect.BidiStreamForClient[v1.UserInput, v1.SystemOutput]
	Discover(context.Context, *connect.Request[v1.DiscoveryRequest]) (*connect.Response[v1.DiscoveryResponse], error)
}

// NewConversationServiceClient constructs a client for the
// sf.codegen.conversation.v1.ConversationService service. By default, it uses the Connect protocol
// with the binary Protobuf Codec, asks for gzipped responses, and sends uncompressed requests. To
// use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or connect.WithGRPCWeb()
// options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewConversationServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ConversationServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &conversationServiceClient{
		converse: connect.NewClient[v1.UserInput, v1.SystemOutput](
			httpClient,
			baseURL+ConversationServiceConverseProcedure,
			connect.WithSchema(conversationServiceConverseMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		discover: connect.NewClient[v1.DiscoveryRequest, v1.DiscoveryResponse](
			httpClient,
			baseURL+ConversationServiceDiscoverProcedure,
			connect.WithSchema(conversationServiceDiscoverMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// conversationServiceClient implements ConversationServiceClient.
type conversationServiceClient struct {
	converse *connect.Client[v1.UserInput, v1.SystemOutput]
	discover *connect.Client[v1.DiscoveryRequest, v1.DiscoveryResponse]
}

// Converse calls sf.codegen.conversation.v1.ConversationService.Converse.
func (c *conversationServiceClient) Converse(ctx context.Context) *connect.BidiStreamForClient[v1.UserInput, v1.SystemOutput] {
	return c.converse.CallBidiStream(ctx)
}

// Discover calls sf.codegen.conversation.v1.ConversationService.Discover.
func (c *conversationServiceClient) Discover(ctx context.Context, req *connect.Request[v1.DiscoveryRequest]) (*connect.Response[v1.DiscoveryResponse], error) {
	return c.discover.CallUnary(ctx, req)
}

// ConversationServiceHandler is an implementation of the
// sf.codegen.conversation.v1.ConversationService service.
type ConversationServiceHandler interface {
	Converse(context.Context, *connect.BidiStream[v1.UserInput, v1.SystemOutput]) error
	Discover(context.Context, *connect.Request[v1.DiscoveryRequest]) (*connect.Response[v1.DiscoveryResponse], error)
}

// NewConversationServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewConversationServiceHandler(svc ConversationServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	conversationServiceConverseHandler := connect.NewBidiStreamHandler(
		ConversationServiceConverseProcedure,
		svc.Converse,
		connect.WithSchema(conversationServiceConverseMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	conversationServiceDiscoverHandler := connect.NewUnaryHandler(
		ConversationServiceDiscoverProcedure,
		svc.Discover,
		connect.WithSchema(conversationServiceDiscoverMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/sf.codegen.conversation.v1.ConversationService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ConversationServiceConverseProcedure:
			conversationServiceConverseHandler.ServeHTTP(w, r)
		case ConversationServiceDiscoverProcedure:
			conversationServiceDiscoverHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedConversationServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedConversationServiceHandler struct{}

func (UnimplementedConversationServiceHandler) Converse(context.Context, *connect.BidiStream[v1.UserInput, v1.SystemOutput]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("sf.codegen.conversation.v1.ConversationService.Converse is not implemented"))
}

func (UnimplementedConversationServiceHandler) Discover(context.Context, *connect.Request[v1.DiscoveryRequest]) (*connect.Response[v1.DiscoveryResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("sf.codegen.conversation.v1.ConversationService.Discover is not implemented"))
}
