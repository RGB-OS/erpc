package erpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/erpc/erpc/auth"
	"github.com/erpc/erpc/common"
	"github.com/erpc/erpc/health"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

type HttpServer struct {
	config *common.ServerConfig
	server *fasthttp.Server
	erpc   *ERPC
	logger *zerolog.Logger
}

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func NewHttpServer(ctx context.Context, logger *zerolog.Logger, cfg *common.ServerConfig, erpc *ERPC) *HttpServer {
	reqMaxTimeout, err := time.ParseDuration(cfg.MaxTimeout)
	if err != nil {
		if cfg.MaxTimeout != "" {
			logger.Error().Err(err).Msgf("failed to parse max timeout duration using 30s default")
		}
		reqMaxTimeout = 30 * time.Second
	}

	srv := &HttpServer{
		config: cfg,
		erpc:   erpc,
		logger: logger,
	}

	srv.server = &fasthttp.Server{
		Handler: fasthttp.TimeoutHandler(
			srv.createRequestHandler(reqMaxTimeout),
			// This is the last resort timeout if nothing could be done in time
			reqMaxTimeout + 1 * time.Second,
			`{"jsonrpc":"2.0","error":{"code":-32603,"message":"request timeout before any upstream responded"}}`,
		),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(logger); err != nil {
			logger.Error().Msgf("http server forced to shutdown: %s", err)
		} else {
			logger.Info().Msg("http server stopped")
		}
	}()

	return srv
}

func (s *HttpServer) createRequestHandler(reqMaxTimeout time.Duration) fasthttp.RequestHandler {
	return func(fastCtx *fasthttp.RequestCtx) {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error().Msgf("unexpected server panic: %v", r)
			}
		}()

		buf := bufPool.Get().(*bytes.Buffer)
		defer bufPool.Put(buf)
		buf.Reset()
		encoder := json.NewEncoder(buf)

		segments := strings.Split(string(fastCtx.Path()), "/")
		if len(segments) != 2 && len(segments) != 3 && len(segments) != 4 {
			handleErrorResponse(s.logger, nil, common.NewErrInvalidUrlPath(string(fastCtx.Path())), fastCtx)
			return
		}

		projectId := segments[1]
		architecture, chainId := "", ""
		isAdmin := false

		if len(segments) == 4 {
			architecture = segments[2]
			chainId = segments[3]
		} else if len(segments) == 3 {
			if segments[2] == "admin" {
				isAdmin = true
			} else {
				handleErrorResponse(s.logger, nil, common.NewErrInvalidUrlPath(string(fastCtx.Path())), fastCtx)
				return
			}
		}

		project, err := s.erpc.GetProject(projectId)
		if err != nil {
			handleErrorResponse(s.logger, nil, err, fastCtx)
			return
		}

		if project.Config.CORS != nil {
			if !s.handleCORS(fastCtx, project.Config.CORS) {
				return
			}

			if fastCtx.IsOptions() {
				return
			}
		}

		body := fastCtx.PostBody()

		s.logger.Debug().Msgf("received request for projectId: %s, architecture: %s with body: %s", projectId, architecture, body)

		var requests []json.RawMessage
		err = sonic.Unmarshal(body, &requests)
		isBatch := err == nil

		if !isBatch {
			requests = []json.RawMessage{body}
		}

		responses := make([]interface{}, len(requests))
		var wg sync.WaitGroup

		var headersCopy fasthttp.RequestHeader
		var argsCopy fasthttp.Args
		fastCtx.Request.Header.CopyTo(&headersCopy)
		fastCtx.QueryArgs().CopyTo(&argsCopy)
  
		for i, reqBody := range requests {
			wg.Add(1)
			go func(index int, rawReq json.RawMessage, headersCopy *fasthttp.RequestHeader, argsCopy *fasthttp.Args) {
				defer func() {
					if r := recover(); r != nil {
						s.logger.Error().Msgf("unexpected server panic: %v", r)
					}
				}()
		
				defer wg.Done()

				requestCtx, cancel := context.WithTimeoutCause(fastCtx, reqMaxTimeout, common.NewErrRequestTimeout(reqMaxTimeout))
				defer cancel()

				nq := common.NewNormalizedRequest(rawReq)
				nq.ApplyDirectivesFromHttpHeaders(headersCopy)

				ap, err := auth.NewPayloadFromHttp(project.Config.Id, nq, headersCopy, argsCopy)
				if err != nil {
					responses[index] = processErrorBody(s.logger, nq, err)
					return
				}

				if isAdmin {
					if err := project.AuthenticateAdmin(requestCtx, nq, ap); err != nil {
						responses[index] = processErrorBody(s.logger, nq, err)
						return
					}
				} else {
					if err := project.AuthenticateConsumer(requestCtx, nq, ap); err != nil {
						responses[index] = processErrorBody(s.logger, nq, err)
						return
					}
				}

				if isAdmin {
					if project.Config.Admin != nil {
						resp, err := project.HandleAdminRequest(requestCtx, nq)
						if err != nil {
							responses[index] = processErrorBody(s.logger, nq, err)
							return
						}
						responses[index] = resp
						return
					} else {
						responses[index] = processErrorBody(
							s.logger,
							nq,
							common.NewErrAuthUnauthorized(
								"",
								"admin is not enabled for this project",
							),
						)
						return
					}
				}

				var networkId string

				if architecture == "" || chainId == "" {
					var req map[string]interface{}
					if err := sonic.Unmarshal(rawReq, &req); err != nil {
						responses[index] = processErrorBody(s.logger, nq, common.NewErrInvalidRequest(err))
						return
					}
					if networkIdFromBody, ok := req["networkId"].(string); ok {
						networkId = networkIdFromBody
						parts := strings.Split(networkId, ":")
						if len(parts) != 2 {
							responses[index] = processErrorBody(s.logger, nq, common.NewErrInvalidRequest(fmt.Errorf(
								"networkId must follow this format: 'architecture:chainId' for example 'evm:42161'",
							)))
							return
						}
						architecture = parts[0]
						chainId = parts[1]
					} else {
						responses[index] = processErrorBody(s.logger, nq, common.NewErrInvalidRequest(fmt.Errorf(
							"networkId must follow this format: 'architecture:chainId' for example 'evm:42161'",
						)))
						return
					}
				} else {
					networkId = fmt.Sprintf("%s:%s", architecture, chainId)
				}

				nw, err := project.GetNetwork(networkId)
				if err != nil {
					responses[index] = processErrorBody(s.logger, nq, err)
					return
				}
				nq.SetNetwork(nw)

				resp, err := project.Forward(requestCtx, networkId, nq)
				if err != nil {
					responses[index] = processErrorBody(s.logger, nq, err)
					return
				}

				responses[index] = resp
			}(i, reqBody, &headersCopy, &argsCopy)
		}

		wg.Wait()

		fastCtx.Response.Header.SetContentType("application/json")

		if isBatch {
			fastCtx.SetStatusCode(fasthttp.StatusOK)
			encoder.Encode(responses)
		} else {
			res := responses[0]

			var rm common.ResponseMetadata
			var ok bool
			rm, ok = res.(common.ResponseMetadata)
			if !ok {
				var jrsp, errObj map[string]interface{}
				if jrsp, ok = res.(map[string]interface{}); ok {
					if errObj, ok = jrsp["error"].(map[string]interface{}); ok {
						if err, ok = errObj["cause"].(error); ok {
							uer := &common.ErrUpstreamsExhausted{}
							if ok = errors.As(err, &uer); ok {
								rm = uer
							} else {
								uer := &common.ErrUpstreamRequest{}
								if ok = errors.As(err, &uer); ok {
									rm = uer
								}
							}
						}
					}
				}
			}

			if ok {
				if rm.FromCache() {
					fastCtx.Response.Header.Set("X-ERPC-Cache", "HIT")
				} else {
					fastCtx.Response.Header.Set("X-ERPC-Cache", "MISS")
				}
				if rm.UpstreamId() != "" {
					fastCtx.Response.Header.Set("X-ERPC-Upstream", rm.UpstreamId())
				}
				fastCtx.Response.Header.Set("X-ERPC-Attempts", fmt.Sprintf("%d", rm.Attempts()))
				fastCtx.Response.Header.Set("X-ERPC-Retries", fmt.Sprintf("%d", rm.Retries()))
				fastCtx.Response.Header.Set("X-ERPC-Hedges", fmt.Sprintf("%d", rm.Hedges()))
			}

			if err, ok := res.(error); ok {
				fastCtx.SetStatusCode(processErrorStatusCode(err))
			} else if resp, ok := res.(map[string]interface{}); ok {
				if errObj, ok := resp["error"].(map[string]interface{}); ok {
					if cause, ok := errObj["cause"].(error); ok {
						fastCtx.SetStatusCode(processErrorStatusCode(cause))
					} else {
						fastCtx.SetStatusCode(fasthttp.StatusOK)
					}
				} else {
					fastCtx.SetStatusCode(fasthttp.StatusOK)
				}
			} else {
				fastCtx.SetStatusCode(fasthttp.StatusOK)
			}
			encoder.Encode(res)
		}

		fastCtx.SetBody(buf.Bytes())
	}
}

func (s *HttpServer) handleCORS(ctx *fasthttp.RequestCtx, corsConfig *common.CORSConfig) bool {
	origin := string(ctx.Request.Header.Peek("Origin"))
	if origin == "" {
		return true
	}

	health.MetricCORSRequestsTotal.WithLabelValues(string(ctx.Path()), origin).Inc()

	allowed := false
	for _, allowedOrigin := range corsConfig.AllowedOrigins {
		if common.WildcardMatch(allowedOrigin, origin) {
			allowed = true
			break
		}
	}

	if !allowed {
		s.logger.Debug().Str("origin", origin).Msg("CORS request from disallowed origin")
		health.MetricCORSDisallowedOriginTotal.WithLabelValues(string(ctx.Path()), origin).Inc()

		if ctx.IsOptions() {
			ctx.SetStatusCode(fasthttp.StatusNoContent)
		} else {
			ctx.Error("CORS request from disallowed origin", fasthttp.StatusForbidden)
		}
		return false
	}

	ctx.Response.Header.Set("Access-Control-Allow-Origin", origin)
	ctx.Response.Header.Set("Access-Control-Allow-Methods", strings.Join(corsConfig.AllowedMethods, ", "))
	ctx.Response.Header.Set("Access-Control-Allow-Headers", strings.Join(corsConfig.AllowedHeaders, ", "))
	ctx.Response.Header.Set("Access-Control-Expose-Headers", strings.Join(corsConfig.ExposedHeaders, ", "))

	if corsConfig.AllowCredentials {
		ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	}

	if corsConfig.MaxAge > 0 {
		ctx.Response.Header.Set("Access-Control-Max-Age", fmt.Sprintf("%d", corsConfig.MaxAge))
	}

	if ctx.IsOptions() {
		health.MetricCORSPreflightRequestsTotal.WithLabelValues(string(ctx.Path()), origin).Inc()
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return false
	}

	return true
}

func processErrorBody(logger *zerolog.Logger, nq *common.NormalizedRequest, err error) interface{} {
	if !common.IsNull(err) {
		if common.HasErrorCode(err, common.ErrCodeEndpointClientSideException) {
			logger.Debug().Err(err).Msgf("forward request errored with client-side exception")
		} else {
			logger.Error().Err(err).Msgf("failed to forward request")
		}
	}

	// TODO extend this section to detect transport mode (besides json-rpc) when more modes are added.
	err = common.TranslateToJsonRpcException(err)
	var jsonrpcVersion string = "2.0"
	var reqId interface{} = nil
	if nq != nil {
		jrr, _ := nq.JsonRpcRequest()
		if jrr != nil {
			jsonrpcVersion = jrr.JSONRPC
			reqId = jrr.ID
		}
	}
	jre := &common.ErrJsonRpcExceptionInternal{}
	if errors.As(err, &jre) {
		return map[string]interface{}{
			"jsonrpc": jsonrpcVersion,
			"id":      reqId,
			"error": map[string]interface{}{
				"code":    jre.NormalizedCode(),
				"message": jre.Message,
				"data":    jre.Details["data"],
				"cause":   err,
			},
		}
	}

	if _, ok := err.(*common.BaseError); ok {
		return err
	} else if serr, ok := err.(common.StandardError); ok {
		return serr
	}

	return common.BaseError{
		Code:    "ErrUnknown",
		Message: "unexpected server error",
		Cause:   err,
	}
}

func processErrorStatusCode(err error) int {
	if e, ok := err.(common.StandardError); ok {
		return e.ErrorStatusCode()
	}
	return fasthttp.StatusInternalServerError
}

func handleErrorResponse(logger *zerolog.Logger, nq *common.NormalizedRequest, err error, ctx *fasthttp.RequestCtx) {
	if !common.IsNull(err) {
		if common.HasErrorCode(err, common.ErrCodeEndpointClientSideException) {
			logger.Debug().Err(err).Msgf("forward request errored with client-side exception")
		} else {
			logger.Error().Err(err).Msgf("failed to forward request")
		}
	}

	ctx.Response.Header.SetContentType("application/json")
	if e, ok := err.(common.StandardError); ok {
		ctx.SetStatusCode(e.ErrorStatusCode())
	} else {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	}

	err = common.TranslateToJsonRpcException(err)

	var jsonrpcVersion string = "2.0"
	var reqId interface{} = nil
	if nq != nil {
		jrr, _ := nq.JsonRpcRequest()
		if jrr != nil {
			jsonrpcVersion = jrr.JSONRPC
			reqId = jrr.ID
		}
	}

	jre := &common.ErrJsonRpcExceptionInternal{}
	if errors.As(err, &jre) {
		writeErr := json.NewEncoder(ctx).Encode(map[string]interface{}{
			"jsonrpc": jsonrpcVersion,
			"id":      reqId,
			"error": map[string]interface{}{
				"code":    jre.NormalizedCode(),
				"message": jre.Message,
				"data":    jre.Details["data"],
				"cause":   err,
			},
		})
		if writeErr != nil {
			logger.Error().Err(writeErr).Msgf("failed to encode error response body")
		}
		return
	}

	var writeErr error

	if _, ok := err.(*common.BaseError); ok {
		writeErr = json.NewEncoder(ctx).Encode(err)
	} else {
		if serr, ok := err.(common.StandardError); ok {
			writeErr = json.NewEncoder(ctx).Encode(serr)
		} else {
			writeErr = json.NewEncoder(ctx).Encode(
				common.BaseError{
					Code:    "ErrUnknown",
					Message: "unexpected server error",
					Cause:   err,
				},
			)
		}
	}

	if writeErr != nil {
		logger.Error().Err(writeErr).Msgf("failed to encode error response body")
	}
}

func (s *HttpServer) Start(logger *zerolog.Logger) error {
	addrV4 := fmt.Sprintf("%s:%d", s.config.HttpHostV4, s.config.HttpPort)
	addrV6 := fmt.Sprintf("%s:%d", s.config.HttpHostV6, s.config.HttpPort)

	var err error
	var ln net.Listener
	var ln4 net.Listener
	var ln6 net.Listener

	if s.config.HttpHostV4 != "" {
		logger.Info().Msgf("starting http server on port: %d IPv4: %s", s.config.HttpPort, addrV4)
		ln4, err = net.Listen("tcp4", addrV4)
		if err != nil {
			return fmt.Errorf("error listening on IPv4: %w", err)
		}
	}
	if s.config.HttpHostV6 != "" {
		logger.Info().Msgf("starting http server on port: %d IPv6: %s", s.config.HttpPort, addrV6)
		ln6, err = net.Listen("tcp6", addrV6)
		if err != nil {
			if ln4 != nil {
				ln4.Close()
			}
			return fmt.Errorf("error listening on IPv6: %w", err)
		}
	}

	if ln4 != nil && ln6 != nil {
		ln = &dualStackListener{ln4, ln6}
	} else if ln4 != nil {
		ln = ln4
	} else if ln6 != nil {
		ln = ln6
	}

	if ln == nil {
		return fmt.Errorf("you must configure at least one of server.httpPortV4 or server.httpPortV6")
	}

	return s.server.Serve(ln)
}

func (s *HttpServer) Shutdown(logger *zerolog.Logger) error {
	logger.Info().Msg("stopping http server...")
	return s.server.Shutdown()
}

type dualStackListener struct {
	ln4, ln6 net.Listener
}

func (dl *dualStackListener) Accept() (net.Conn, error) {
	// Use a channel to communicate the result of Accept
	type result struct {
		conn net.Conn
		err  error
	}
	ch := make(chan result, 2)

	// Try to accept from both listeners
	go func() {
		conn, err := dl.ln4.Accept()
		ch <- result{conn, err}
	}()
	go func() {
		conn, err := dl.ln6.Accept()
		ch <- result{conn, err}
	}()

	// Return the first successful connection, or the last error
	var lastErr error
	for i := 0; i < 2; i++ {
		res := <-ch
		if res.err == nil {
			return res.conn, nil
		}
		lastErr = res.err
	}
	return nil, lastErr
}

func (dl *dualStackListener) Close() error {
	err4 := dl.ln4.Close()
	err6 := dl.ln6.Close()
	if err4 != nil {
		return err4
	}
	return err6
}

func (dl *dualStackListener) Addr() net.Addr {
	return dl.ln4.Addr()
}
