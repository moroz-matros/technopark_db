package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgxpool"
	forum "github.com/moroz-matros/technopark_db/application/app"
	"github.com/moroz-matros/technopark_db/application/app/models"
	"net/http"
	"strconv"
	"time"
)

type Database struct {
	pool   *pgxpool.Pool
}


/*
func (d Database) GetPostsFlat(slugOrId string, limit int, since int64, desc bool) (models.Posts, *models.CustomError) {
	var posts models.Posts
	var order, slug string
	var flagIsId bool
	if desc {
		order = "DESC"
	} else {
		order = "ASC"
	}
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		flagIsId = false
		slug = slugOrId
	}
	if flagIsId {
		err = pgxscan.Select(context.Background(), d.pool, &posts,
			`SELECT p.id, p.parent_id, p.author, p.message, 
		p.is_edited, p.forum, p.thread, p.created
		FROM posts p WHERE p.thread = $1 AND
		p.id > $2
		ORDER BY created ` + order +
				` LIMIT $3   `, id, since, limit)
	} else {
		err = pgxscan.Select(context.Background(), d.pool, &posts,
			`SELECT p.id, p.parent_id, p.author, p.message, 
		p.is_edited, p.forum, p.thread, p.created
		FROM posts p WHERE p.id > $2
		JOIN threads t ON t.slug = $1 AND p.thread = t.id
		ORDER BY created ` + order +
				` LIMIT $3   `, slug, since, limit)
	}
	if errors.As(err, &sql.ErrNoRows) {
		return models.Posts{}, &models.CustomError{
			Code:    404,
			Message: "thread with this slug or id does not exist",
		}
	}
	if err != nil {
		return models.Posts{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return posts, nil
}

func (d Database) GetPostsTree(slugOrId string, limit int, since int64, desc bool) (models.Posts, *models.CustomError) {
	var posts models.Posts
	var post models.Post
	var order, slug string
	var flagIsId bool
	if desc {
		order = "DESC"
	} else {
		order = "ASC"
	}
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		flagIsId = false
		slug = slugOrId
	}
	if flagIsId {
		err = pgxscan.Select(context.Background(), d.pool, &posts,
			`SELECT p.id, p.parent_id, p.author, p.message, 
		p.is_edited, p.forum, p.thread, p.created
		FROM posts p WHERE p.id > $2
		JOIN threads t ON t.slug = $1 AND p.thread = t.id
		ORDER BY created ` + order +
				` LIMIT $3   `, slug, since, limit)

	}



}

func (d Database) GetPostsParentTree(slugOrId string, limit int, since int64, desc bool) (models.Posts, *models.CustomError) {
	var posts models.Posts
	var order string
	if desc {
		order = "DESC"
	} else {
		order = "ASC"
	}

}


 */
func NewDatabase(conn *pgxpool.Pool) forum.Repository {
	return &Database{pool: conn}
}

func (d Database) CreateForum(forum *models.Forum) *models.CustomError {
	_, err := d.pool.Exec(context.Background(),
		`INSERT INTO forums (title, user, slug) VALUES ($1, $2, $3)`,
		forum.Title, forum.User, forum.Slug)

	if err != nil {
		return &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return nil
}

func (d Database) CheckUser(nickname string) (int32, bool, *models.CustomError) {
	var id int32
	err := d.pool.
		QueryRow(context.Background(),
			`SELECT id FROM users WHERE nickname = $1`, nickname).Scan(&id)

	if errors.As(err, &pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}
	return id, true, nil
}

func (d Database) CheckForumBySlug(slug string) (int32, bool, *models.CustomError) {
	var id int32
	err := d.pool.
		QueryRow(context.Background(),
			`SELECT id FROM forums WHERE slug = $1`, slug).Scan(&id)

	if errors.As(err, &pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}
	return id, true, nil
}

func (d Database) GetForum(slug string) (models.Forum, *models.CustomError) {
	var f models.Forum
	err := pgxscan.Select(context.Background(), d.pool, &f,
		`SELECT id, title, user, slug
		FROM forums WHERE slug = $1`, slug)
	if errors.As(err, &sql.ErrNoRows) {
		return models.Forum{}, &models.CustomError{
			Code:    404,
			Message: "forum with this slug does not exist",
		}
	}
	if err != nil {
		return models.Forum{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}
	return f, nil
}


func (d Database) GetUserByPost(postId int64) (models.User, *models.CustomError){
	var user models.User
	err := pgxscan.Select(context.Background(), d.pool, &user,
		`SELECT nickname, fullname, about, email
		FROM users, posts
		WHERE posts.id = $1 AND nickname = author`, postId)

	if err != nil {
		return models.User{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return user, nil
}

func (d Database) GetForumThreads(slug string, limit int, since time.Time, desc bool) (models.Threads, *models.CustomError) {
	var threads models.Threads
	var order string
	if desc {
		order = "DESC"
	} else {
		order = "ASC"
	}

	err := pgxscan.Select(context.Background(), d.pool, &threads,
		`SELECT title, author, forum, message, slug, created
		FROM threads 
		JOIN forums p ON forum = $1
		WHERE created >= $2
		ORDER BY created ` + order +
			` LIMIT $3   `, slug, since, limit)

	if err != nil {
		return models.Threads{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return threads, nil
}

func (d Database) GetForumUsers(slug string, limit int, since string, desc bool) (models.Users, *models.CustomError) {
	var users models.Users
	var order string
	if desc {
		order = "DESC"
	} else {
		order = "ASC"
	}
	err := pgxscan.Select(context.Background(), d.pool, &users,
		`SELECT nickname, fullname, about, email
		FROM users 
		JOIN posts p ON p.author = nickname AND p.forum = $1
		JOIN threads t ON t.author = nickname AND t.forum = $1
		WHERE nickname > $2
		ORDER BY nickname ` + order +
			` LIMIT $3   `, slug, since, limit)

	if err != nil {
		return models.Users{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return users, nil
}

func (d Database) GetThread(slug string) (models.Thread, *models.CustomError) {
	var t models.Thread
	err := pgxscan.Select(context.Background(), d.pool, &t,
		`SELECT id, title, author, forum, message, slug, created
		FROM threads WHERE slug = $1`, slug)
	if errors.As(err, &sql.ErrNoRows) {
		return models.Thread{}, &models.CustomError{
			Code:    404,
			Message: "thread with this slug does not exist",
		}
	}
	if err != nil {
		return models.Thread{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}
	return t, nil
}

func (d Database) CreateThread(thread models.Thread, data time.Time) *models.CustomError {
	resp, err := d.pool.Exec(context.Background(),
		`INSERT INTO threads 
		(title, slug, message, author, forum created) 
		VALUES ($1, $2, $3, $4, $5, $6)`,
		thread.Title, thread.Slug, thread.Message,
		thread.Author, thread.Author, data)
	if err != nil {
		return &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	if resp.RowsAffected() == 0 {
		return &models.CustomError{
			Code:    409,
			Message: "thread already exists",
		}
	}

	return nil
}

func (d Database) CountThreads(forumId uint64) (int32, *models.CustomError) {
	var count int32
	err := d.pool.
		QueryRow(context.Background(),
			`SELECT COUNT (*) FROM threads WHERE forum_id = $1`, forumId).Scan(&count)
	if err != nil && !errors.As(err, &sql.ErrNoRows) {
		return 0, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return count, nil
}

func (d Database) CountPosts(forumId uint64) (uint64, *models.CustomError) {
	var count uint64
	err := d.pool.
		QueryRow(context.Background(),
			`SELECT COUNT (*) FROM posts WHERE forum_id = $1`, forumId).Scan(&count)
	if err != nil && !errors.As(err, &sql.ErrNoRows) {
		return 0, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return count, nil
}

func (d Database) UpdatePost(postId int64, newPost models.PostUpdate) (models.Post, *models.CustomError) {
	var editedPost models.Post
	err := d.pool.
		QueryRow(context.Background(),
		`UPDATE posts SET message = $1, is_edited = $2
		WHERE id = $3
		RETURNING (id, parent_id, author, message, is_edited,
		forum, thread, created`, newPost.Message, true, postId).Scan(&editedPost.Id,
		&editedPost.Parent, &editedPost.Author,
		&editedPost.Message, &editedPost.IsEdited, &editedPost.Forum,
		&editedPost.Thread, &editedPost.Created)

	if errors.As(err, &sql.ErrNoRows) {
		return models.Post{}, &models.CustomError{
			Code:    404,
			Message: "post does not exist",
		}
	}
	if err != nil {
		return models.Post{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return editedPost, nil
}


func (d Database) GetPostById(id int64) (models.Post, *models.CustomError) {
	var post models.Post
	err := pgxscan.Select(context.Background(), d.pool, &post,
		`SELECT id, parent_id, author, message, is_edited, forum, thread, created
		FROM posts 
		WHERE id = $1`, id)
	if errors.As(err, &sql.ErrNoRows) {
		return models.Post{}, &models.CustomError{
			Code:    404,
			Message: "post does not exist",
		}
	}
	if err != nil {
		return models.Post{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}
	return post, nil
}

func (d Database) GetThreadByPost(postId int64) (models.Thread, *models.CustomError) {
	var t models.Thread
	err := pgxscan.Select(context.Background(), d.pool, &t,
		`SELECT t.id, t.title, t.author, t.forum, t.message, t.slug, t.created
		FROM threads t, posts p
		WHERE p.id = $1 AND t.id = p.thread`, postId)
	if errors.As(err, &sql.ErrNoRows) {
		return models.Thread{}, &models.CustomError{
			Code:    404,
			Message: "thread does not exist",
		}
	}
	if err != nil {
		return models.Thread{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}
	return t, nil
}

func (d Database) GetForumByPost(postId int64) (models.Forum, *models.CustomError) {
	var f models.Forum
	err := pgxscan.Select(context.Background(), d.pool, &f,
		`SELECT f.id, f.title, f.user, f.slug
		FROM forums f, posts p
		WHERE p.id = $1 AND p.forum = f.slug`, postId)

	if err != nil {
		return models.Forum{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}
	return f, nil
}

func (d Database) ClearAll() *models.CustomError {
	_, err := d.pool.Exec(context.Background(),
		`DELETE FROM users`)
	if err != nil {
		return &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return nil
}

func (d Database) GetServiceInfo() (models.Status, *models.CustomError) {
	var answer models.Status
	err := d.pool.
		QueryRow(context.Background(),
			`SELECT
		(SELECT COUNT (*) from users) as users,
		(SELECT COUNT (*) from forums) as forums,
		(SELECT COUNT (*) from threads) as threads,
		(SELECT COUNT (*) from posts) as 
		FROM users, threads, forums, posts`).Scan(&answer.User,
			&answer.Forum, &answer.Thread, &answer.Post)
	if err != nil {
		return models.Status{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return answer, nil
}

func (d Database) GetLastPostInThread(slugOrId string) (int64, *models.CustomError) {
	var slug string
	flagIsId := true
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		flagIsId = false
		slug = slugOrId
	}
	var idLast int64
	if flagIsId {
		err = d.pool.
			QueryRow(context.Background(),
				`SELECT max(id) FROM posts 
			WHERE thread = $1`, id).Scan(&idLast)
	} else {
		err = d.pool.
			QueryRow(context.Background(),
				`SELECT max(id) FROM posts 
			JOIN threads on slug = $1 AND threads.id = posts.thread`, slug).Scan(&idLast)
	}
	if errors.As(err, &sql.ErrNoRows) {
		return 0, &models.CustomError{
			Code:    404,
			Message: "thread does not exist",
		}
	}
	if err != nil {
		return 0, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return idLast, nil
}

func (d Database) AddPosts(posts models.Posts,  slugOrId string, lastId int64) (models.Posts, *models.CustomError) {
	var id int64
	posts[0].Parent = lastId
	for i, elem := range posts {
		err := d.pool.QueryRow(context.Background(),
			`INSERT INTO posts
			(parent_id, author, message, is_edited, forum, thread, created) 
			VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
			elem.Parent, elem.Author, elem.Message, elem.IsEdited,
			elem.Forum, elem.Thread, elem.Created).Scan(&id)
		if err != nil {
			return models.Posts{}, &models.CustomError{
				Code:    500,
				Message: err.Error(),
			}
		}
		if i < len(posts) {
			posts[i+1].Parent = id
		}
	}

	return posts, nil
}

func (d Database) UpdateThread(thread models.ThreadUpdate, slugOrId string) (models.Thread, *models.CustomError) {
	var slug string
	flagIsId := true
	var editedThread models.Thread
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		flagIsId = false
		slug = slugOrId
	}
	if flagIsId {
		err = d.pool.
			QueryRow(context.Background(),
				`UPDATE threads SET title = $1, message = $2
		WHERE id = $3 
		RETURNING (id, title, author, forum, message, slug,
		created`, thread.Title, thread.Message, id).Scan(&editedThread.Id,
			&editedThread.Title, &editedThread.Author,
			&editedThread.Forum, &editedThread.Message,
			&editedThread.Slug, &editedThread.Created)
		if errors.As(err, &sql.ErrNoRows) {
			return models.Thread{}, &models.CustomError{
				Code:    404,
				Message: "thread with this id does not exist",
			}
		}
	} else {
		err = d.pool.
			QueryRow(context.Background(),
				`UPDATE threads SET title = $1, message = $2
		WHERE slug = $3 
		RETURNING (id, title, author, forum, message, slug,
		created`, thread.Title, thread.Message, slug).Scan(&editedThread.Id,
			&editedThread.Title, &editedThread.Author,
			&editedThread.Forum, &editedThread.Message,
			&editedThread.Slug, &editedThread.Created)
		if errors.As(err, &sql.ErrNoRows) {
			return models.Thread{}, &models.CustomError{
				Code:    404,
				Message: "thread with this id does not exist",
			}
		}
	}

	return editedThread, nil
}

func (d Database) GetThreadBySlugOrId(slugOrId string) (models.Thread, *models.CustomError) {
	var slug string
	flagIsId := true
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		flagIsId = false
		slug = slugOrId
	}
	if flagIsId {
		var t models.Thread
		err = pgxscan.Select(context.Background(), d.pool, &t,
			`SELECT id, title, author, forum, message, slug, created
		FROM threads WHERE id = $1`, id)
		if errors.As(err, &sql.ErrNoRows) {
			return models.Thread{}, &models.CustomError{
				Code:    404,
				Message: "thread with this id does not exist",
			}
		}
		return t, nil
	} else {
		return d.GetThread(slug)
	}
}

func (d Database) AddVote(vote models.Vote, id int32) *models.CustomError {
	_, err := d.pool.Exec(context.Background(),
			`INSERT INTO votes
			(thread_id, user, voice)
			VALUES ($1, $2, $3)`, id, vote.Nickname, vote.Voice)

	if errors.As(err, &sql.ErrNoRows) {
		return &models.CustomError{
			Code:    http.StatusConflict,
			Message: "already exists",
		}
	}
	if err != nil {
		return  &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return nil
}

func (d Database) UpdateVote(vote models.Vote, id int32) *models.CustomError {
	_, err := d.pool.Exec(context.Background(),
		`UPDATE votes 
		SET voice = $1 WHERE thread_id = $2 AND user = $3`,
		vote.Voice, id, vote.Nickname)
	if err != nil {
		return &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return nil
}

func (d Database) GetVotes(id int32) (int32, *models.CustomError) {
	var count int32
	err := d.pool.
		QueryRow(context.Background(),
			`SELECT COUNT (*) from votes
			WHERE thread_id = $1 AND voice = $2`, id, +1).Scan(&count)
	if err != nil {
		return 0, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return count, nil
}

func (d Database) AddUser(user models.User) (models.User, *models.CustomError) {
	_, err := d.pool.Exec(context.Background(),
		`INSERT INTO users 
		(nickname, fullname, about, email) 
		VALUES ($1, $2, $3, $4)`,
		user.Nickname, user.Fullname, user.About, user.Email)
	if errors.As(err, &sql.ErrNoRows){
		return user, &models.CustomError{
			Code:    http.StatusConflict,
			Message: "nickname or email already exists",
		}
	}
	if err != nil {
		return models.User{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return user, nil
}

func (d Database) ReturnUsers(nickname string, email string) (models.Users, *models.CustomError) {
	var users models.Users
	err := pgxscan.Select(context.Background(), d.pool, &users,
		`SELECT nickname, fullname, about, email, 
		FROM users WHERE nickname = $1 or email = $2`, nickname, email)
	if err != nil {
		return models.Users{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}
	return users, nil
}

func (d Database) GetUser(nickname string) (models.User, *models.CustomError) {
	var user models.User
	err := pgxscan.Select(context.Background(), d.pool, &user,
		`SELECT nickname, fullname, about, email, 
		FROM users WHERE nickname = $1`, nickname)
	if errors.As(err, &sql.ErrNoRows){
		return user, &models.CustomError{
			Code:    404,
			Message: "user with this nickname does not exist",
		}
	}
	if err != nil {
		return models.User{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}
	return user, nil
}

func (d Database) UpdateUser(nickname string, update models.UserUpdate) *models.CustomError {
	_, err := d.pool.Exec(context.Background(),
		`UPDATE users 
		set fullname = $1, about = $2, email = $3
		WHERE nickname = $4`,
		update.Fullname, update.About, update.Email, nickname)
	if errors.As(err, &sql.ErrNoRows){
		return &models.CustomError{
			Code:    404,
			Message: "user does not exist",
		}
	}
	if err != nil {
		return &models.CustomError{
			Code:    409,
			Message: "Conflict in data",
		}
	}

	return nil
}