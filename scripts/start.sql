-- create user postgre with password 'postgre';
-- create database forum owner postgre;
-- grant all privileges on database forum to postgre;

create table if not exists users
(
    id bigserial primary key,
    nickname varchar(60) not null,
    fullname varchar(60),
    about    text,
    email    varchar(60) not null
);

create table if not exists forums (
    id bigserial primary key,
    title text not null,
    u varchar(60) not null,
    slug varchar(60) not null
);

create table if not exists threads (
    id bigserial primary key,
    title text not null,
    slug varchar(60) not null,
    message text not null,
    author varchar(60) not null,
    forum varchar(60) not null,
    created timestamp with time zone
);

create table if not exists posts (
    id bigserial primary key,
    parent_id bigint references posts (id) on delete cascade,
    author varchar(60) not null,
    message text not null,
    is_edited boolean not null,
    forum varchar(60) not null,
    thread bigint references threads (id) on delete cascade,
    created timestamp with time zone,
    path text
);
CREATE OR REPLACE FUNCTION update_posts()
    RETURNS trigger AS
  $$
    BEGIN
    NEW.path = CONCAT(
        CASE WHEN IFNULL(, '0'), '.', New.id);
    RETURN NEW;
    END;
  $$
LANGUAGE 'plpgsql';


CREATE TRIGGER set_path BEFORE INSERT ON posts
    FOR EACH ROW
    EXECUTE PROCEDURE update_posts();

create type state as enum ('-1', '+1');

create table if not exists votes (
    thread_id bigint references threads (id) on delete cascade,
    u varchar(60) not null,
    voice state not null
);