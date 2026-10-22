package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type WishlistItem struct {
	MovieID int64     `json:"movieId"`
	Title   string    `json:"title"`
	TitleTh *string   `json:"titleTh"`
	Poster  *string   `json:"poster"`
	Note    *string   `json:"note"`
	AddedAt time.Time `json:"addedAt"`
}

func (s *Store) ListWishlist(ctx context.Context, memberID int64) ([]WishlistItem, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT w.movie_id, m.title, m.title_th,
		        'https://image.tmdb.org/t/p/w500' || m.poster_path,
		        w.note, w.added_at
		   FROM wishlist_items w
		   JOIN movies m ON m.id = w.movie_id
		  WHERE w.member_id = $1
		  ORDER BY w.added_at DESC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []WishlistItem{}
	for rows.Next() {
		var item WishlistItem
		if err := rows.Scan(&item.MovieID, &item.Title, &item.TitleTh,
			&item.Poster, &item.Note, &item.AddedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) PutWishlistItem(ctx context.Context, memberID, movieID int64, note *string) (WishlistItem, error) {
	var item WishlistItem
	err := s.Pool.QueryRow(ctx,
		`WITH saved AS (
		     INSERT INTO wishlist_items (member_id, movie_id, note)
		     VALUES ($1, $2, $3)
		     ON CONFLICT (member_id, movie_id)
		     DO UPDATE SET note = EXCLUDED.note
		     RETURNING movie_id, note, added_at
		 )
		 SELECT saved.movie_id, m.title, m.title_th,
		        'https://image.tmdb.org/t/p/w500' || m.poster_path,
		        saved.note, saved.added_at
		   FROM saved
		   JOIN movies m ON m.id = saved.movie_id`, memberID, movieID, note).
		Scan(&item.MovieID, &item.Title, &item.TitleTh, &item.Poster, &item.Note, &item.AddedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return WishlistItem{}, ErrNoMovie
	}
	return item, err
}

func (s *Store) DeleteWishlistItem(ctx context.Context, memberID, movieID int64) error {
	tag, err := s.Pool.Exec(ctx,
		`DELETE FROM wishlist_items WHERE member_id = $1 AND movie_id = $2`,
		memberID, movieID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
