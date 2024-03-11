package requestid

import (
	"context"
	"net/textproto"
	"strings"

	"go.unistack.org/micro/v4/client"
	"go.unistack.org/micro/v4/logger"
	"go.unistack.org/micro/v4/metadata"
	"go.unistack.org/micro/v4/options"
	"go.unistack.org/micro/v4/server"
	"go.unistack.org/micro/v4/util/id"
)

func init() {
	requestIDLog := strings.ToLower(DefaultMetadataKey)
	logger.DefaultContextAttrFuncs = append(logger.DefaultContextAttrFuncs, func(ctx context.Context) []interface{} {
		if v, ok := ctx.Value(XRequestIDKey{}).(string); ok {
			return []interface{}{requestIDLog, v}
		}
		return nil
	})
}

type XRequestIDKey struct{}

// DefaultMetadataKey contains metadata key x-request-id
var DefaultMetadataKey = textproto.CanonicalMIMEHeaderKey("x-request-id")

// DefaultMetadataFunc wil be used if user not provide own func to fill metadata
var DefaultMetadataFunc = func(ctx context.Context) (context.Context, error) {
	var xid string

	cid, cok := ctx.Value(XRequestIDKey{}).(string)
	if cok && cid != "" {
		xid = cid
	}

	imd, iok := metadata.FromIncomingContext(ctx)
	if !iok || imd == nil {
		imd = metadata.New(1)
		ctx = metadata.NewIncomingContext(ctx, imd)
	}

	omd, ook := metadata.FromOutgoingContext(ctx)
	if !ook || omd == nil {
		omd = metadata.New(1)
		ctx = metadata.NewOutgoingContext(ctx, omd)
	}

	if xid == "" {
		var id string
		if id, iok = imd.Get(DefaultMetadataKey); iok && id != "" {
			xid = id
		}
		if id, ook = omd.Get(DefaultMetadataKey); ook && id != "" {
			xid = id
		}
	}

	if xid == "" {
		var err error
		xid, err = id.New()
		if err != nil {
			return ctx, err
		}
	}

	if !cok {
		ctx = context.WithValue(ctx, XRequestIDKey{}, xid)
	}

	if !iok {
		imd.Set(DefaultMetadataKey, xid)
	}

	if !ook {
		omd.Set(DefaultMetadataKey, xid)
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
