package ctxspan

import (
	"context"
	"fmt"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/remote/transmeta"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/utils"
	"github.com/cloudwego/kitex/transport"
	"strconv"
)

const TTHeader = transport.TTHeader

const (
	framedTransportType   = "framed"
	unframedTransportType = "unframed"

	// for biz error
	bizStatus  = "biz-status"
	bizMessage = "biz-message"
	bizExtra   = "biz-extra"
)

// TTHeader handlers.
var (
	ClientTTHeaderHandler remote.MetaHandler = &clientTTHeaderHandler{}
	ServerTTHeaderHandler remote.MetaHandler = &serverTTHeaderHandler{}
)

// clientTTHeaderHandler implement remote.MetaHandler
type clientTTHeaderHandler struct{}

// WriteMeta of clientTTHeaderHandler writes headers of TTHeader protocol to transport
func (ch *clientTTHeaderHandler) WriteMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	if !isTTHeader(msg) {
		return ctx, nil
	}
	ri := msg.RPCInfo()
	transInfo := msg.TransInfo()

	hd := map[uint16]string{
		transmeta.FromService: ri.From().ServiceName(),
		transmeta.FromMethod:  ri.From().Method(),
		transmeta.ToService:   ri.To().ServiceName(),
		transmeta.ToMethod:    ri.To().Method(),
		transmeta.MsgType:     strconv.Itoa(int(msg.MessageType())),
		transmeta.SpanContext: GetSpanStringFromContext(ctx),
	}
	if msg.ProtocolInfo().TransProto&transport.Framed == transport.Framed {
		hd[transmeta.TransportType] = framedTransportType
	} else {
		hd[transmeta.TransportType] = unframedTransportType
	}

	transInfo.PutTransIntInfo(hd)
	return ctx, nil
}

// ReadMeta of clientTTHeaderHandler reads headers of TTHeader protocol from transport
func (ch *clientTTHeaderHandler) ReadMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	if !isTTHeader(msg) {
		return ctx, nil
	}
	ri := msg.RPCInfo()
	transInfo := msg.TransInfo()
	strInfo := transInfo.TransStrInfo()
	intInfo := transInfo.TransIntInfo()

	if spanStr, ok := intInfo[transmeta.SpanContext]; ok && spanStr != "" {
		traceId, spanId := GetTraceIDAndSpanIDFromSpanString(spanStr)
		if traceId != "" && spanId != "" {
			ctx = SetSpanContextWithId(ctx, traceId, spanId)
		}
	}

	if code, err := strconv.Atoi(strInfo[bizStatus]); err == nil && code != 0 {
		if setter, ok := ri.Invocation().(rpcinfo.InvocationSetter); ok {
			if bizExtra := strInfo[bizExtra]; bizExtra != "" {
				extra, err := utils.JSONStr2Map(bizExtra)
				if err != nil {
					return ctx, fmt.Errorf("malformed header info, extra: %s", bizExtra)
				}
				setter.SetBizStatusErr(kerrors.NewBizStatusErrorWithExtra(int32(code), strInfo[bizMessage], extra))
			} else {
				setter.SetBizStatusErr(kerrors.NewBizStatusError(int32(code), strInfo[bizMessage]))
			}
		}
	}

	return ctx, nil
}

// serverTTHeaderHandler implement remote.MetaHandler
type serverTTHeaderHandler struct{}

// ReadMeta of serverTTHeaderHandler reads headers of TTHeader protocol to transport
func (sh *serverTTHeaderHandler) ReadMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	if !isTTHeader(msg) {
		return ctx, nil
	}
	ri := msg.RPCInfo()
	transInfo := msg.TransInfo()
	intInfo := transInfo.TransIntInfo()

	if spanStr, ok := intInfo[transmeta.SpanContext]; ok && spanStr != "" {
		traceId, spanId := GetTraceIDAndSpanIDFromSpanString(spanStr)
		if traceId != "" && spanId != "" {
			ctx = SetSpanContextWithId(ctx, traceId, spanId)
		}
	}

	ci := rpcinfo.AsMutableEndpointInfo(ri.From())
	if ci != nil {
		if v := intInfo[transmeta.FromService]; v != "" {
			ci.SetServiceName(v)
		}
		if v := intInfo[transmeta.FromMethod]; v != "" {
			ci.SetMethod(v)
		}
	}
	return ctx, nil
}

// WriteMeta of serverTTHeaderHandler writes headers of TTHeader protocol to transport
func (sh *serverTTHeaderHandler) WriteMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	if !isTTHeader(msg) {
		return ctx, nil
	}
	ri := msg.RPCInfo()
	transInfo := msg.TransInfo()
	intInfo := transInfo.TransIntInfo()
	strInfo := transInfo.TransStrInfo()

	intInfo[transmeta.SpanContext] = GetSpanStringFromContext(ctx)

	intInfo[transmeta.MsgType] = strconv.Itoa(int(msg.MessageType()))

	if bizErr := ri.Invocation().BizStatusErr(); bizErr != nil {
		strInfo[bizStatus] = strconv.Itoa(int(bizErr.BizStatusCode()))
		strInfo[bizMessage] = bizErr.BizMessage()
		if len(bizErr.BizExtra()) != 0 {
			strInfo[bizExtra], _ = utils.Map2JSONStr(bizErr.BizExtra())
		}
	}

	return ctx, nil
}

func isTTHeader(msg remote.Message) bool {
	transProto := msg.ProtocolInfo().TransProto
	return transProto&transport.TTHeader == transport.TTHeader
}
