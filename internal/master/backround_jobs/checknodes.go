package backroundjobs

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/0xveya/gns3util/pkg/state"
	"github.com/0xveya/gns3util/pkg/state/pb"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type NodesCheckJob struct {
	Store    *state.StateManager
	Interval time.Duration
	Nodes    []Node
	Logger   *slog.Logger
	Ca       []byte
}

type Node struct {
	ID             string
	HealthCheckURL string
	// TODO: this should ofc support new networking when i add it
}

func (j *NodesCheckJob) Run(ctx context.Context, certPEM []byte) {
	j.Ca = certPEM

	ticker := time.NewTicker(j.Interval)
	defer ticker.Stop()
	tracer := otel.Tracer("backgroundjobs")

	for {
		select {
		case <-ctx.Done():
			j.Logger.Info("nodes check job stopped")
			return
		case <-ticker.C:
			runCtx, span := tracer.Start(ctx, "NodesHealthCheck.Iteration")

			err := j.executeIteration(runCtx)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			span.End()
		}
	}
}

func (j *NodesCheckJob) executeIteration(ctx context.Context) error {
	j.addNodes(ctx)

	for _, node := range j.Nodes {
		select {
		case <-ctx.Done():
			j.Logger.InfoContext(ctx, "nodes check job stopped during node iteration")
			return nil
		default:
		}

		healthy := j.checkNode(ctx, node)

		if err := j.Store.UpdateNodeHealth(ctx, node.ID, healthy); err != nil {
			j.Logger.ErrorContext(ctx, "failed to update node health status",
				"node_id", node.ID,
				"error", err,
			)
		}

		j.Logger.InfoContext(ctx, "checked node health", "node_id", node.ID, "healthy", healthy)
	}
	return nil
}

func (j *NodesCheckJob) checkNode(ctx context.Context, node Node) bool {
	ctx, span := otel.Tracer("backgroundjobs").Start(ctx, "checkNode")
	defer span.End()
	span.SetAttributes(attribute.String("node.url", node.HealthCheckURL))
	caCertPool, certErr := x509.SystemCertPool()
	if certErr != nil || caCertPool == nil {
		caCertPool = x509.NewCertPool()
	}
	caCertPool.AppendCertsFromPEM(j.Ca)

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		node.HealthCheckURL,
		http.NoBody,
	)
	if err != nil {
		j.Logger.ErrorContext(ctx, "failed to create health check request", "node_id", node.ID, "error", err)
		return false
	}

	req.Header.Set("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		j.Logger.ErrorContext(ctx, "failed to perform health check request", "node_id", node.ID, "error", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		j.Logger.WarnContext(ctx, "health check failed", "node_id", node.ID, "status_code", resp.StatusCode)
		return false
	}

	return true
}

func (j *NodesCheckJob) addNodes(ctx context.Context) {
	ctx, span := otel.Tracer("backgroundjobs").Start(ctx, "addNodes")
	defer span.End()
	nodes, getNodesErr := j.Store.GetNodes(ctx)
	if getNodesErr != nil {
		j.Logger.ErrorContext(ctx, "failed to get nodes from state manager",
			"error", getNodesErr,
		)
		return
	}

	j.Nodes = j.Nodes[:0]
	for _, node := range nodes {
		j.Nodes = append(j.Nodes, Node{
			ID:             node.Id,
			HealthCheckURL: buildHealthURL(node),
		})
	}
}

func buildHealthURL(node *pb.Node) string {
	return fmt.Sprintf("https://%s:%d/healthz", node.IpAddress, node.ApiPort)
}
