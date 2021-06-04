package models

import "time"

type User struct {
	Nickname string `json:"nickname"`
	Fullname string `json:"fullname"`
	About    string `json:"about"`
	Email    string `json:"email"`
}
type Forum struct {
	Id uint64 `json:"id"`
	Title   string `json:"title"`
	User    string `json:"user"`
	Slug    string `json:"slug"`
	Posts   int64  `json:"posts"`
	Threads int32  `json:"threads"`
}

//easyjson:json
type Users []User

type UserUpdate struct {
	Fullname string `json:"fullname"`
	About    string `json:"about"`
	Email    string `json:"email"`
}

type Thread struct {
	Id      int32  `json:"id"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	Forum   string `json:"forum"`
	Message string `json:"message"`
	Votes   int32  `json:"votes"`
	Slug    string `json:"slug"`
	Created time.Time `json:"created"`
}

//easyjson:json
type Threads []Thread

type ThreadUpdate struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type Error struct {
	Message string `json:"message"`
}

type CustomError struct {
	Code    int
	Message string
}

type Post struct {
	Id       int64  `json:"id"`
	Parent   int64  `json:"parent"`
	Author   string `json:"author"`
	Message  string `json:"message"`
	IsEdited bool   `json:"isEdited"`
	Forum    string `json:"forum"`
	Thread   int32  `json:"thread"`
	Created  time.Time `json:"created"`
	Path string `json:"path"`
}

//easyjson:json
type Posts []*Post

type PostUpdate struct {
	Message string `json:"message"`
}

type PostFull struct {
	Post   Post   `json:"post"`
	Author *User   `json:"author,omitempty"`
	Thread *Thread `json:"thread,omitempty"`
	Forum  *Forum   `json:"forum,omitempty"`
}


type Vote struct {
	Nickname string `json:"nickname"`
	Voice    int32  `json:"voice"`
}

type Status struct {
	User int32 `json:"user"`
	Forum int32 `json:"forum"`
	Thread int32 `json:"thread"`
	Post int32 `json:"post"`
}
