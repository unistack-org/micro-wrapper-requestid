package requestid // import "go.unistack.org/micro-wrapper-requestid/v4"

import (
	"context"
	"net/textproto"

	"go.unistack.org/micro/v4/client"
	"go.unistack.org/micro/v4/metadata"
	"go.unistack.org/micro/v4/options"
	"go.unistack.org/micro/v4/server"
	"go.unistack.org/micro/v4/util/id"
)

var XRequestIDKey struct{}

// DefaultMetadataKey contains metadata key x-request-id
var DefaultMetadataKey = textproto.CanonicalMIMEHeaderKey("x-request-id")

// DefaultMetadataFunc wil be used if user not provide own func to fill metadata
var DefaultMetadataFunc = func(ctx context.Context) (context.Context, error) {
	var xid string
	var err error
	var ook, iok bool

	if _, ok := ctx.Value(XRequestIDKey).(string); !ok {
		xid, err = id.New()
		if err != nil {
			return ctx, err
		}
		ctx = context.WithValue(ctx, XRequestIDKey, xid)
	}

	imd, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		imd = metadata.New(1)
		imd.Set(DefaultMetadataKey, xid)
	} else if _, ok = imd.Get(DefaultMetadataKey); !ok {
		imd.Set(DefaultMetadataKey, xid)
	} else {
		iok = true
	}

	omd, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		omd = metadata.New(1)
		omd.Set(DefaultMetadataKey, xid)
	} else if _, ok = omd.Get(DefaultMetadataKey); !ok {
		omd.Set(DefaultMetadataKey, xid)
	} else {
		ook = true
	}

	if !iok {
		ctx = metadata.NewIncomingContext(ctx, imd)
	}
	if !ook {
		ctx = metadata.NewOutgoingContext(ctx, omd)
	}

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
			if ctx, err = DefaultMetadataFunc(ctx); err != nil {
				return err
			}
			return fn(ctx, addr, req, rsp, opts)
		}
	}
}

func (w *wrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...options.Option) error {
	var err error
	if ctx, err = DefaultMetadataFunc(ctx); err != nil {
		return err
	}
	return w.Client.Call(ctx, req, rsp, opts...)
}

func (w *wrapper) Stream(ctx context.Context, req client.Request, opts ...options.Option) (client.Stream, error) {
	var err error
	if ctx, err = DefaultMetadataFunc(ctx); err != nil {
		return nil, err
	}
	return w.Client.Stream(ctx, req, opts...)
}

func NewServerHandlerWrapper() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			var err error
			if ctx, err = DefaultMetadataFunc(ctx); err != nil {
				return err
			}
			return fn(ctx, req, rsp)
		}
	}
}
