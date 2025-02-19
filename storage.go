package main

import (
    "encoding/json"
    "os"
)

func loadFavorites() error {
    file, err := os.ReadFile(favoritesFile)
    if err != nil {
        if os.IsNotExist(err) {
            favorites = Favorites{Countries: []string{}}
            return nil
        }
        return err
    }
    return json.Unmarshal(file, &favorites)
}

func saveFavorites() error {
    data, err := json.Marshal(favorites)
    if err != nil {
        return err
    }
    return os.WriteFile(favoritesFile, data, 0644)
}