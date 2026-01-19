package audit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"process-api/pkg/logging"
	"time"

	api "process-api/pkg/audit/api/v1"

	"braces.dev/errtrace"
	"github.com/tidwall/sjson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuditServiceClient struct {
	client api.GwAuditServiceClient
}

type AuditRequest struct {
	Id             string
	AuditRequestId *string
	PageNumber     int64 `json:"pageNumber"`
	PageSize       int64 `json:"pageSize"`
	From           time.Time
	To             time.Time
}

func New(ctx context.Context) (*AuditServiceClient, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		logging.Logger.With("error", err).Error("UserCacheDir failed")
		return nil, err
	}
	cacheDir = filepath.Join(cacheDir, "netxd", "gateway")
	if err = os.MkdirAll(cacheDir, 0o755); err != nil {
		logging.Logger.With("error", err).Error("failed to create app directory")
		return nil, err
	}

	conn, err := grpc.NewClient(fmt.Sprint("unix://", filepath.Join(cacheDir, "grpc.sock")), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logging.Logger.With("error", err).Error("grpc connect failed")
		return nil, err
	}
	return &AuditServiceClient{client: api.NewGwAuditServiceClient(conn)}, nil
}

func (c *AuditServiceClient) GetAudits(ctx context.Context, req AuditRequest) ([]byte, error) {
	var gwr api.GetAuditsRequest
	if !req.From.IsZero() {
		gwr.From = timestamppb.New(req.From)
	}
	if !req.To.IsZero() {
		gwr.To = timestamppb.New(req.To)
	}
	if req.AuditRequestId != nil {
		gwr.Filter = &api.GetAuditsRequest_RequestId{RequestId: *req.AuditRequestId}
	}
	gwr.PageNumber, gwr.PageSize = req.PageNumber, req.PageSize

	resp, err := c.client.GetAudits(ctx, &gwr, grpc.WaitForReady(false), grpc.MaxCallRecvMsgSize(16*1024*1024))
	if err != nil {
		return nil, errtrace.Wrap(err)
	}

	bs, err := protojson.Marshal(resp)
	if err == nil && req.Id != "" {
		bs, err = sjson.SetBytes(bs, "id", req.Id)
	}
	if err != nil {
		return nil, errtrace.Wrap(err)
	}
	return bs, nil
}
