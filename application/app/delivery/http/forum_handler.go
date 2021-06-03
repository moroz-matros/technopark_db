package http

import (
	"github.com/gorilla/mux"
	"github.com/labstack/echo"
	"github.com/mailru/easyjson"
	forum "github.com/moroz-matros/technopark_db/application/app"
	"github.com/moroz-matros/technopark_db/application/app/models"
	"github.com/moroz-matros/technopark_db/pkg/constants"
	"net/http"
	"strconv"
	"time"
)

type ForumHandler struct {
	uc forum.Usecase
}

func CreateForumHandler(e *echo.Echo, useCase forum.Usecase) {
	forumHandler := ForumHandler{}

	e.POST("/api/forum/create", forumHandler.CreateForum)
	e.GET("/api/forum/{slug}/details", forumHandler.GetForum)
	e.POST("/api/forum/{slug}/create", forumHandler.CreateThread)
	e.GET("/api/forum/{slug}/users", forumHandler.GetForumUsers)
	e.GET("/api/forum/{slug}/threads", forumHandler.GetForumThreads)
	e.GET("/api/post/{id}/details", forumHandler.GetThreadDetails)
	e.POST("/api/post/{id}/details", forumHandler.ChangeMessage)
	e.POST("/api/service/clear", forumHandler.ClearDatabase)
	e.GET("/api/service/status", forumHandler.GetServiceInfo)
	e.POST("/api/thread/{slug_or_id}/create", forumHandler.CreatePosts)
	e.GET("/api/thread/{slug_or_id}/details", forumHandler.GetThreadInfo)
	e.POST("/api/thread/{slug_or_id}/details", forumHandler.ChangeThreadInfo)
	//e.GET("/api/thread/{slug_or_id}/posts", forumHandler.GetPosts)
	e.POST("/api/thread/{slug_or_id}/vote", forumHandler.MakeVote)
	e.POST("/api/user/{nickname}/create", forumHandler.CreateUser)
	e.GET("/api/user/{nickname}/profile", forumHandler.GetUser)
	e.POST("/api/user/{nickname}/profile", forumHandler.UpdateUser)




}

func (h ForumHandler) CreateForum(c echo.Context) error {
	defer c.Request().Body.Close()

	frm := &models.Forum{}

	err := easyjson.UnmarshalFromReader(c.Request().Body, frm)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	f, e := h.uc.CreateForum(frm)
	if e != nil {
		return c.JSON(e.Code, models.Error{Message: e.Message})
	}

	return c.JSON(http.StatusCreated, f)
}

func (h ForumHandler) GetForum(c echo.Context) error {
	defer c.Request().Body.Close()

	vars := mux.Vars(c.Request())
	slug := vars["slug"]

	frm, err := h.uc.GetForum(slug)
	if err != nil {
		return c.JSON(err.Code, models.Error{Message: err.Message})
	}

	return c.JSON(http.StatusOK, frm)
}

func (h ForumHandler) CreateThread(c echo.Context) error {
	defer c.Request().Body.Close()

	thread := &models.Thread{}

	vars := mux.Vars(c.Request())
	slug := vars["slug"]

	err := easyjson.UnmarshalFromReader(c.Request().Body, thread)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	thread.Forum = slug

	t, e := h.uc.CreateThread(*thread)
	if e != nil {
		if e.Code == 404 {
			return c.JSON(e.Code, e.Message)
		}
		if e.Code == 409 {
			return c.JSON(e.Code, t)
		}
	}
	return c.JSON(201, t)

}

func (h ForumHandler) GetForumUsers(c echo.Context) error {
	defer c.Request().Body.Close()

	vars := mux.Vars(c.Request())
	slug := vars["slug"]
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit == 0 {
		limit = 100
	}
	since := c.QueryParam("since")
	flag := c.QueryParam("desc")
	var desc bool
	if flag == "true" {
		desc = true
	} else {
		desc = false
	}

	users, err := h.uc.GetForumUsers(slug, limit, since, desc)
	if err != nil {
		return c.JSON(err.Code, err.Message)
	}

	return c.JSON(200, users)
}

func (h ForumHandler) GetForumThreads(c echo.Context) error {
	defer c.Request().Body.Close()

	vars := mux.Vars(c.Request())
	slug := vars["slug"]
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit == 0 {
		limit = 100
	}
	since, _ := time.Parse(constants.Time, c.QueryParam("since"))
	flag := c.QueryParam("desc")
	var desc bool
	if flag == "true" {
		desc = true
	} else {
		desc = false
	}

	threads, err := h.uc.GetForumThreads(slug, limit, since, desc)
	if err != nil {
		return c.JSON(err.Code, err.Message)
	}

	return c.JSON(200, threads)
}

func (h ForumHandler) GetThreadDetails(c echo.Context) error {
	defer c.Request().Body.Close()

	vars := mux.Vars(c.Request())
	id, _ := strconv.Atoi(vars["id"])
	params := c.QueryParam("related")
	answer, err := h.uc.GetThreadDetails(int64(id), params)
	if err != nil {
		return c.JSON(err.Code, err.Code)
	}

	return c.JSON(200, answer)
}

func (h ForumHandler) ChangeMessage(c echo.Context) error {
	defer c.Request().Body.Close()

	var newPost models.PostUpdate

	err := easyjson.UnmarshalFromReader(c.Request().Body, &newPost)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	vars := mux.Vars(c.Request())
	id, _ := strconv.Atoi(vars["id"])
	edited, e := h.uc.UpdatePost(int64(id), newPost)
	if e != nil {
		return c.JSON(e.Code, e.Message)
	}

	return c.JSON(200, edited)
}

func (h ForumHandler) ClearDatabase(c echo.Context) error {
	defer c.Request().Body.Close()

	err := h.uc.ClearDatabase()
	if err != nil {
		return c.JSON(err.Code, err.Message)
	}

	return c.JSON(200, "deleted successfully")
}

func (h ForumHandler) GetServiceInfo(c echo.Context) error {
	defer c.Request().Body.Close()

	answer, err := h.uc.GetServiceInfo()
	if err != nil {
		return c.JSON(err.Code, err.Message)
	}

	return c.JSON(200, answer)
}

func (h ForumHandler) CreatePosts(c echo.Context) error {
	defer c.Request().Body.Close()

	var newPosts models.Posts

	err := easyjson.UnmarshalFromReader(c.Request().Body, &newPosts)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	now := time.Now()
	for _, elem := range newPosts {
		elem.Created = now.Format(constants.Time)
	}

	vars := mux.Vars(c.Request())
	slugOrId := vars["slug_or_id"]
	posts, e := h.uc.AddPosts(newPosts, slugOrId)
	if e != nil {
		return c.JSON(e.Code, e.Message)
	}

	//TODO: 409 ???

	return c.JSON(200, posts)
}

func (h ForumHandler) GetThreadInfo(c echo.Context) error {
	defer c.Request().Body.Close()

	vars := mux.Vars(c.Request())
	slugOrId := vars["slug_or_id"]

	thread, err := h.uc.GetThreadBySlugOrId(slugOrId)
	if err != nil {
		return c.JSON(err.Code, err.Message)
	}

	return c.JSON(200, thread)
}

func (h ForumHandler) ChangeThreadInfo(c echo.Context) error {
	defer c.Request().Body.Close()

	var thread models.ThreadUpdate

	err := easyjson.UnmarshalFromReader(c.Request().Body, &thread)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	vars := mux.Vars(c.Request())
	slugOrId := vars["slug_or_id"]
	t, e := h.uc.UpdateThread(thread, slugOrId)
	if e != nil {
		return c.JSON(e.Code, e.Message)
	}

	return c.JSON(200, t)
}

/*
func (h ForumHandler) GetPosts(c echo.Context) error {
	defer c.Request().Body.Close()

	vars := mux.Vars(c.Request())
	slugOrId := vars["slug_or_id"]
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit == 0 {
		limit = 100
	}
	since := c.QueryParam("since")
	flag := c.QueryParam("desc")
	var desc bool
	if flag == "true" {
		desc = true
	} else {
		desc = false
	}
	sort := c.QueryParam("sort")
	if sort == "" {
		sort = "flat"
	}



}

 */

func (h ForumHandler) MakeVote(c echo.Context) error {
	defer c.Request().Body.Close()

	var vote models.Vote
	vars := mux.Vars(c.Request())
	slugOrId := vars["slug_or_id"]

	err := easyjson.UnmarshalFromReader(c.Request().Body, &vote)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	thread, e := h.uc.AddVote(vote, slugOrId)
	if e != nil {
		return c.JSON(e.Code, e.Message)
	}

	return c.JSON(200, thread)
}

func (h ForumHandler) CreateUser(c echo.Context) error {
	defer c.Request().Body.Close()

	vars := mux.Vars(c.Request())
	nickname := vars["nickname"]

	var user models.User
	err := easyjson.UnmarshalFromReader(c.Request().Body, &user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	u, users, e := h.uc.AddUser(user, nickname)
	if users != nil {
		return c.JSON(e.Code, users)
	}
	if e != nil {
		return c.JSON(e.Code, e.Message)
	}

	return c.JSON(201, u)
}

func (h ForumHandler) GetUser(c echo.Context) error {
	defer c.Request().Body.Close()

	vars := mux.Vars(c.Request())
	nickname := vars["nickname"]

	user, err := h.uc.GetUser(nickname)
	if err != nil {
		return c.JSON(err.Code, err.Message)
	}

	return c.JSON(200, user)
}

func (h ForumHandler) UpdateUser(c echo.Context) error {
	defer c.Request().Body.Close()

	vars := mux.Vars(c.Request())
	nickname := vars["nickname"]

	var user models.UserUpdate
	err := easyjson.UnmarshalFromReader(c.Request().Body, &user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	u, e := h.uc.UpdateUser(nickname, user)
	if e != nil {
		return c.JSON(e.Code, e.Message)
	}


	return c.JSON(200, u)
}