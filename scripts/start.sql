-- create user postgre with password 'postgre';
-- create database forum owner postgre;
-- grant all privileges on database forum to postgre;


create table if not exists users
(
    id bigserial primary key,
    nickname citext unique not null,
    fullname text,
    about    text,
    email    citext unique not null
);

create index idx_nickname on users (nickname);

create table if not exists forums (
    id bigserial primary key,
    title text not null,
    u citext not null,
    slug citext unique not null,
    foreign key (u) References users(nickname)
);

create index idx_forum_slug on forums(slug);

create table if not exists threads (
    id bigserial primary key,
    title text not null,
    slug citext not null,
    message text not null,
    author  citext not null,
    forum citext not null,
    created timestamp with time zone,
    foreign key (author) references users(nickname),
    foreign key (forum) references forums(slug)
);

create index idx_thread_slug on forums(slug);

create table if not exists posts (
    id bigserial primary key,
    parent bigint not null,
    author citext not null,
    message text not null,
    is_edited boolean not null,
    forum citext not null,
    thread bigint references threads (id) on delete cascade,
    created timestamp with time zone,
    path text,
    foreign key (author) references users(nickname),
    foreign key (forum) references forums(slug)
);

create index idx_parent_thread on posts(id, thread);

CREATE OR REPLACE FUNCTION update_posts()
    RETURNS trigger AS
  $$
    BEGIN
    NEW.path = CONCAT(
        coalesce((select path from posts where id = NEW.parent), '0'), '.', New.id);
    RETURN NEW;
    END;
  $$
LANGUAGE 'plpgsql';


CREATE TRIGGER set_path BEFORE INSERT ON posts
    FOR EACH ROW
    EXECUTE PROCEDURE update_posts();

create type state as enum ('-1', '1');

create table if not exists votes (
    thread_id bigint references threads (id) on delete cascade,
    u citext not null,
    voice state not null,
    foreign key (u) references users(nickname)
);