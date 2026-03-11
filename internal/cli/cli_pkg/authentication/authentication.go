package authentication

import (
	"fmt"
	"log"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/messageUtils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/api"
	"github.com/0xveya/gns3util/pkg/api/endpoints"
	"github.com/0xveya/gns3util/pkg/api/schemas"
	"github.com/0xveya/gns3util/pkg/utils/nwutils"
)

func TryKeys(kf *pathutils.KeyFileV2, cfg *config.GlobalOptions) ([]byte, error) {
	for i := range kf.StandaloneGNS3 {
		entry := &kf.StandaloneGNS3[i]
		if nwutils.NormalizeURL(cfg.Server) == nwutils.NormalizeURL(entry.URL) {
			result, success := tryKey(entry.AccessToken, cfg)
			if success {
				return result, nil
			}
		}
	}

	for i := range kf.Clusters {
		cluster := &kf.Clusters[i]

		for j := range cluster.GNS3Servers {
			entry := &cluster.GNS3Servers[j]

			if nwutils.NormalizeURL(cfg.Server) == nwutils.NormalizeURL(entry.URL) {
				result, success := tryKey(entry.AccessToken, cfg)
				if success {
					return result, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no working API-Key found for the server %s. Please use the %s command to authenticate", messageUtils.Bold(cfg.Server), messageUtils.Bold("auth login"))
}

func tryKey(token string, cfg *config.GlobalOptions) ([]byte, bool) {
	settings := api.NewSettings(
		api.WithBaseURL(cfg.Server),
		api.WithVerify(!cfg.Insecure),
		api.WithToken(token),
	)

	ep := endpoints.GetEndpoints{}
	client := api.NewGNS3Client(settings)
	reqOpts := api.NewRequestOptions(settings).
		WithURL(ep.Me()).
		WithMethod(api.GET)

	body, resp, err := client.Do(reqOpts)
	if err != nil {
		log.Fatalf("API error: %v", err)
	}
	defer func() {
		if resp != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode == 200 {
		return body, true
	}
	return body, false
}

func SaveAuthData(cfg *config.GlobalOptions, token schemas.Token, username string) error {
	keyFileLocation, err := pathutils.ResolveKeyFilePath(cfg.KeyFile)
	if err != nil {
		return err
	}

	kf, err := pathutils.LoadGNS3KeysFile(keyFileLocation)
	if err != nil {
		return err
	}

	newEntry := pathutils.GNS3ServerEntry{
		URL:         cfg.Server,
		User:        username,
		AccessToken: *token.AccessToken,
		TokenType:   *token.TokenType,
	}

	found := false
	for i, entry := range kf.StandaloneGNS3 {
		if nwutils.NormalizeURL(entry.URL) == nwutils.NormalizeURL(cfg.Server) {
			kf.StandaloneGNS3[i] = newEntry
			found = true
			break
		}
	}
	if !found {
		kf.StandaloneGNS3 = append(kf.StandaloneGNS3, newEntry)
	}

	return pathutils.SaveKeysFile(keyFileLocation, kf)
}

func GetKeyForServer(cfg *config.GlobalOptions) (string, error) {
	keyFileLocation, err := pathutils.ResolveKeyFilePath(cfg.KeyFile)
	if err != nil {
		return "", err
	}

	kf, err := pathutils.LoadGNS3KeysFile(keyFileLocation)
	if err != nil {
		return "", err
	}

	for _, entry := range kf.StandaloneGNS3 {
		if nwutils.NormalizeURL(entry.URL) == nwutils.NormalizeURL(cfg.Server) {
			return entry.AccessToken, nil
		}
	}

	for i := range kf.Clusters {
		cluster := &kf.Clusters[i]

		for j := range cluster.GNS3Servers {
			entry := &cluster.GNS3Servers[j]

			if nwutils.NormalizeURL(entry.URL) == nwutils.NormalizeURL(cfg.Server) {
				return entry.AccessToken, nil
			}
		}
	}
	return "", fmt.Errorf("could not find a matching access token for the server %s, please use the %s command to login", cfg.Server, messageUtils.Bold("auth login"))
}
