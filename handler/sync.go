package handler

import (
	"context"
	pb "github.com/liuyp5181/gateway/api"
	"github.com/liuyp5181/gateway/data"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"sync"
)

type ServiceImpl struct {
	pb.UnimplementedGreeterServer
}

var external sync.Map

func (s *ServiceImpl) SyncExternal(ctx context.Context, req *pb.SyncExternalReq) (resp *pb.SyncExternalRes, err error) {
	if len(req.ServiceName) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "service_name is nil")
	}
	list, err := data.QueryExternalList(req.ServiceName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query err: %v", err)
	}
	var m = make(map[string]*data.External)
	for _, v := range list {
		if v.Status == 0 {
			continue
		}
		mds := strings.Split(v.Method, ",")
		for _, md := range mds {
			m[md] = v
		}
	}
	external.Store(req.ServiceName, m)
	resp = &pb.SyncExternalRes{}
	return
}

func (s *ServiceImpl) SyncUserPower(ctx context.Context, req *pb.SyncUserPowerReq) (resp *pb.SyncUserPowerRes, err error) {
	return nil, err
}
