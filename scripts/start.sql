create user postgre with password 'postgre';
create database forum owner postgre;
grant all privileges on DATABASE forum to postgre;

create table if not exists users
(
    id bigserial primary key,
    nickname varchar(60) not null,
    fullname varchar(60),
    about    text,
    email    varchar(60) not null,
);

create table forums (
    id bigserial primary key,
    title varchar(60) not null,
    user varchar(60) not null,
    slug varchar(60) not null,
);

create table threads (
    id bigserial primary key,
    title varchar(60) not null,
    slug varchar(60) not null,
    message text not null,
    author varchar(60) not null,
    forum varchar(60) not null,
    created date,
);

create table posts (
    id bigserial primary key,
    parent_id bigint references posts (id) on delete cascade,
    author varchar(60) not null,
    message text not null,
    is_edited boolean not null,
    forum varchar(60) not null,
    thread bigint references threads (id) on delete cascade,
    created date,
    path text,
);

CREATE TRIGGER set_path BEFORE INSERT ON posts
    FOR EACH ROW SET NEW.path =
  CONCAT(IFNULL((select path from posts where id = NEW.parent_id), '0'), '.', New.id);

create type state as enum (-1, +1);

create table votes (
    thread_id bigint references threads (id) on delete cascade,
    user varchar(60) not null,
    voice state not null,
);