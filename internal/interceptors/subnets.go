package interceptors

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TrustedSubnetInterceptor проверяет, что переданный в метадате запроса x-real-ip
// IP-адрес клиента входит в доверенную подсеть
func TrustedSubnetInterceptor(trustedSubnet string) func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		clientIPSlice, ok := md["x-real-ip"]

		if info.FullMethod == "/shortener.URLShortenerService/GetStats" {
			if !ok || len(clientIPSlice) == 0 {
				return nil, status.Errorf(codes.PermissionDenied, "client ip is not provided")
			}
			clientIP := clientIPSlice[0]

			if trustedSubnet == "" || clientIP == "" {
				return nil, status.Errorf(codes.PermissionDenied, "trusted subnet or client ip are empty")
			}

			_, cidr, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				return nil, status.Errorf(codes.PermissionDenied, "Cannot parse trusted subnet")
			}

			if !cidr.Contains(net.ParseIP(clientIP)) {
				return nil, status.Errorf(codes.PermissionDenied, "Permission Denied")
			}
		}

		return handler(ctx, req)
	}
}
