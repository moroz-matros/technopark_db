package usecase

import (
	forum "github.com/moroz-matros/technopark_db/application/app"
	"github.com/moroz-matros/technopark_db/application/app/models"
	"net/http"
	"strings"
	"time"
)

type ForumUC struct {
	repo forum.Repository
}

func NewForum(repoDatabase forum.Repository) forum.Usecase {
	return &ForumUC{repo: repoDatabase}
}

func (f ForumUC) CreateForum(forum *models.Forum) (models.Forum, *models.CustomError) {
	nickname, _, flag, err := f.repo.CheckUser(forum.User)
	if err != nil {
		return models.Forum{}, err
	}
	if !flag {
		return models.Forum{}, &models.CustomError{
			Code:    404,
			Message: "user does not exist",
		}
	}

	_, flag, err = f.repo.CheckForumBySlug(forum.Slug)
	if err != nil {
		return models.Forum{}, err
	}
	if flag {
		frm, err := f.repo.GetForum(forum.Slug)
		if err != nil {
			return models.Forum{}, err
		}
		return frm, &models.CustomError{
			Code:    409,
			Message: "forum already exists",
		}
	}
	forum.User = nickname

	err = f.repo.CreateForum(forum)
	if err != nil {
		return models.Forum{}, err
	}

	return *forum, nil
}

func (f ForumUC) GetForum(slug string) (models.Forum, *models.CustomError) {
	frm, err := f.repo.GetForum(slug)
	if err != nil {
		return frm, err
	}
	count, err := f.repo.CountPosts(frm.Id)
	if err != nil {
		return frm, err
	}
	frm.Posts = int64(count)
	cnt, err := f.repo.CountThreads(frm.Id)
	if err != nil {
		return frm, err
	}
	frm.Threads = cnt

	return frm, nil
}

func (f ForumUC) GetThreadDetails(id int64, params string) (models.PostFull, *models.CustomError) {
	var answer models.PostFull

	post, err := f.repo.GetPostById(id)
	if err != nil {
		return models.PostFull{}, err
	}
	answer.Post = post
	if strings.Contains(params, "user") {
		user, err := f.repo.GetUserByPost(id)
		if err != nil {
			return models.PostFull{}, err
		}
		answer.Author = user
	}
	if strings.Contains(params, "forum") {
		frm, err := f.repo.GetForumByPost(id)
		if err != nil {
			return models.PostFull{}, err
		}
		answer.Forum = frm
	}
	if strings.Contains(params, "thread") {
		thread, err := f.repo.GetThreadByPost(id)
		if err != nil {
			return models.PostFull{}, err
		}
		answer.Thread = thread
	}

	return answer, nil
}

func (f ForumUC) GetForumThreads(slug string, limit int, since time.Time, desc bool) (models.Threads, *models.CustomError) {
	_, flag, err := f.repo.CheckForumBySlug(slug)
	if err != nil {
		return models.Threads{}, err
	}
	if !flag {
		return models.Threads{}, &models.CustomError{
			Code:    404,
			Message: "forum does not exist",
		}
	}

	return f.repo.GetForumThreads(slug, limit, since, desc)
}

func (f ForumUC) GetForumUsers(slug string, limit int, since string, desc bool) (models.Users, *models.CustomError) {
	_, flag, err := f.repo.CheckForumBySlug(slug)
	if err != nil {
		return models.Users{}, err
	}
	if !flag {
		return models.Users{}, &models.CustomError{
			Code:    404,
			Message: "forum does not exist",
		}
	}

	return f.repo.GetForumUsers(slug, limit, since, desc)
}

func (f ForumUC) CreateThread(thread models.Thread) (models.Thread, *models.CustomError) {
	nickname, _, flag, err := f.repo.CheckUser(thread.Author)
	if err != nil {
		return thread, err
	}
	if !flag {
		return thread, &models.CustomError{
			Code:    404,
			Message: "author or thread not found",
		}
	}
	_, flag, err = f.repo.CheckForumBySlug(thread.Forum)
	if err != nil {
		return thread, err
	}
	if !flag {
		return thread, &models.CustomError{
			Code:    404,
			Message: "author or thread not found",
		}
	}
	thread.Author = nickname
	thread, err = f.repo.CreateThread(thread, thread.Created)
	if err != nil {
		if err.Code == 409 {
			thread, err2 := f.repo.GetThread(thread.Slug)
			if err2 != nil {
				return thread, err2
			}
			return thread, err
		}
	}

	return thread, nil
}

func (f ForumUC) UpdatePost(postId int64, newPost models.PostUpdate) (models.Post, *models.CustomError) {
	return f.repo.UpdatePost(postId, newPost)
}

func (f ForumUC) ClearDatabase() *models.CustomError {
	return f.repo.ClearAll()
}

func (f ForumUC) GetServiceInfo() (models.Status, *models.CustomError) {
	return f.repo.GetServiceInfo()
}

func (f ForumUC) AddPosts(posts models.Posts, slugOrId string) (models.Posts, *models.CustomError) {
	//lastId, err := f.repo.GetLastPostInThread(slugOrId)
	//if err != nil {
	//	return models.Posts{}, err
	//}
	/*id, e := strconv.Atoi(slugOrId)
	idd := int32(id)
	log.Println(slugOrId)
	if e != nil {
		thread, err := f.repo.GetThread(slugOrId)
		log.Println(thread.Id, err)
		if err != nil {
			return posts, err
		}
		idd = thread.Id
	}

	 */
	thread, err := f.GetThreadBySlugOrId(slugOrId)
	if err != nil {
		return posts, err
	}
	for _, elem := range posts {
		elem.Thread = thread.Id
	}

	posts, err = f.repo.AddPosts(posts, slugOrId)
	if err != nil {
		return models.Posts{}, err
	}

	return posts, err
}

func (f ForumUC) GetThreadBySlugOrId(slugOrId string) (models.Thread, *models.CustomError) {
	return f.repo.GetThreadBySlugOrId(slugOrId)
}

func (f ForumUC) UpdateThread(thread models.ThreadUpdate, slugOrId string) (models.Thread, *models.CustomError) {
	return f.repo.UpdateThread(thread, slugOrId)
}

func (f ForumUC) AddVote(vote models.Vote, slugOrId string) (models.Thread, *models.CustomError) {
	thread, err := f.repo.GetThreadBySlugOrId(slugOrId)
	if err != nil {
		return models.Thread{}, err
	}

	err = f.repo.AddVote(vote, thread.Id)
	if err != nil && err.Code == http.StatusConflict {
		err = f.repo.UpdateVote(vote, thread.Id)
		if err != nil {
			return models.Thread{}, err
		}
	}
	if err != nil {
		return models.Thread{}, err
	}
	voices, err := f.repo.GetVotes(thread.Id)
	if err != nil {
		return models.Thread{}, err
	}
	thread.Votes = voices

	return thread, nil
}

func (f ForumUC) AddUser(user models.User, nickname string) (*models.User, *models.Users, *models.CustomError) {
	user.Nickname = nickname
	users, err := f.repo.ReturnUsers(user.Nickname, user.Email)
	if users != nil {
		return nil, &users, err
	}
	if err != nil {
		return nil, nil, err
	}
	u, err := f.repo.AddUser(user)
	//if err != nil && err.Code == http.StatusConflict {
	//	users, e := f.repo.ReturnUsers(user.Nickname, user.Email)
	//	if e != nil {
	//		return nil, nil, e
	//	}
	//	return nil, &users, err
	//}
	if err != nil {
		return nil, nil, err
	}
	return &u, nil, nil
}

func (f ForumUC) GetUser(nickname string) (models.User, *models.CustomError) {
	return f.repo.GetUser(nickname)
}

func (f ForumUC) UpdateUser(nickname string, update models.UserUpdate) (models.User, *models.CustomError) {
	user, err := f.repo.GetUser(nickname)
	flag := true
	if err != nil {
		return models.User{}, err
	}
	count := 0
	if len(update.Fullname) == 0 {
		update.Fullname = user.Fullname
		count +=1
	}
	if len(update.Email) == 0 {
		flag = false
		update.Email = user.Email
		count +=1
	}
	if len(update.About) == 0 {
		update.About = user.About
		count +=1
	}
	if count == 3 {
		return user, nil
	}
	if flag {
		users, err := f.repo.ReturnUsers("", update.Email)
		if users != nil {
			return models.User{}, err
		}
		if err != nil {
			return models.User{}, err
		}
	}

	err = f.repo.UpdateUser(nickname, update)
	if err != nil {
		return models.User{}, err
	}
	user = models.User{
		Nickname: nickname,
		Fullname: update.Fullname,
		About:    update.About,
		Email:    update.Email,
	}

	return user, nil
}

/*
func (f ForumUC) GetPosts(slugOrId string, limit int, since int64, desc bool, sort string) (models.Posts, *models.CustomError) {
	var posts models.Posts
	if sort == "flat" {
		return f.repo.GetPostsFlat(slugOrId, limit, since, desc)
	}
	if sort == "tree" {
		if desc == false {
			post, err := f.repo.GetPostWithSince(since)
			if err != nil {
				return posts, err
			}
			posts = append(posts, post)
			for i := 1; i < limit; i++ {
				post, err = f.repo.GetPostByParentId(post.Id)
				if err != nil && err.Code == 404 {

				}
			}

		}
	}
	if sort == "parent_tree" {

	}
}

*/
