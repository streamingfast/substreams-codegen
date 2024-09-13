package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"
	"time"

	connect "connectrpc.com/connect"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	_ "github.com/streamingfast/substreams-codegen/evm-events-calls"
	_ "github.com/streamingfast/substreams-codegen/evm-minimal"
	_ "github.com/streamingfast/substreams-codegen/injective-events"
	_ "github.com/streamingfast/substreams-codegen/injective-minimal"
	_ "github.com/streamingfast/substreams-codegen/sol-minimal"
	_ "github.com/streamingfast/substreams-codegen/sol-transactions"
	_ "github.com/streamingfast/substreams-codegen/starknet-events"
	_ "github.com/streamingfast/substreams-codegen/starknet-minimal"
	_ "github.com/streamingfast/substreams-codegen/vara-extrinsics"
	_ "github.com/streamingfast/substreams-codegen/vara-minimal"
)

func (s *server) Discover(ctx context.Context, req *connect.Request[pbconvo.DiscoveryRequest]) (*connect.Response[pbconvo.DiscoveryResponse], error) {
	var generators []*pbconvo.DiscoveryResponse_Generator
	for _, conv := range codegen.ListConversationHandlers() {
		generators = append(generators, &pbconvo.DiscoveryResponse_Generator{
			Id:          conv.ID,
			Title:       conv.Title,
			Description: conv.Description,
		})

	}
	return connect.NewResponse(&pbconvo.DiscoveryResponse{
		Generators: generators,
	}), nil
}

type eventLogger struct {
	loggedEvents []string
}

func (e *eventLogger) logEvent(event string) {
	if os.Getenv("SUBSTREAMS_DEV_DEBUG_EVENTS") == "true" {
		fmt.Println(event)
	}
	e.loggedEvents = append(e.loggedEvents, event)
}

func (s *server) Converse(ctx context.Context, stream *connect.BidiStream[pbconvo.UserInput, pbconvo.SystemOutput]) (err error) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("internal error first defer", zap.Any("panic", r))
			err = fmt.Errorf("internal error: %v", r)
		}
	}()

	s.logger.Info("new conversation")
	closeOnce := sync.Once{}
	sendFunc := func(msg *pbconvo.SystemOutput, err error) {
		if msg == nil {
			closeOnce.Do(func() {
				if closer, ok := stream.Conn().(interface{ Close(error) error }); ok {
					closer.Close(err)
				}
			})
		}
		stream.Send(msg)
	}

	req, err := stream.Receive()
	if err != nil {
		return err
	}

	start, ok := req.Entry.(*pbconvo.UserInput_Start_)
	if !ok {
		return fmt.Errorf("begin with UserInput_Start message")
	}
	if start.Start.Version < 1 {
		return fmt.Errorf("unsupported protocol version %d, please upgrade your `substreams` client", start.Start.Version)
	}

	convo := codegen.Registry[start.Start.GeneratorId]
	if convo == nil {
		return fmt.Errorf("no conversation handler found for topic ID %q", start.Start.GeneratorId)
	}

	evts := &eventLogger{}
	begin := time.Now()
	s.logger.Info("launching thread")
	evts.logEvent(fmt.Sprintf("   0┃ [Start, hydrate: %t] %s", start.Start.Hydrate != nil, start.Start.GeneratorId))

	msgWrapFactory := codegen.NewMsgWrapFactory(sendFunc)
	conversation := convo.Factory()
	conversation.SetFactory(msgWrapFactory)

	readNextCmd := func() loop.Msg {
		select {
		case <-ctx.Done():
			return loop.NewQuitMsg(ctx.Err())
		default:
		}

		req, err := stream.Receive()
		if err != nil {
			return loop.NewQuitMsg(err)
		}

		reflectType := msgWrapFactory.LastInput()
		if reflectType == nil {
			// TODO: make this a "BadRequest" or InvalidRequest error, shown to the user
			return loop.NewQuitMsg(fmt.Errorf("message type %q was not registered or does not exist", req.FromActionId))
		}
		newMsg := reflect.New(reflectType)
		newProtoMsg := newMsg.Interface().(protoreflect.ProtoMessage)

		switch entry := req.Entry.(type) {
		case *pbconvo.UserInput_Confirmation_:
			cnt, err := proto.Marshal(entry.Confirmation)
			if err != nil {
				return loop.NewQuitMsg(fmt.Errorf("marshal type %T: %w", entry.Confirmation, err))
			}
			err = proto.Unmarshal(cnt, newProtoMsg)
			if err != nil {
				return loop.NewQuitMsg(fmt.Errorf("unmarshal into type %T from %T: %w", newProtoMsg, entry.Confirmation, err))
			}
			return codegen.IncomingMessage{Msg: newMsg.Elem().Interface()}

		case *pbconvo.UserInput_Selection_:
			cnt, err := proto.Marshal(entry.Selection)
			if err != nil {
				return loop.NewQuitMsg(fmt.Errorf("marshal type %T: %w", entry.Selection, err))
			}
			err = proto.Unmarshal(cnt, newProtoMsg)
			if err != nil {
				return loop.NewQuitMsg(fmt.Errorf("unmarshal into type %T from %T: %w", newProtoMsg, entry.Selection, err))
			}
			return codegen.IncomingMessage{Msg: newMsg.Elem().Interface()}

		case *pbconvo.UserInput_TextInput_:
			cnt, err := proto.Marshal(entry.TextInput)
			if err != nil {
				return loop.NewQuitMsg(fmt.Errorf("marshal type %T: %w", entry.TextInput, err))
			}
			err = proto.Unmarshal(cnt, newProtoMsg)
			if err != nil {
				return loop.NewQuitMsg(fmt.Errorf("unmarshal into type %T from %T: %w", newProtoMsg, entry.TextInput, err))
			}
			return codegen.IncomingMessage{Msg: newMsg.Elem().Interface()}

		case *pbconvo.UserInput_DownloadedFiles_:
			cnt, err := proto.Marshal(entry.DownloadedFiles)
			if err != nil {
				return loop.NewQuitMsg(fmt.Errorf("marshal type %T: %w", entry.DownloadedFiles, err))
			}
			err = proto.Unmarshal(cnt, newProtoMsg)
			if err != nil {
				return loop.NewQuitMsg(fmt.Errorf("unmarshal into type %T from %T: %w", newProtoMsg, entry.DownloadedFiles, err))
			}
			return codegen.IncomingMessage{Msg: newMsg.Elem().Interface()}

		case *pbconvo.UserInput_File:
			return loop.NewQuitMsg(fmt.Errorf("file upload not supported here"))

		default:
			return loop.NewQuitMsg(fmt.Errorf("unknown entry type %T", entry))
		}
	}

	initCmd := loop.Batch(
		func() loop.Msg {
			return codegen.MsgStart{UserInput_Start: *start.Start}
		},
		readNextCmd,
	)

	var lastMessageIsIncoming bool
	var lastState string
	msgWrapFactory.SetupLoop(func(msg loop.Msg) loop.Cmd {
		asJSON, _ := json.Marshal(msg)
		asJSON, _ = sjson.DeleteBytes(asJSON, "state")
		s.logger.Debug("main Loop", zap.Any("loop_msg_type", msg), zap.String("content", string(asJSON)))
		switch msg := msg.(type) {
		case *pbconvo.SystemOutput:
			ev := msg.Humanize(int(time.Since(begin).Seconds()))
			if lastMessageIsIncoming {
				ev = "\n" + ev
				lastMessageIsIncoming = false
			}
			evts.logEvent(ev)
			lastState = msg.State
			sendFunc(msg, nil)
			return nil
		case codegen.IncomingMessage:
			lastMessageIsIncoming = true
			evts.logEvent("\n" + msg.Humanize(int(time.Since(begin).Seconds())))
			return loop.Batch(func() loop.Msg { return msg.Msg }, readNextCmd)
		}

		s.logger.Debug("updating")
		if os.Getenv("SUBSTREAMS_DEV_DEBUG_CONVERSATION") == "true" {
			fmt.Printf("convo Update message: %T %#v\n-> state: %#v\n\n", msg, msg, conversation.GetState())
		}

		cmd := conversation.Update(msg)
		return cmd
	})

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		defer func() error {
			if r := recover(); r != nil {
				s.logger.Error("internal error second defer", zap.Any("panic", r))
				err = fmt.Errorf("internal error: %v", r)
				return err
			}

			return nil
		}()

		err = msgWrapFactory.Run(ctx, initCmd)
		if err != nil {
			evts.logEvent(fmt.Sprintf("ERROR %q AFTER %d seconds", err.Error(), int(time.Since(begin).Seconds())))
			s.sessionLogger.SaveSession(start.Start.GeneratorId, evts.loggedEvents, lastState)
			s.logger.Warn("failed to save session", zap.Error(err))
			return err
		}
		evts.logEvent(fmt.Sprintf("COMPLETED IN %d seconds", int(time.Since(begin).Seconds())))

		if err := s.sessionLogger.SaveSession(start.Start.GeneratorId, evts.loggedEvents, lastState); err != nil {
			s.logger.Warn("failed to save session", zap.Error(err))
		}
		return io.EOF
	})

	if err != nil && errors.Is(err, io.EOF) {
		return fmt.Errorf("cound not generate substreams: %w", err)
	}

	if err := g.Wait(); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("conversation error: %w", err)
	}

	if err != nil {
		return err
	}

	return nil
}
