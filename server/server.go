package server

import (
	_ "embed"
	"net/http"
	"regexp"
	"strings"

	"connectrpc.com/connect"
	dgrpcserver "github.com/streamingfast/dgrpc/server"
	connectweb "github.com/streamingfast/dgrpc/server/connectrpc"
	"github.com/streamingfast/shutter"
	"github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1/pbconvoconnect"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type server struct {
	*shutter.Shutter
	httpListenAddr     string
	connectWebServer   *connectweb.ConnectWebServer
	corsHostRegexAllow *regexp.Regexp
	logger             *zap.Logger
}

func New(
	httpListenAddr string,
	corsHostRegexAllow *regexp.Regexp,
	logger *zap.Logger,
) *server {
	return &server{
		Shutter:            shutter.New(),
		httpListenAddr:     httpListenAddr,
		corsHostRegexAllow: corsHostRegexAllow,
		logger:             logger,
	}
}

func (s *server) Run() {
	s.logger.Info("starting server")

	tracerProvider := otel.GetTracerProvider()
	options := []dgrpcserver.Option{
		dgrpcserver.WithLogger(s.logger),
		dgrpcserver.WithHealthCheck(dgrpcserver.HealthCheckOverGRPC|dgrpcserver.HealthCheckOverHTTP, s.healthzHandler()),
		dgrpcserver.WithPostUnaryInterceptor(otelgrpc.UnaryServerInterceptor(otelgrpc.WithTracerProvider(tracerProvider))),
		dgrpcserver.WithPostStreamInterceptor(otelgrpc.StreamServerInterceptor(otelgrpc.WithTracerProvider(tracerProvider))),
		dgrpcserver.WithGRPCServerOptions(grpc.MaxRecvMsgSize(150 * 1024 * 1024)),
		dgrpcserver.WithConnectReflection("sf.codegen.conversation.v1.ConversationService"),
		dgrpcserver.WithConnectCORS(s.corsOption()),
		//dgrpcserver.WithConnectWebHTTPHandlers([]dgrpcserver.HTTPHandlerGetter{
		//	func() (string, http.Handler) {
		//		return "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//			http.Redirect(w, r, "https://codegen.substreams.dev", http.StatusNotFound)
		//		})
		//	},
		//}),
	}
	if strings.Contains(s.httpListenAddr, "*") {
		s.logger.Info("grpc server with insecure server")
		options = append(options, dgrpcserver.WithInsecureServer())
	} else {
		s.logger.Info("grpc server with plain text server")
		options = append(options, dgrpcserver.WithPlainTextServer())
	}

	convoHandlerGetter := func(opts ...connect.HandlerOption) (string, http.Handler) {
		return pbconvoconnect.NewConversationServiceHandler(s, opts...)
	}

	srv := connectweb.New([]connectweb.HandlerGetter{convoHandlerGetter}, options...)
	addr := strings.ReplaceAll(s.httpListenAddr, "*", "")

	s.OnTerminating(func(err error) {
		s.logger.Info("shutting down connect web server")
		srv.Shutdown(nil)
	})

	srv.Launch(addr)
	<-srv.Terminated()
}
