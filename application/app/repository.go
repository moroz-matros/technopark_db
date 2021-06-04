package forum

import (
	"github.com/moroz-matros/technopark_db/application/app/models"
	"time"
)

type Repository interface {
	CreateForum(forum *models.Forum) *models.CustomError
	CheckForumBySlug(slug string) (int32, bool, *models.CustomError)
	CheckUser(nickname string) (string, int32, bool, *models.CustomError)
	GetForum(slug string) (models.Forum, *models.CustomError)
	CountPosts(forumId uint64) (uint64, *models.CustomError)
	CountThreads(forumId uint64) (int32, *models.CustomError)
	CreateThread(thread models.Thread, data time.Time) (models.Thread, *models.CustomError)
	GetThread(slug string) (models.Thread, *models.CustomError)
	GetForumUsers(slug string, limit int, since string, desc bool) (models.Users, *models.CustomError)
	GetForumThreads(slug string, limit int, since time.Time, desc bool) (models.Threads, *models.CustomError)
	GetPostById(id int64) (models.Post, *models.CustomError)
	GetUserByPost(postId int64) (models.User, *models.CustomError)
	GetForumByPost(postId int64) (models.Forum, *models.CustomError)
	GetThreadByPost(postId int64) (models.Thread, *models.CustomError)
	UpdatePost(postId int64, newPost models.PostUpdate) (models.Post, *models.CustomError)
	ClearAll() *models.CustomError
	GetServiceInfo() (models.Status, *models.CustomError)
	AddPosts(posts models.Posts, slugOrId string) (models.Posts, *models.CustomError)
	GetLastPostInThread(slugOrId string) (int64, *models.CustomError)
	GetThreadBySlugOrId(slugOrId string) (models.Thread, *models.CustomError)
	UpdateThread(thread models.ThreadUpdate, slugOrId string) (models.Thread, *models.CustomError)
	AddVote(vote models.Vote, id int32) *models.CustomError
	UpdateVote(vote models.Vote, id int32) *models.CustomError
	GetVotes(id int32) (int32, *models.CustomError)
	AddUser(user models.User) (models.User, *models.CustomError)
	ReturnUsers(nickname string, email string) (models.Users, *models.CustomError)
	GetUser(nickname string) (models.User, *models.CustomError)
	UpdateUser(nickname string, update models.UserUpdate) *models.CustomError
	//GetPostsFlat(slugOrId string, limit int, since int64, desc bool) (models.Posts, *models.CustomError)
	//GetPostsTree(slugOrId string, limit int, since int64, desc bool) (models.Posts, *models.CustomError)
	//GetPostsParentTree(slugOrId string, limit int, since int64, desc bool) (models.Posts, *models.CustomError)
}
