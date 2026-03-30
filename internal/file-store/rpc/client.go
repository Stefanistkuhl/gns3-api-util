package rpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"path/filepath"

	sharedpb "github.com/0xveya/gns3util/internal/shared/pb"
	pb "github.com/0xveya/gns3util/internal/shared/pb/master"
	"storj.io/drpc/drpcconn"
)

type MasterSyncClient struct {
	client pb.DRPCMasterSyncServiceClient
	conn   *drpcconn.Conn
}

func NewMasterSyncClient(
	ctx context.Context,
	addr string,
	tlsDir string,
) (*MasterSyncClient, error) {
	certFile := filepath.Join(filepath.Clean(tlsDir), "node.crt")
	keyFile := filepath.Join(filepath.Clean(tlsDir), "node.key")
	caFile := filepath.Join(filepath.Clean(tlsDir), "ca.crt")

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load client cert/key: %w", err)
	}

	caData, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("load ca cert: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caData) {
		return nil, fmt.Errorf("append ca cert failed")
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS13,
		ServerName:   host,
	}

	var dialer net.Dialer
	rawConn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial master drpc: %w", err)
	}

	tlsConn := tls.Client(rawConn, tlsConfig)
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		_ = rawConn.Close()
		return nil, fmt.Errorf("tls handshake: %w", err)
	}

	conn := drpcconn.New(tlsConn)

	return &MasterSyncClient{
		client: pb.NewDRPCMasterSyncServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *MasterSyncClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *MasterSyncClient) CheckPermission(
	ctx context.Context,
	userID, jti string, roles []string, action sharedpb.Action, resource sharedpb.Resource,
) (bool, error) {
	resp, err := c.client.CheckPermission(ctx, &pb.CheckPermissionRequest{
		UserId:           userID,
		Jti:              jti,
		RoleNames:        roles,
		RequiredResource: resource,
		RequiredAction:   action,
	})
	if err != nil {
		return false, err
	}
	return resp.Allowed, nil
}

func (c *MasterSyncClient) RegisterJob(
	ctx context.Context,
	jobName, nodeID, interval, description string,
) error {
	_, err := c.client.RegisterJob(ctx, &pb.RegisterJobRequest{
		JobName:     jobName,
		NodeId:      nodeID,
		Interval:    interval,
		Description: description,
	})
	return err
}
