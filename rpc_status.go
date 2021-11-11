package main

import (
	"errors"
	"net/http"

	hspb "gitlab.com/jdb.com.la/tapwater/proto/http"
	edpb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var StatusBindingFailure = func() *status.Status {
	s, _ := status.New(codes.InvalidArgument, "Binding JSON body failure. Please pass a valid JSON body.").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "BINDING_FAILURE",
				Domain: "jdbbank.com.la",
			})
	return s
}()

var StatusUnauthenticated = func() *status.Status {
	s, _ := status.New(codes.Unauthenticated, "ID token not valid. Please pass a valid ID token.").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "TOKEN_INVALID",
				Domain: "jdbbank.com.la",
				Metadata: map[string]string{
					"service": "jdbbank.com.la/myaccount",
				}})
	return s
}()

var StatusSessionExpired = func() *status.Status {
	s, _ := status.New(codes.Unauthenticated, "Session has been expired. Please make a new session and try again.").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "SESSION_EXPIRED",
				Domain: "jdbbank.com.la",
			})
	return s
}()

var StatusPermissionDenied = func() *status.Status {
	s, _ := status.New(codes.PermissionDenied, "You does't have sufficient permission to perform action.").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "INSUFFICIENT_PERMISSION",
				Domain: "jdbbank.com.la/myaccount",
			})
	return s
}()

var StatusNoInfo = func() *status.Status {
	s, _ := status.New(codes.NotFound, "info not found.").
		WithDetails(
			&edpb.ResourceInfo{
				ResourceType: "BILL INFO",
				ResourceName: "[google.rpc.Code.NotFound]",
			})
	return s
}()

var StatusDenyCcy = func() *status.Status {
	s, _ := status.New(codes.FailedPrecondition, "info not found.").
		WithDetails(
			&edpb.BadRequest_FieldViolation{
				Field:       "Account['currency']",
				Description: "must be LAK only for this bill",
			})
	return s
}()

func gRPCStatusFromErr(err error) *status.Status {
	switch {
	case err == nil:
		return status.New(codes.OK, "OK")
	case errors.Is(err, ErrPermissionDenied):
		return StatusPermissionDenied
	case errors.Is(err, ErrNoInfo):
		return StatusNoInfo
	case errors.Is(err, ErrNotAllowCcy):
		return StatusDenyCcy
	}

	return status.New(codes.Internal, "Internal Server Error")
}

func httpStatusPbFromRPC(s *status.Status) *hspb.Error {
	return &hspb.Error{
		Error: &hspb.Error_Status{
			Code:    int32(httpStatusFromCode(s.Code())),
			Message: s.Message(),
			Details: s.Proto().Details,
		},
	}
}

// httpStatusFromCode converts a gRPC error code into the corresponding HTTP response status.
// See: https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
func httpStatusFromCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		// Note, this deliberately doesn't translate to the similarly named '412 Precondition Failed' HTTP response status.
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	}

	return http.StatusInternalServerError
}
