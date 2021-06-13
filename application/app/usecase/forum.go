package usecase

import (
	forum "github.com/moroz-matros/technopark_db/application/app"
	"github.com/moroz-matros/technopark_db/application/app/models"
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
	nickname, flag, err := f.repo.CheckUser(forum.User)
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
	count, err := f.repo.CountPosts(frm.Slug)
	if err != nil {
		return frm, err
	}
	frm.Posts = count
	cnt, err := f.repo.CountThreads(frm.Slug)
	if err != nil {
		return frm, err
	}
	frm.Threads = cnt

	return frm, nil
}

func (f ForumUC) GetPostDetails(id int64, params string) (models.PostFull, *models.CustomError) {
	answer := models.PostFull{}

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
		answer.Author = &user
	}
	if strings.Contains(params, "forum") {
		frm, err := f.repo.GetForumByPost(id)
		if err != nil {
			return models.PostFull{}, err
		}
		frm.Posts, err = f.repo.CountPosts(frm.Slug)
		frm.Threads, err = f.repo.CountThreads(frm.Slug)
		answer.Forum = &frm
	}
	if strings.Contains(params, "thread") {
		thread, err := f.repo.GetThreadByPost(id)
		if err != nil {
			return models.PostFull{}, err
		}
		votes, err := f.repo.GetVotes(thread.Id)

		if err != nil {
			return models.PostFull{}, err
		}
		thread.Votes = votes
		answer.Thread = &thread
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


	threads, err := f.repo.GetForumThreads(slug, limit, since, desc)
	if err != nil {
		return models.Threads{}, err
	}
	for _, elem := range threads {
		elem.Votes , _ = f.repo.GetVotes(elem.Id)
	}
	return threads, nil
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
	nickname, flag, err := f.repo.CheckUser(thread.Author)
	if err != nil {
		return thread, err
	}
	if !flag {
		return thread, &models.CustomError{
			Code:    404,
			Message: "author or thread not found",
		}
	}
	var forumSlug string
	forumSlug, flag, err = f.repo.CheckForumBySlug(thread.Forum)
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
	thread.Forum = forumSlug

	if thread.Slug != "" {
		thread2, e := f.repo.GetThread(thread.Slug)
		if e == nil {
			return thread2, &models.CustomError{
				Code:    409,
				Message: "thread already exists",
			}
		}
		if e.Code != 404 {
			return models.Thread{}, e
		}
	}


	thread, err = f.repo.CreateThread(thread, thread.Created)
	if err != nil {
		return thread, err
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
	thread, err := f.GetThreadBySlugOrId(slugOrId)
	if err != nil {
		return posts, err
	}

	posts, err = f.repo.AddPosts(posts, thread.Id, thread.Forum)
	if err != nil {
		return models.Posts{}, err
	}

	return posts, err
}

func (f ForumUC) GetThreadBySlugOrId(slugOrId string) (models.Thread, *models.CustomError) {
	t, err := f.repo.GetThreadBySlugOrId(slugOrId)
	if err != nil {
		return models.Thread{}, err
	}
	return t, nil
}

func (f ForumUC) UpdateThread(thread models.ThreadUpdate, slugOrId string) (models.Thread, *models.CustomError) {
	return f.repo.UpdateThread(thread, slugOrId)
}

func (f ForumUC) AddVote(vote models.Vote, slugOrId string) (models.Thread, *models.CustomError) {
	thread, err := f.repo.GetThreadBySlugOrId(slugOrId)
	if err != nil {
		return models.Thread{}, err
	}

	name, voice, flag, err := f.repo.CheckVote(vote.Nickname, thread.Id)
	if err != nil {
		return models.Thread{}, err
	}
	if flag {
		vote.Nickname = name
		err = f.repo.UpdateVote(vote, thread.Id)
		if err != nil {
			return models.Thread{}, err
		}
	} else {
		_, flag, err = f.repo.CheckUser(vote.Nickname)
		if err != nil {
			return models.Thread{}, err
		}
		if !flag {
			return models.Thread{}, &models.CustomError{
				Code:    404,
				Message: "user does not exist",
			}
		}
		err = f.repo.AddVote(vote, thread.Id)
		if err != nil {
			return models.Thread{}, err
		}
	}

	thread.Votes = thread.Votes - voice + vote.Voice

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


func (f ForumUC) GetPosts(slugOrId string, limit int, since int64, desc bool, sort string) (models.Posts, *models.CustomError) {
	var posts models.Posts
	_, err := f.repo.GetThreadBySlugOrId(slugOrId)
	if err != nil {
		return models.Posts{}, err
	}
	if sort == "flat" {
		return f.repo.GetPostsFlat(slugOrId, limit, since, desc)
	}
	if sort == "tree" {
		return f.repo.GetPostsTree(slugOrId, limit, since, desc)
	}
	if sort == "parent_tree" {
		parents, err := f.repo.GetPostsParent(slugOrId, limit, since, desc)
		if err != nil {
			return posts, err
		}
		return f.repo.GetPostsChild(slugOrId, desc, parents)

	}
	return posts, nil
}


