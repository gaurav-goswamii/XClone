package repository

import (
	"backend/models"
	"database/sql"
	"errors"
	"strings"
)

type PostgresCommentRepository struct {
	db *sql.DB
}

func NewPostgresCommentRepository(db *sql.DB) *PostgresCommentRepository {
	return &PostgresCommentRepository{db: db}
}

func (r *PostgresCommentRepository) Create(comment *models.Comment) (*models.CommentResponse, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Exec(`
		INSERT INTO comments (id, post_id, user_id, content)
		VALUES ($1, $2, $3, $4)
	`, comment.ID, comment.PostID, comment.UserID, comment.Content)
	if err != nil {
		return nil, err
	}

	result, err := tx.Exec(
		`UPDATE posts SET replies = replies + 1 WHERE id = $1`,
		comment.PostID,
	)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected != 1 {
		return nil, ErrPostNotFound
	}

	response, err := r.getResponseByIDTx(tx, comment.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return response, nil
}

func (r *PostgresCommentRepository) ListByPost(postID string, limit, offset int) ([]*models.CommentResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT c.id, c.post_id, c.user_id, u.name, u.username,
		       COALESCE(u.avatar_url, ''), c.content, c.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = $1
		ORDER BY c.created_at ASC, c.id ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(query, strings.TrimSpace(postID), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]*models.CommentResponse, 0)
	for rows.Next() {
		var comment models.CommentResponse
		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.AuthorName,
			&comment.AuthorUsername,
			&comment.AuthorAvatar,
			&comment.Content,
			&comment.CreatedAt,
		); err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *PostgresCommentRepository) GetByID(id string) (*models.Comment, error) {
	query := `
		SELECT id, post_id, user_id, content, created_at
		FROM comments
		WHERE id = $1
	`
	var comment models.Comment
	err := r.db.QueryRow(query, strings.TrimSpace(id)).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.Content,
		&comment.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}
	return &comment, nil
}

func (r *PostgresCommentRepository) Delete(id, postID, userID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.Exec(
		`DELETE FROM comments WHERE id = $1 AND post_id = $2 AND user_id = $3`,
		strings.TrimSpace(id),
		strings.TrimSpace(postID),
		strings.TrimSpace(userID),
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return ErrCommentNotFound
	}

	result, err = tx.Exec(
		`UPDATE posts SET replies = GREATEST(replies - 1, 0) WHERE id = $1`,
		strings.TrimSpace(postID),
	)
	if err != nil {
		return err
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return ErrPostNotFound
	}

	return tx.Commit()
}

type commentRowScanner interface {
	Scan(dest ...any) error
}

func scanCommentResponse(row commentRowScanner) (*models.CommentResponse, error) {
	var comment models.CommentResponse
	err := row.Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.AuthorName,
		&comment.AuthorUsername,
		&comment.AuthorAvatar,
		&comment.Content,
		&comment.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}
	return &comment, nil
}

const commentResponseQuery = `
	SELECT c.id, c.post_id, c.user_id, u.name, u.username,
	       COALESCE(u.avatar_url, ''), c.content, c.created_at
	FROM comments c
	JOIN users u ON c.user_id = u.id
	WHERE c.id = $1
`

func (r *PostgresCommentRepository) getResponseByIDTx(
	tx *sql.Tx,
	id string,
) (*models.CommentResponse, error) {
	return scanCommentResponse(tx.QueryRow(commentResponseQuery, strings.TrimSpace(id)))
}

func (r *PostgresCommentRepository) getResponseByID(id string) (*models.CommentResponse, error) {
	return scanCommentResponse(r.db.QueryRow(commentResponseQuery, strings.TrimSpace(id)))
}
