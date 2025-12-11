package server

import (
	"forgetting-curve/backend/api/student/v1"
	"forgetting-curve/backend/internal/biz"
	"forgetting-curve/backend/internal/conf"
	"forgetting-curve/backend/internal/server/middleware"
	"forgetting-curve/backend/internal/service"
	"github.com/go-kratos/aegis/ratelimit/bbr"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	httptransport "github.com/go-kratos/kratos/v2/transport/http"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"net/http"
	"reflect"

	"github.com/go-kratos/kratos/v2/middleware/recovery"
)

// NewHTTPServer new a HTTP server.
func NewHTTPServer(c *conf.Server, studentService *service.StudentService, ocrService biz.OCRService, logger log.Logger) *httptransport.Server {
	addr := ":0"
	if c != nil && c.HTTP != nil && c.HTTP.Addr != "" {
		addr = c.HTTP.Addr
	}

	opts := []httptransport.ServerOption{
		httptransport.Address(addr),
		httptransport.ResponseEncoder(responseEncoder),
		httptransport.ErrorEncoder(errorEncoder),
	}
	opts = append(opts, httptransport.Middleware(
		recovery.Recovery(),
		tracing.Server(
			tracing.WithTracerProvider(otel.GetTracerProvider()),
			tracing.WithPropagator(
				propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{}),
			),
		),
		logging.Server(logger),
		middleware.ServerMetrics(),
		middleware.Validator(),
		middleware.TraceparentMiddleware(),
		middleware.MetaData(),
		ratelimit.Server(ratelimit.WithLimiter(bbr.NewLimiter())),
	))

	srv := httptransport.NewServer(opts...)
	v1.RegisterStudentServiceHTTPServer(srv, studentService)

	// 注册 OCR 路由
	if ocrService != nil {
		ocrHandler := NewOCRHandler(ocrService, logger)
		srv.Route("/api/ocr").POST("", func(ctx httptransport.Context) error {
			ocrHandler.HandleOCR(ctx.Response(), ctx.Request())
			return nil
		})
	}

	return srv
}

type BizResp interface {
	GetRet() *v1.BaseResponse
}

func responseEncoder(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if v == nil {
		return nil
	}
	if rd, ok := v.(httptransport.Redirector); ok {
		url, code := rd.Redirect()
		http.Redirect(w, r, url, code)
		return nil
	}
	bizResp, ok := v.(BizResp)
	if ok && bizResp != nil && bizResp.GetRet() == nil {
		val := reflect.ValueOf(v).Elem()
		ret := val.FieldByName("Ret")
		if ret.CanSet() {
			ret.Set(reflect.ValueOf(&v1.BaseResponse{Code: 0, Message: "", Reason: ""}))
		}
	}
	codec, _ := httptransport.CodecForRequest(r, "Accept")
	data, err := codec.Marshal(v)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	return err
}

func errorEncoder(w http.ResponseWriter, r *http.Request, err error) {
	se := errors.FromError(err)
	codec, _ := httptransport.CodecForRequest(r, "Accept")
	bizCode := v1.ErrorReason_value[se.Reason]
	if bizCode == 0 {
		bizCode = se.GetCode()
	}
	rsp := v1.Response{
		Ret: &v1.BaseResponse{
			Code:    bizCode,
			Reason:  se.Reason,
			Message: se.Message,
		},
	}
	body, err := codec.Marshal(&rsp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(se.Code))
	_, _ = w.Write(body)
}
