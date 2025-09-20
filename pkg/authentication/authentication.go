package authentication

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/stefanistkuhl/gns3util/pkg/api"
	"github.com/stefanistkuhl/gns3util/pkg/api/endpoints"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/pathUtils"
)

func LoadKeys(keyFileLocation string) ([]pathUtils.GNS3Key, error) {
	var filePath string

	if keyFileLocation == "" {
		dir, err := pathUtils.GetGNS3Dir()
		if err != nil {
			return nil, err
		}
		filePath = filepath.Join(dir, "gns3key")
	} else {
		expanded, err := pathUtils.ExpandPath(keyFileLocation)
		if err != nil {
			return nil, err
		}
		filePath = expanded

		if info, err := os.Stat(filePath); err == nil && info.IsDir() {
			filePath = filepath.Join(filePath, "gns3key")
		} else if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}

	keys, err := pathUtils.LoadGNS3KeysFile(filePath)
	return keys, err
}

func TryKeys(keys []pathUtils.GNS3Key, cfg config.GlobalOptions) ([]byte, error) {
	for _, key := range keys {
		if normalizeURL(cfg.Server) == normalizeURL(key.ServerURL) {
			result, success := tryKey(key, cfg)
			if success {
				return result, nil
			}
		}
	}
	return nil, fmt.Errorf("No working API-Key found for the server %s. Please use the %s command to authenticate.", messageUtils.Bold(cfg.Server), messageUtils.Bold("auth login"))
}

func normalizeURL(url string) string {
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}

	if colonIndex := strings.Index(url, ":"); colonIndex != -1 {
		url = url[:colonIndex]
	}

	return url
}

func tryKey(key pathUtils.GNS3Key, cfg config.GlobalOptions) ([]byte, bool) {
	settings := api.NewSettings(
		api.WithBaseURL(cfg.Server),
		api.WithVerify(!cfg.Insecure),
		api.WithToken(key.AccessToken),
	)

	ep := endpoints.GetEndpoints{}

	client := api.NewGNS3Client(settings)
	reqOpts := api.
		NewRequestOptions(settings).
		WithURL(ep.Me()).
		WithMethod(api.GET)

	body, resp, err := client.Do(reqOpts)
	if err != nil {
		log.Fatalf("API error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return body, true
	} else {
		return body, false
	}

}

func SaveAuthData(cfg config.GlobalOptions, token schemas.Token, username string) error {
	var keyFileLocation string
	if cfg.KeyFile != "" {
		k, err := pathUtils.ExpandPath(cfg.KeyFile)
		if err != nil {
			return err
		}
		keyFileLocation = k
	} else {
		k, err := pathUtils.GetGNS3Dir()
		if err != nil {
			return err
		}
		keyFileLocation = filepath.Join(k, "gns3key")
	}

	keys, err := LoadKeys(cfg.KeyFile)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	}

	newKey := pathUtils.GNS3Key{
		ServerURL:   cfg.Server,
		User:        username,
		AccessToken: *token.AccessToken,
		TokenType:   *token.TokenType,
	}

	found := false
	for i, key := range keys {
		if key.ServerURL == newKey.ServerURL {
			keys[i] = newKey
			found = true
			break
		}
	}
	if !found {
		keys = append(keys, newKey)
	}

	f, err := os.OpenFile(keyFileLocation, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open key file %q: %w", keyFileLocation, err)
	}
	defer f.Close()

	for _, key := range keys {
		buff_key, err := json.Marshal(key)
		if err != nil {
			return err
		}
		_, err = f.Write(buff_key)
		if err != nil {
			return fmt.Errorf("failed to write key to file: %w", err)
		}
		_, err = f.WriteString("\n")
		if err != nil {
			return fmt.Errorf("failed to write newline to file: %w", err)
		}
	}

	return nil
}

func GetKeyForServer(cfg config.GlobalOptions) (string, error) {

	var keyFileLocation string
	if cfg.KeyFile != "" {
		k, err := pathUtils.ExpandPath(cfg.KeyFile)
		if err != nil {
			return "", err
		}
		keyFileLocation = k
	} else {
		k, err := pathUtils.GetGNS3Dir()
		if err != nil {
			return "", err
		}
		keyFileLocation = filepath.Join(k, "gns3key")
	}

	keys, err := LoadKeys(keyFileLocation)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	}
	for _, key := range keys {
		if normalizeURL(key.ServerURL) == normalizeURL(cfg.Server) {
			return key.AccessToken, nil
		}
	}
	return "", fmt.Errorf("Could not find find a matching access token for the server %s, please use the %s command to login to the server.", cfg.Server, messageUtils.Bold("auth login"))
}
