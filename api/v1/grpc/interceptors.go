package grpc

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// StatusInterceptor is used in order to check, depending on the service's status if the call should be continued or not
func StatusInterceptor(srv *PushService) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {

		// if a request tries to access any other api call rather than the Status call
		// while the service's status is not ok, block the request
		if info.FullMethod != "/PushService/Status" && srv.status != "ok" {
			return nil, status.Error(codes.Internal, ServiceUnavailable)
		}

		return handler(ctx, req)
	}
}

// AuthInterceptor provides ACL based access to the service using certificate DNs
func AuthInterceptor(acl []string, tlsEnabled bool) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {

		// if tls is not enabled skip the authorisation process
		if !tlsEnabled {
			return handler(ctx, req)
		}

		p, ok := peer.FromContext(ctx)
		if ok {
			if p != nil {
				if p.AuthInfo != nil {
					if p.AuthInfo.AuthType() == "tls" {
						tls := p.AuthInfo.(credentials.TLSInfo)
						if len(tls.State.PeerCertificates) > 0 {
							for _, c := range acl {
								if c == tls.State.PeerCertificates[0].Subject.CommonName {
									return handler(ctx, req)
								}
							}
							log.WithFields(
								log.Fields{
									"type":  "error_log",
									"acl":   acl,
									"error": fmt.Sprintf("Provided certificate's cn: %v didn't match any ACL entry", tls.State.PeerCertificates[0].Subject.ToRDNSequence().String()),
								},
							).Error("unauthorised access to the service")
						} else {
							log.WithFields(
								log.Fields{
									"type":  "error_log",
									"error": "No certificate provided",
								},
							).Error("")
						}
					} else {
						log.WithFields(
							log.Fields{
								"type":  "error_log",
								"error": fmt.Sprintf("Peer information AuthInfo is of type %v instead of tls", p.AuthInfo.AuthType()),
							},
						).Error("")
					}
				} else {
					log.WithFields(
						log.Fields{
							"type":  "error_log",
							"error": "Peer information found in the context contains no AuthInfo",
						},
					).Error("")
				}
			} else {
				log.WithFields(
					log.Fields{
						"type":  "error_log",
						"error": "Peer information found in the context is nil",
					},
				).Error("")
			}
		} else {
			log.WithFields(
				log.Fields{
					"type":  "error_log",
					"error": "No peer information found in the context",
				},
			).Error("")
		}
		return nil, status.Error(codes.Unauthenticated, "UNAUTHORISED")
	}
}
