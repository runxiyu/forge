package main

import (
	"context"
)

func query_list[T any](ctx context.Context, query string, args ...any) ([]T, error) {
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []T
	for rows.Next() {
		var item T
		if err := rows.Scan(&item); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func query_name_desc_list(ctx context.Context, query string, args ...any) ([]name_desc_t, error) {
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []name_desc_t{}
	for rows.Next() {
		var name, description string
		if err := rows.Scan(&name, &description); err != nil {
			return nil, err
		}
		result = append(result, name_desc_t{name, description})
	}
	return result, rows.Err()
}

type name_desc_t struct {
	Name        string
	Description string
}
