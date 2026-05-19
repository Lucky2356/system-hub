package ui

import (
	"slices"

	"github.com/Lucky2356/system-hub/internal/appstate"
	"github.com/Lucky2356/system-hub/internal/config"
)

func isFavoriteService(name string) bool {
	cfg := appstate.GetConfig()
	return slices.Contains(cfg.FavoriteServices, name)
}

func isFavoriteContainer(name string) bool {
	cfg := appstate.GetConfig()
	return slices.Contains(cfg.FavoriteContainers, name)
}

func toggleFavoriteService(name string) error {
	cfg := appstate.GetConfig()

	if slices.Contains(cfg.FavoriteServices, name) {
		cfg.FavoriteServices = removeString(cfg.FavoriteServices, name)
	} else {
		cfg.FavoriteServices = append(cfg.FavoriteServices, name)
	}

	return saveRuntimeConfig(cfg)
}

func toggleFavoriteContainer(name string) error {
	cfg := appstate.GetConfig()

	if slices.Contains(cfg.FavoriteContainers, name) {
		cfg.FavoriteContainers = removeString(cfg.FavoriteContainers, name)
	} else {
		cfg.FavoriteContainers = append(cfg.FavoriteContainers, name)
	}

	return saveRuntimeConfig(cfg)
}

func removeString(items []string, target string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item != target {
			result = append(result, item)
		}
	}
	return result
}

func saveRuntimeConfig(cfg config.Config) error {
	cfg.Normalize()

	if err := config.Save(cfg); err != nil {
		return err
	}

	appstate.SetConfig(cfg)
	return nil
}