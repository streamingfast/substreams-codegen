// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: sf/codegen/remotebuild/v1/remotebuild.proto

package pbbuildconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "github.com/streamingfast/substreams-codegen/pb/sf/codegen/remotebuild/v1"
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
	// BuildServiceName is the fully-qualified name of the BuildService service.
	BuildServiceName = "sf.remotebuild.v1.BuildService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// BuildServiceBuildProcedure is the fully-qualified name of the BuildService's Build RPC.
	BuildServiceBuildProcedure = "/sf.remotebuild.v1.BuildService/Build"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	buildServiceServiceDescriptor     = v1.File_sf_codegen_remotebuild_v1_remotebuild_proto.Services().ByName("BuildService")
	buildServiceBuildMethodDescriptor = buildServiceServiceDescriptor.Methods().ByName("Build")
)

// BuildServiceClient is a client for the sf.remotebuild.v1.BuildService service.
type BuildServiceClient interface {
	Build(context.Context, *connect.Request[v1.BuildRequest]) (*connect.ServerStreamForClient[v1.BuildResponse], error)
}

// NewBuildServiceClient constructs a client for the sf.remotebuild.v1.BuildService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewBuildServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) BuildServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &buildServiceClient{
		build: connect.NewClient[v1.BuildRequest, v1.BuildResponse](
			httpClient,
			baseURL+BuildServiceBuildProcedure,
			connect.WithSchema(buildServiceBuildMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// buildServiceClient implements BuildServiceClient.
type buildServiceClient struct {
	build *connect.Client[v1.BuildRequest, v1.BuildResponse]
}

// Build calls sf.remotebuild.v1.BuildService.Build.
func (c *buildServiceClient) Build(ctx context.Context, req *connect.Request[v1.BuildRequest]) (*connect.ServerStreamForClient[v1.BuildResponse], error) {
	return c.build.CallServerStream(ctx, req)
}

// BuildServiceHandler is an implementation of the sf.remotebuild.v1.BuildService service.
type BuildServiceHandler interface {
	Build(context.Context, *connect.Request[v1.BuildRequest], *connect.ServerStream[v1.BuildResponse]) error
}

// NewBuildServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewBuildServiceHandler(svc BuildServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	buildServiceBuildHandler := connect.NewServerStreamHandler(
		BuildServiceBuildProcedure,
		svc.Build,
		connect.WithSchema(buildServiceBuildMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/sf.remotebuild.v1.BuildService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case BuildServiceBuildProcedure:
			buildServiceBuildHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedBuildServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedBuildServiceHandler struct{}

func (UnimplementedBuildServiceHandler) Build(context.Context, *connect.Request[v1.BuildRequest], *connect.ServerStream[v1.BuildResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("sf.remotebuild.v1.BuildService.Build is not implemented"))
}
