package forum

import (
	"github.com/moroz-matros/technopark_db/application/app/models"
	"time"
)

type Usecase interface {
	CreateForum(forum *models.Forum) (models.Forum, *models.CustomError)
	GetForum(slug string) (models.Forum, *models.CustomError)
	CreateThread(thread models.Thread) (models.Thread, *models.CustomError)
	GetForumUsers(slug string, limit int, since string, desc bool) (models.Users, *models.CustomError)
	GetForumThreads(slug string, limit int, since time.Time, desc bool) (models.Threads, *models.CustomError)
	GetPostDetails(id int64, params string) (models.PostFull, *models.CustomError)
	UpdatePost(postId int64, newPost models.PostUpdate) (models.Post, *models.CustomError)
	ClearDatabase() *models.CustomError
	GetServiceInfo() (models.Status, *models.CustomError)
	AddPosts(posts models.Posts, slugOrId string) (models.Posts, *models.CustomError)
	GetThreadBySlugOrId(slugOrId string) (models.Thread, *models.CustomError)
	UpdateThread(thread models.ThreadUpdate, slugOrId string) (models.Thread, *models.CustomError)
	AddVote(vote models.Vote, slugOrId string) (models.Thread, *models.CustomError)
	AddUser(user models.User, nickname string) (*models.User, *models.Users, *models.CustomError)
	GetUser(nickname string) (models.User, *models.CustomError)
	UpdateUser(nickname string, update models.UserUpdate) (models.User, *models.CustomError)
	GetPosts(slugOrId string, limit int, since int64, desc bool, sort string) (models.Posts, *models.CustomError)

}
