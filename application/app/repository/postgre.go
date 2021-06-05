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

func (d Database) CheckParentPost(threadId int32, postId int64) (bool, *models.CustomError) {
	var id int64
	err := d.pool.QueryRow(context.Background(),
		`SELECT id
		FROM posts 
		WHERE id = $1 AND thread = $2`, postId, threadId).Scan(&id)
	if errors.As(err, &sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}
	return true, nil
}

func (d Database) GetPostsParent(slugOrId string, limit int, since int64, desc bool) (models.Posts, *models.CustomError) {
	posts := models.Posts{}
	var order, slug string
	flagIsId := true
	if desc {
		order = "DESC"
	} else {
		order = "ASC"
	}
	var s string
	if order == "DESC" && since != 0 {
		s = " < "
	} else {
		s = " > "
	}
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		flagIsId = false
		slug = slugOrId
	}
	if flagIsId {
		if since == 0 {
			err = pgxscan.Select(context.Background(), d.pool, &posts,
				`SELECT p.id, p.parent, p.author, p.message, 
		p.is_edited, p.forum, p.thread, p.created, p.path
		FROM posts p WHERE p.thread = $1 AND
		 p.parent = '0' 
		ORDER BY p.path ` + order +`, p.created ` + order +`, p.id ` + order +
					` LIMIT $2   `, id, limit)
		} else {
			err = pgxscan.Select(context.Background(), d.pool, &posts,
				`SELECT p.id, p.parent, p.author, p.message, 
		p.is_edited, p.forum, p.thread, p.created, p.path
		FROM posts p WHERE p.thread = $1 AND
		p.parent = '0' and p.path `+ s + `(SELECT path FROM posts where id = $2) AND (SELECT POSITION(path IN (SELECT path FROM posts where id = $2))) = '0'
		ORDER BY p.path ` + order +`, p.created ` + order +`, p.id ` + order +
					` LIMIT $3   `, id, since, limit)
		}

	} else {
		if since == 0 {
			err = pgxscan.Select(context.Background(), d.pool, &posts,
				`SELECT p.id, p.parent, p.author, p.message, 
		p.is_edited, p.forum, p.thread, p.created, p.path
		FROM posts p 
		JOIN threads t ON t.slug = $1 AND p.thread = t.id
		WHERE p.parent = '0'
		ORDER BY p.path ` + order +`, p.created ` + order +`, p.id ASC`  +
					` LIMIT $2   `, slug, limit)
		} else {
			err = pgxscan.Select(context.Background(), d.pool, &posts,
				`SELECT p.id, p.parent, p.author, p.message, 
		p.is_edited, p.forum, p.thread, p.created, p.path
		FROM posts p 
		JOIN threads t ON t.slug = $1 AND p.thread = t.id
		WHERE p.parent = '0' and p.path `+ s + `(SELECT path FROM posts where id = $2) AND (SELECT POSITION(path IN (SELECT path FROM posts where id = $2))) = '0'
		ORDER BY p.path ` + order +`, p.created ` + order +`, p.id ASC`  +
					` LIMIT $3   `, slug, since, limit)
		}

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

func (d Database) GetPostsChild(slugOrId string, desc bool, posts models.Posts) (models.Posts, *models.CustomError) {
	var order string
	var children models.Posts
	if desc {
		order = "DESC"
	} else {
		order = "ASC"
	}

	newPosts := models.Posts{}
	for _, elem := range posts {
		path := elem.Path + "."
		err := pgxscan.Select(context.Background(), d.pool, &children,
			`SELECT p.id, p.parent, p.author, p.message, 
		p.is_edited, p.forum, p.thread, p.created
		FROM posts p WHERE p.thread = $1 AND 
		LEFT(p.path, length($2)) = $2 
		ORDER BY p.path ASC, p.created ` + order +`, p.id ASC`,
		elem.Thread, path)
		if err != nil {
			return models.Posts{}, &models.CustomError{
				Code:    500,
				Message: err.Error(),
			}
		}
		newPosts = append(newPosts, elem)
		for _, child := range children {
			newPosts = append(newPosts, child)
		}
		children = children[:0]

	}

	return newPosts, nil
}

func (d Database) GetPostsFlat(slugOrId string, limit int, since int64, desc bool) (models.Posts, *models.CustomError) {
	posts := models.Posts{}
	var order, slug string
	flagIsId := true
	if desc {
		order = "DESC"
	} else {
		order = "ASC"
	}
	var s string
	if order == "DESC" {
		s = " < "
	} else {
		s = " > "
	}
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		flagIsId = false
		slug = slugOrId
	}
	if flagIsId {
		if since == 0 {
			err = pgxscan.Select(context.Background(), d.pool, &posts,
				`SELECT p.id, p.parent, p.author, p.message, 
		p.is_edited, p.forum, p.thread, p.created
		FROM posts p WHERE p.thread = $1 
		ORDER BY created ` + order +`, p.id ` + order +
					` LIMIT $2   `, id, limit)
		} else {
			err = pgxscan.Select(context.Background(), d.pool, &posts,
				`SELECT p.id, p.parent, p.author, p.message, 
		p.is_edited, p.forum, p.thread, p.created
		FROM posts p WHERE p.thread = $1 AND
		p.id `+s+` $2
		ORDER BY created ` + order +`, p.id ` + order +
					` LIMIT $3   `, id, since, limit)
		}

	} else {
		if since == 0 {
			err = pgxscan.Select(context.Background(), d.pool, &posts,
				`SELECT p.id, p.parent, p.author, p.message, 
		p.is_edited, p.forum, p.thread, p.created
		FROM posts p 
		JOIN threads t ON t.slug = $1 AND p.thread = t.id
		ORDER BY created ` + order +` , p.id ` + order +
					` LIMIT $2   `, slug, limit)
		} else {
			err = pgxscan.Select(context.Background(), d.pool, &posts,
				`SELECT p.id, p.parent, p.author, p.message, 
		p.is_edited, p.forum, p.thread, p.created
		FROM posts p 
		JOIN threads t ON t.slug = $1 AND p.thread = t.id
		WHERE p.id `+s+` $2
		ORDER BY created ` + order +` , p.id ` + order +
					` LIMIT $3   `, slug, since, limit)
		}

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
	posts := models.Posts{}
	var order, slug string
	flagIsId := true
	if desc {
		order = "DESC"
	} else {
		order = "ASC"
	}
	var s string
	if order == "DESC" && since != 0 {
		s = " < "
	} else {
		s = " > "
	}
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		flagIsId = false
		slug = slugOrId
	}
	if flagIsId {
		if since == 0 {
			err = pgxscan.Select(context.Background(), d.pool, &posts,
				`SELECT p.id, p.parent, p.author, p.message, 
		p.is_edited, p.forum, p.thread, p.created
		FROM posts p WHERE p.thread = $1 
		ORDER BY p.path ` + order +`, p.created ` + order +`, p.id ASC` +
					` LIMIT $2   `, id, limit)
		} else {
			err = pgxscan.Select(context.Background(), d.pool, &posts,
				`SELECT p.id, p.parent, p.author, p.message, 
		p.is_edited, p.forum, p.thread, p.created
		FROM posts p WHERE p.thread = $1 AND
		path `+ s + `(SELECT path FROM posts where id = $2)
		ORDER BY p.path ` + order +`, p.created ` + order +`, p.id ASC` +
					` LIMIT $3   `, id, since, limit)
		}

	} else {
		if since == 0{
			err = pgxscan.Select(context.Background(), d.pool, &posts,
				`SELECT p.id, p.parent, p.author, p.message, 
		p.is_edited, p.forum, p.thread, p.created
		FROM posts p 
		JOIN threads t ON t.slug = $1 AND p.thread = t.id
		ORDER BY p.path ` + order +`, p.created ` + order +`, p.id ASC` +
					` LIMIT $2   `, slug, limit)
		} else {
			err = pgxscan.Select(context.Background(), d.pool, &posts,
				`SELECT p.id, p.parent, p.author, p.message, 
		p.is_edited, p.forum, p.thread, p.created
		FROM posts p 
		JOIN threads t ON t.slug = $1 AND p.thread = t.id
		WHERE path `+ s + `(SELECT path FROM posts where id = $2)
		ORDER BY p.path ` + order +`, p.created ` + order +`, p.id ASC` +
					` LIMIT $3   `, slug, since, limit)
		}
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

func NewDatabase(conn *pgxpool.Pool) forum.Repository {
	return &Database{pool: conn}
}

func (d Database) CreateForum(forum *models.Forum) *models.CustomError {
	_, err := d.pool.Exec(context.Background(),
		`INSERT INTO forums (title, u, slug) VALUES ($1, $2, $3)`,
		forum.Title, forum.User, forum.Slug)

	if err != nil {
		return &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return nil
}

func (d Database) CheckUser(nickname string) (string, bool, *models.CustomError) {
	err := d.pool.
		QueryRow(context.Background(),
			`SELECT nickname FROM users WHERE nickname = $1`, nickname).Scan(&nickname)

	if errors.As(err, &pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}
	return nickname, true, nil
}

func (d Database) CheckForumBySlug(slug string) (string, bool, *models.CustomError) {
	var name string
	err := d.pool.
		QueryRow(context.Background(),
			`SELECT slug FROM forums WHERE slug = $1`, slug).Scan(&name)

	if errors.As(err, &pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}
	return name, true, nil
}

func (d Database) GetForum(slug string) (models.Forum, *models.CustomError) {
	var f models.Forum
	err := d.pool.QueryRow(context.Background(),
		`SELECT id, title, u, slug
		FROM forums WHERE slug = $1`, slug).Scan(
			&f.Id, &f.Title, &f.User, &f.Slug)
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
	err := d.pool.QueryRow(context.Background(),
		`SELECT nickname, fullname, about, email
		FROM users, posts
		WHERE posts.id = $1 AND nickname = author`, postId).Scan(
			&user.Nickname, &user.Fullname, &user.About, &user.Email)

	if err != nil {
		return models.User{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return user, nil
}

func (d Database) GetForumThreads(slug string, limit int, since time.Time, desc bool) (models.Threads, *models.CustomError) {
	threads := models.Threads{}
	var order string
	if desc {
		order = "DESC"
	} else {
		order = "ASC"
	}
	var s string
	if order == "DESC" && since.String() != "0001-01-01 00:00:00 +0000 UTC" {
		s = " <= "
	} else {
		s = " >= "
	}
	err := pgxscan.Select(context.Background(), d.pool, &threads,
		`SELECT t.id, t.title, t.author, t.forum, t.message, t.slug, t.created
		FROM threads t
		WHERE t.forum = $1 AND t.created `+s+` $2
		ORDER BY t.created ` + order +
			` LIMIT $3`, slug, since, limit)

	if err != nil {
		return models.Threads{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return threads, nil
}

func (d Database) GetForumUsers(slug string, limit int, since string, desc bool) (models.Users, *models.CustomError) {
	users := models.Users{}
	var order string
	if desc {
		order = "DESC"
	} else {
		order = "ASC"
	}
	var s string
	if order == "DESC" {
		s = " < "
	} else {
		s = " > "
	}
	var err error
	if since != "" {
		err = pgxscan.Select(context.Background(), d.pool, &users,
			`SELECT nickname, fullname, about, email from(
					SELECT nickname, fullname, about, email
					FROM users
					JOIN threads t ON t.author = nickname AND t.forum = $1
					WHERE CAST(lower(nickname) as bytea) ` + s + ` CAST(lower($2) as bytea)
					UNION
					SELECT nickname, fullname, about, email
					FROM users
					JOIN posts p ON p.author = nickname AND p.forum = $1
					WHERE CAST(lower(nickname) as bytea) ` + s + ` CAST(lower($2) as bytea)) help
				ORDER BY CAST(lower(nickname) as bytea) `+order+` 
				LIMIT $3;`, slug, since, limit)
	} else {
		err = pgxscan.Select(context.Background(), d.pool, &users,
			`SELECT nickname, fullname, about, email from(
					SELECT nickname, fullname, about, email
					FROM users
					JOIN threads t ON t.author = nickname AND t.forum = $1
					UNION
					SELECT nickname, fullname, about, email
					FROM users
					JOIN posts p ON p.author = nickname AND p.forum = $1) help
				ORDER BY CAST(lower(nickname) as bytea) `+order+` 
				LIMIT $2;`, slug, limit)
	}

	if err != nil {
		return users, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return users, nil
}

func (d Database) GetThread(slug string) (models.Thread, *models.CustomError) {
	var t models.Thread
	err := d.pool.QueryRow(context.Background(),
		`SELECT id, title, author, forum, message, slug, created
		FROM threads WHERE slug = $1`, slug).Scan(
			&t.Id, &t.Title, &t.Author, &t.Forum, &t.Message, &t.Slug, &t.Created)
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

func (d Database) CreateThread(thread models.Thread, data time.Time) (models.Thread, *models.CustomError) {
	err := d.pool.QueryRow(context.Background(),
		`INSERT INTO threads 
		(title, slug, message, author, forum, created) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		thread.Title, thread.Slug, thread.Message,
		thread.Author, thread.Forum, data).Scan(&thread.Id)
	if errors.As(err, &sql.ErrNoRows) {
		return thread, &models.CustomError{
			Code:    409,
			Message: "thread already exists",
		}
	}
	if err != nil {
		return models.Thread{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return thread, nil
}

func (d Database) CountThreads(frm string) (int32, *models.CustomError) {
	var count int32
	err := d.pool.
		QueryRow(context.Background(),
			`SELECT COUNT (*) FROM threads WHERE forum = $1`, frm).Scan(&count)
	if err != nil && !errors.As(err, &sql.ErrNoRows) {
		return 0, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}

	return count, nil
}

func (d Database) CountPosts(frm string) (int64, *models.CustomError) {
	var count int64
	err := d.pool.
		QueryRow(context.Background(),
			`SELECT COUNT (*) FROM posts WHERE forum = $1`, frm).Scan(&count)
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
	isEdited := true
	if newPost.Message == "" {
		isEdited = false
	}
	post, err := d.GetPostById(postId)
	if err != nil {
		return models.Post{}, err
	}
	if post.Message == newPost.Message {
		return post, err
	}
	e := d.pool.
		QueryRow(context.Background(),
		`UPDATE posts SET message = COALESCE(NULLIF($1, ''), message), is_edited = $2
		WHERE id = $3
		RETURNING *`, newPost.Message, isEdited, postId).Scan(&editedPost.Id,
		&editedPost.Parent, &editedPost.Author,
		&editedPost.Message, &editedPost.IsEdited, &editedPost.Forum,
		&editedPost.Thread, &editedPost.Created, &editedPost.Path)
	if errors.As(e, &sql.ErrNoRows) {
		return models.Post{}, &models.CustomError{
			Code:    404,
			Message: "post does not exist",
		}
	}

	if e != nil {
		return models.Post{}, &models.CustomError{
			Code:    500,
			Message: e.Error(),
		}
	}

	return editedPost, nil
}


func (d Database) GetPostById(id int64) (models.Post, *models.CustomError) {
	var post models.Post
	err := d.pool.QueryRow(context.Background(),
		`SELECT id, parent, author, message, is_edited, forum, thread, created
		FROM posts 
		WHERE id = $1`, id).Scan(&post.Id, &post.Parent,
			&post.Author, &post.Message, &post.IsEdited,
			&post.Forum, &post.Thread, &post.Created)
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
	err := d.pool.QueryRow(context.Background(),
		`SELECT t.id, t.title, t.author, t.forum, t.message, t.slug, t.created
		FROM threads t, posts p
		WHERE p.id = $1 AND t.id = p.thread`, postId).Scan(
		&t.Id, &t.Title, &t.Author, &t.Forum, &t.Message, &t.Slug, &t.Created)
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
	err := d.pool.QueryRow(context.Background(),
		`SELECT f.id, f.title, f.u, f.slug
		FROM forums f, posts p
		WHERE p.id = $1 AND p.forum = f.slug`, postId).Scan(
			&f.Id, &f.Title, &f.User, &f.Slug)

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
		`TRUNCATE users, forums, threads, posts, votes`)
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
		(SELECT COUNT (*) from users) as "user",
		(SELECT COUNT (*) from forums) as forum,
		(SELECT COUNT (*) from threads) as thread,
		(SELECT COUNT (*) from posts) as post`).Scan(&answer.User,
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

func (d Database) AddPosts(posts models.Posts, threadId int32, form string) (models.Posts, *models.CustomError) {
	//var id int64
	for _, elem := range posts {
		if elem.Parent != 0 {
			flag, err := d.CheckParentPost(threadId, elem.Parent)
			if err != nil {
				return models.Posts{}, err
			}
			if !flag {
				return models.Posts{}, &models.CustomError{
					Code:    409,
					Message: "parent in another place",
				}
			}
		}
		elem.Thread = threadId
		
		err := d.pool.QueryRow(context.Background(),
			`INSERT INTO posts
			(parent, author, message, is_edited, thread, created, forum) 
			VALUES ($1, $2, $3, $4, $5, $6, 
			(select t.forum from threads t
			where t.id = $5)) RETURNING id, forum`,
			elem.Parent, elem.Author, elem.Message, elem.IsEdited,
			elem.Thread, elem.Created).Scan(&elem.Id, &elem.Forum)
		if errors.As(err, &sql.ErrNoRows) {
			return posts, &models.CustomError{
				Code:    404,
				Message: "user does not exist",
			}
		}
		if err != nil {
			return models.Posts{}, &models.CustomError{
				Code:    500,
				Message: err.Error(),
			}
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
				`UPDATE threads SET title=COALESCE(NULLIF($1, ''), title), message=COALESCE(NULLIF($2, ''), message) 
		WHERE id = $3 
		RETURNING *`, thread.Title, thread.Message, id).Scan(&editedThread.Id,
			&editedThread.Title, &editedThread.Slug,
			&editedThread.Message,
			&editedThread.Author, &editedThread.Forum,
			 &editedThread.Created)
		if errors.As(err, &sql.ErrNoRows) {
			return models.Thread{}, &models.CustomError{
				Code:    404,
				Message: "thread with this id does not exist",
			}
		}
	} else {
		err = d.pool.
			QueryRow(context.Background(),
				`UPDATE threads SET title=COALESCE(NULLIF($1, ''), title), message=COALESCE(NULLIF($2, ''), message) 
		WHERE slug = $3 
		RETURNING *`, thread.Title, thread.Message, slug).Scan(&editedThread.Id,
			&editedThread.Title, &editedThread.Slug,
			&editedThread.Message,
			&editedThread.Author, &editedThread.Forum,
			&editedThread.Created)
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
		err = d.pool.QueryRow(context.Background(),
			`SELECT id, title, author, forum, message, slug, created
		FROM threads WHERE id = $1`, id).Scan(&t.Id,
			&t.Title, &t.Author, &t.Forum, &t.Message, &t.Slug, &t.Created)
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
			(thread_id, u, voice)
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
		SET voice = $1 WHERE thread_id = $2 AND u = $3`,
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
			`select coalesce((SELECT SUM(CASE WHEN voice = '1' THEN 1 else -1 END)
    			from votes
				WHERE thread_id = $1), 0);`, id).Scan(&count)
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
		`SELECT DISTINCT nickname, fullname, about, email 
		FROM users WHERE nickname = $1 or email = $2`, nickname, email)
	if len(users) == 0 {
		return nil, nil
	}
	if err != nil {
		return models.Users{}, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}
	return users, &models.CustomError{
		Code:    409,
		Message: "user with this nickname or login does exist",
	}
}

func (d Database) GetUser(nickname string) (models.User, *models.CustomError) {
	var user models.User
	err := d.pool.QueryRow(context.Background(),
		`SELECT nickname, fullname, about, email 
		FROM users WHERE nickname = $1`,
		nickname).Scan(
			&user.Nickname, &user.Fullname, &user.About, &user.Email)
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
	//TODO: лишний поход

	_, err := d.pool.Exec(context.Background(),
		`UPDATE users 
		set fullname = COALESCE(NULLIF($1, ''), fullname), about = COALESCE(NULLIF($2, ''), about), email = COALESCE(NULLIF($3, ''), email)
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

func (d Database) CheckVote(nickname string, threadId int32) (string, bool, *models.CustomError) {
	var name string
	err := d.pool.
		QueryRow(context.Background(),
			`SELECT u FROM votes WHERE u = $1 AND thread_id = $2`,
			nickname, threadId).Scan(&name)
	if errors.As(err, &pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, &models.CustomError{
			Code:    500,
			Message: err.Error(),
		}
	}
	return nickname, true, nil
}