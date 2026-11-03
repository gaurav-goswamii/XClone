package repository

import (
	"backend/models"
	"database/sql"
	"errors"
	"strings"
	"time"
	"github.com/lib/pq"
)

type PostgresUserRepository struct {
	db *sql.DB
}	

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) GetByID(id string) (*models.User, error) {
	query := `
		SELECT id, username, name, email, password_hash, birth_date, avatar_url, bio, is_verified, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var user models.User
	err := r.db.QueryRow(query, strings.TrimSpace(id)).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.BirthDate,
		&user.AvatarURL,
		&user.Bio,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) GetByEmail(email string) (*models.User, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	query := `
		SELECT id, username, name, email, password_hash, birth_date, avatar_url, bio, is_verified, created_at, updated_at
		FROM users
		WHERE LOWER(email) = $1
	`
	var user models.User
	err := r.db.QueryRow(query, normalized).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.BirthDate,
		&user.AvatarURL,
		&user.Bio,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) GetByUsername(username string) (*models.User, error) {
	normalized := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(username), "@"))
	query := `
		SELECT id, username, name, email, password_hash, birth_date, avatar_url, bio, is_verified, created_at, updated_at
		FROM users
		WHERE LOWER(username) = $1
	`
	var user models.User
	err := r.db.QueryRow(query, normalized).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.BirthDate,
		&user.AvatarURL,
		&user.Bio,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) Create(user *models.User) error {
	emailNorm := strings.ToLower(strings.TrimSpace(user.Email))
	handleNorm := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(user.Username), "@"))

	now := time.Now()
	createdAt := user.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	updatedAt := now

	query := `
		INSERT INTO users (id, username, name, email, password_hash, birth_date, avatar_url, bio, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(query,
		user.ID,
		handleNorm,
		strings.TrimSpace(user.Name),
		emailNorm,
		user.PasswordHash,
		user.BirthDate,
		user.AvatarURL,
		user.Bio,
		user.IsVerified,
		createdAt,
		updatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return ErrUserAlreadyExists
			}
		}
		return err
	}

	user.Username = handleNorm
	user.Email = emailNorm
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	return nil
}

func (r *PostgresUserRepository) UpdateProfile(id string, name string, avatarURL string, bio string) error {
	query := `
		UPDATE users
		SET name = $1, avatar_url = $2, bio = $3, updated_at = $4
		WHERE id = $5
	`
	res, err := r.db.Exec(query, strings.TrimSpace(name), strings.TrimSpace(avatarURL), strings.TrimSpace(bio), time.Now(), strings.TrimSpace(id))
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

type PostgresPostRepository struct {
	db *sql.DB
}

func NewPostgresPostRepository(db *sql.DB) *PostgresPostRepository {
	return &PostgresPostRepository{db: db}
}

func (r *PostgresPostRepository) Create(post *models.Post, media []models.PostMedia) error {
	now := time.Now()
	if post.CreatedAt.IsZero() {
		post.CreatedAt = now
	}
	post.UpdatedAt = now

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	postQuery := `
		INSERT INTO posts (id, user_id, content, likes, retweets, replies, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	if _, err := tx.Exec(postQuery,
		post.ID,
		post.UserID,
		post.Content,
		post.Likes,
		post.Retweets,
		post.Replies,
		post.CreatedAt,
		post.UpdatedAt,
	); err != nil {
		return err
	}

	mediaQuery := `
		INSERT INTO post_media (post_id, ordinal, url, public_id, media_type, width, height)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	for i := range media {
		mediaType := strings.TrimSpace(media[i].MediaType)
		if mediaType == "" {
			mediaType = "image"
		}
		if _, err := tx.Exec(mediaQuery,
			post.ID,
			media[i].Ordinal,
			media[i].URL,
			media[i].PublicID,
			mediaType,
			media[i].Width,
			media[i].Height,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *PostgresPostRepository) mediaByPost(postIDs []string) (map[string][]models.PostMedia, error) {
	grouped := make(map[string][]models.PostMedia, len(postIDs))
	if len(postIDs) == 0 {
		return grouped, nil
	}

	query := `
		SELECT post_id, ordinal, url, public_id, media_type, width, height
		FROM post_media
		WHERE post_id = ANY($1)
		ORDER BY post_id, ordinal
	`
	rows, err := r.db.Query(query, pq.Array(postIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var postID string
		var media models.PostMedia
		if err := rows.Scan(
			&postID,
			&media.Ordinal,
			&media.URL,
			&media.PublicID,
			&media.MediaType,
			&media.Width,
			&media.Height,
		); err != nil {
			return nil, err
		}
		grouped[postID] = append(grouped[postID], media)
	}
	return grouped, rows.Err()
}

func (r *PostgresPostRepository) attachMedia(posts []*models.PostResponse) error {
	if len(posts) == 0 {
		return nil
	}

	postIDs := make([]string, 0, len(posts))
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
	}

	grouped, err := r.mediaByPost(postIDs)
	if err != nil {
		return err
	}

	for _, post := range posts {
		if media := grouped[post.ID]; media != nil {
			post.Media = media
		} else {
			post.Media = []models.PostMedia{}
		}
	}
	return nil
}

func (r *PostgresPostRepository) GetByID(id string) (*models.PostResponse, error) {
	query := `
		SELECT p.id, p.user_id, u.name, u.username, COALESCE(u.avatar_url, ''),
		       p.content, p.likes, p.retweets, p.replies,
		       p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = $1
	`
	var post models.PostResponse
	err := r.db.QueryRow(query, strings.TrimSpace(id)).Scan(
		&post.ID,
		&post.UserID,
		&post.AuthorName,
		&post.AuthorUsername,
		&post.AuthorAvatar,
		&post.Content,
		&post.Likes,
		&post.Retweets,
		&post.Replies,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	if err := r.attachMedia([]*models.PostResponse{&post}); err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostgresPostRepository) ListFeed(limit, offset int) ([]*models.PostResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT p.id, p.user_id, u.name, u.username, COALESCE(u.avatar_url, ''),
		       p.content, p.likes, p.retweets, p.replies,
		       p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]*models.PostResponse, 0)
	for rows.Next() {
		var post models.PostResponse
		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.AuthorName,
			&post.AuthorUsername,
			&post.AuthorAvatar,
			&post.Content,
			&post.Likes,
			&post.Retweets,
			&post.Replies,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	if err := r.attachMedia(posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostgresPostRepository) ListByUser(username string, limit, offset int) ([]*models.PostResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT p.id, p.user_id, u.name, u.username, COALESCE(u.avatar_url, ''),
		       p.content, p.likes, p.retweets, p.replies,
		       p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE LOWER(u.username) = LOWER($1)
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(query, strings.TrimSpace(username), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]*models.PostResponse, 0)
	for rows.Next() {
		var post models.PostResponse
		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.AuthorName,
			&post.AuthorUsername,
			&post.AuthorAvatar,
			&post.Content,
			&post.Likes,
			&post.Retweets,
			&post.Replies,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	if err := r.attachMedia(posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostgresPostRepository) Update(
	id, userID, content string,
	media []models.PostMedia,
) error {
	trimmedID := strings.TrimSpace(id)

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		UPDATE posts
		SET content = $1, updated_at = $2
		WHERE id = $3 AND user_id = $4
	`

	res, err := tx.Exec(
		query,
		content,
		time.Now(),
		trimmedID,
		strings.TrimSpace(userID),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPostNotFound
	}

	if _, err := tx.Exec(`DELETE FROM post_media WHERE post_id = $1`, trimmedID); err != nil {
		return err
	}

	mediaQuery := `
		INSERT INTO post_media
			(post_id, ordinal, url, public_id, media_type, width, height)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	for i := range media {
		mediaType := strings.TrimSpace(media[i].MediaType)
		if mediaType == "" {
			mediaType = "image"
		}

		if _, err := tx.Exec(
			mediaQuery,
			trimmedID,
			media[i].Ordinal,
			media[i].URL,
			media[i].PublicID,
			mediaType,
			media[i].Width,
			media[i].Height,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresPostRepository) GetMediaByPostID(postID string,)([]models.PostMedia, error){
	query := `SELECT ordinal, url, public_id, media_type, width, height
		FROM post_media
		WHERE post_id = $1
		ORDER BY ordinal`
	rows, err := r.db.Query(query, strings.TrimSpace(postID))
	if err!= nil {
		return nil, err
	}	
	defer rows.Close()
	var media []models.PostMedia

	for rows.Next() {
		var item models.PostMedia

		if err := rows.Scan(
			&item.Ordinal,
			&item.URL,
		&item.PublicID,
		&item.MediaType,
		&item.Width,
		&item.Height,
		); err != nil {
			return nil, err
		}
		media = append(media, item)
	}
	 if err := rows.Err(); err != nil {
		return nil, err
	 }
	 return media, nil
}

func (r *PostgresPostRepository) Delete(id, userID string) error {
	query := `
		DELETE FROM posts
		WHERE id = $1 AND user_id = $2
	`
	res, err := r.db.Exec(query, strings.TrimSpace(id), strings.TrimSpace(userID))
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPostNotFound
	}
	return nil
}

type PostgresChatRepository struct {
	db *sql.DB
}

func NewPostgresChatRepository(db *sql.DB) *PostgresChatRepository {
	return &PostgresChatRepository{db: db}
}

func (r *PostgresChatRepository) Create(chat *models.Chat) error {
	query := `INSERT INTO chat (id, user1_id, user2_id)
		VALUES ($1, $2, $3)
		RETURNING created_at`

	return r.db.QueryRow(
		query,
		chat.ID,
		chat.User1ID,
		chat.User2ID,
	).Scan(&chat.CreatedAt)
}

func (r *PostgresChatRepository) GetByUsers(
	user1ID string,
	user2ID string,
)(*models.Chat, error){

	user1ID, user2ID =canonicalPair(user1ID, user2ID)
	 
	query:=`
	SELECT id, user1_id, user2_id, created_at
	FROM chat
	WHERE user1_id = $1
	AND user2_id = $2`

	var chat models.Chat

	err := r.db.QueryRow(
		query,
		user1ID,
		user2ID,
	).Scan(
		&chat.ID,
		&chat.User1ID,
		&chat.User2ID,
		&chat.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChatNotFound
		}
		return nil, err
	}

	return &chat, nil
}

func(r* PostgresChatRepository) GetByID(
	id string,
)(*models.Chat, error){
	query:=`
	SELECT id, user1_id, user2_id, created_at
        FROM chat
        WHERE id = $1 `

		var chat models.Chat

		err := r.db.QueryRow(query, id).Scan(
			&chat.ID,
			&chat.User1ID,
			&chat.User2ID,
			&chat.CreatedAt,
		)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrChatNotFound
			}
			return nil, err
		}
		return &chat, nil
	}

type PostgresMessageRepository struct {
	db *sql.DB
}

func NewPostgresMessageRepository(db *sql.DB) *PostgresMessageRepository {
	return &PostgresMessageRepository{db: db}
}

func (r *PostgresMessageRepository) Create(message *models.Message) error {

 query:=`INSERT INTO message (id, chat_id, sender_id, content)
 VALUES($1, $2, $3, $4)
 RETURNING seq, created_at`

 	err := r.db.QueryRow(
    query,
    message.ID,
    message.ChatID,
    message.SenderID,
    message.Content,
).Scan(&message.Seq, &message.CreatedAt)

return err
}

func (r *PostgresMessageRepository) ListByChat(chatID string) ([]models.Message, error) {
	query := `
	SELECT id, chat_id, sender_id, content, seq, created_at
        FROM message
        WHERE chat_id = $1
        ORDER BY seq ASC`

		rows, err := r.db.Query(query, chatID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		messages := []models.Message{}

		for rows.Next() {
			var message models.Message

			err := rows.Scan(
				&message.ID,
				&message.ChatID,
				&message.SenderID,
				&message.Content,
				&message.Seq,
				&message.CreatedAt,
			)
			if err != nil {
				return nil, err
			}
			messages = append(messages, message)
		}

		if err:= rows.Err(); err != nil {
			return nil, err
		}
		return messages, nil
}

func canonicalPair(a, b string) (string, string) {
    if a < b {
        return a, b
    }
    return b, a
}