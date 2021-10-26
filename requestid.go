package requestid // import "go.unistack.org/micro-wrapper-requestid/v3"

import (
	"context"
	"net/textproto"

	"go.unistack.org/micro/v3/client"
	"go.unistack.org/micro/v3/metadata"
	"go.unistack.org/micro/v3/server"
	"go.unistack.org/micro/v3/util/id"
)

// MetadataKey contains metadata key
var MetadataKey = textproto.CanonicalMIMEHeaderKey("x-request-id")

// MetadataFunc wil be used if user not provide own func to fill metadata
var MetadataFunc = func(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(1)
	}
	if _, ok = md.Get(MetadataKey); ok {
		return ctx, nil
	}
	uid, err := id.New()
	if err != nil {
		return ctx, err
	}
	md.Set(MetadataKey, uid)
	ctx = metadata.NewIncomingContext(ctx, md)
	return ctx, nil
}

type wrapper struct {
	client.Client
}

func NewClientWrapper() client.Wrapper {
	return func(c client.Client) client.Client {
		handler := &wrapper{
			Client: c,
		}
		return handler
	}
}

func NewClientCallWrapper() client.CallWrapper {
	return func(fn client.CallFunc) client.CallFunc {
		return func(ctx context.Context, addr string, req client.Request, rsp interface{}, opts client.CallOptions) error {
			var err error
			if ctx, err = MetadataFunc(ctx); err != nil {
				return err
			}
			return fn(ctx, addr, req, rsp, opts)
		}
	}
}

func (w *wrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	var err error
	if ctx, err = MetadataFunc(ctx); err != nil {
		return err
	}
	return w.Client.Call(ctx, req, rsp, opts...)
}

func (w *wrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	var err error
	if ctx, err = MetadataFunc(ctx); err != nil {
		return nil, err
	}
	return w.Client.Stream(ctx, req, opts...)
}

func (w *wrapper) Publish(ctx context.Context, msg client.Message, opts ...client.PublishOption) error {
	var err error
	if ctx, err = MetadataFunc(ctx); err != nil {
		return err
	}
	return w.Client.Publish(ctx, msg, opts...)
}

func NewServerHandlerWrapper() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			var err error
			if ctx, err = MetadataFunc(ctx); err != nil {
				return err
			}
			return fn(ctx, req, rsp)
		}
	}
}

func NewServerSubscriberWrapper() server.SubscriberWrapper {
	return func(fn server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, msg server.Message) error {
			var err error
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				md = metadata.New(1)
			}
			if id, ok := msg.Header()[MetadataKey]; ok {
				md.Set(MetadataKey, id)
				ctx = metadata.NewIncomingContext(ctx, md)
			} else if ctx, err = MetadataFunc(ctx); err != nil {
				return err
			}
			return fn(ctx, msg)
		}
	}
}
