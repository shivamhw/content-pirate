package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/iyear/tdl/core/storage"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/key"
	"github.com/iyear/tdl/pkg/tclient"
)

var DataDir = "/home/shivamhw/Code/reddit-pirate/Rdata/"
var (
	defaultBoltPath = filepath.Join(DataDir, "data")
	DefaultBoltStorage = map[string]string{
		kv.DriverTypeKey: kv.DriverBolt.String(),
		"path":           defaultBoltPath,
	}
)

type Store struct {
	Kvd storage.Storage
	Stg kv.Storage
	BasePath string
}


func GetOrCreateStore(ctx context.Context, path string) (*Store, error) {
	return NewStore(ctx, path, false)
}

func NewStore(ctx context.Context, path string, clean bool) (*Store, error) {
	userPath := filepath.Join(DataDir, path) 
	if _, err := os.Stat(userPath); os.IsNotExist(err) {
		slog.Info("Creating new store", "path", userPath)
		os.MkdirAll(userPath, 0755)
	} else if clean {
		slog.Info("Cleaning existing store", "path", userPath)
		os.RemoveAll(userPath)
		os.MkdirAll(userPath, 0755)
	} else {
		slog.Info("Using existing store", "path", userPath)
	}

	DefaultBoltStorage["path"] = userPath
	stg, err := kv.NewWithMap(DefaultBoltStorage)
	if err != nil {
		return nil, err
	}
	
	kvd, err := stg.Open("default")
	if err != nil {
		return nil, err
	}
	// Initialize KV storage
	if err := kvd.Set(ctx, key.App(), []byte(tclient.AppDesktop)); err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	slog.Info("Store created", "path", userPath)
	return &Store{Kvd: kvd, Stg: stg, BasePath: userPath}, nil
}

func (s *Store) Close() error {
	return s.Stg.Close()
}
