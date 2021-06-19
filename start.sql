CREATE EXTENSION citext;

create unlogged table if not exists users
(
    id bigserial primary key,
    nickname citext COLLATE "C" unique not null,
    fullname text,
    about    text,
    email    citext unique not null
);

create index idx_users_nickname on users using hash(nickname);
create index idx_users_email on users using hash(email);

create unlogged table if not exists forums (
    id bigserial primary key,
    title text not null,
    u citext not null,
    slug citext unique not null,
    posts bigint default 0,
    threads bigint default 0,
    foreign key (u) References users(nickname)
);

create index idx_forum_slug on forums using hash(slug);

create unlogged table if not exists threads (
    id bigserial primary key,
    title text not null,
    slug citext not null,
    message text not null,
    author  citext not null,
    forum citext not null,
    created timestamp with time zone,
    votes int,
    foreign key (author) references users(nickname),
    foreign key (forum) references forums(slug)
);

create index idx_thread_slug on threads using hash(slug);
create index idx_thread_forum on threads(forum);
create index idx_thread_forum_created on threads(forum, created);

create unlogged table if not exists posts (
    id bigserial primary key,
    parent bigint not null,
    author citext not null,
    message text not null,
    is_edited boolean not null,
    forum citext not null,
    thread bigint references threads (id),
    created timestamp with time zone,
    way bigint[],
    foreign key (author) references users(nickname),
    foreign key (forum) references forums(slug)
);

create index idx_posts_thread on posts (thread);
create index idx_posts_way on posts (way);
create index idx_posts_way2 on posts ((way[2]));
create index idx_posts_forum on posts (forum);

CREATE OR REPLACE FUNCTION update_posts()
    RETURNS trigger AS
  $$
declare
parent_way   bigint[];
parent_thread   bigint;
    BEGIN
    if (new.parent = 0) then
        new.way = array[0,new.id];
    else
        select p.way, p.thread
        from posts p
        where p.id = new.parent
        into parent_way, parent_thread;
        if parent_thread != new.thread or parent_thread is null then
            RAISE EXCEPTION USING ERRCODE = '00409';
        end if;
        new.way := parent_way || new.id;
    end if;
    RETURN NEW;
    END;
  $$
LANGUAGE 'plpgsql';


CREATE TRIGGER set_way BEFORE INSERT ON posts
    FOR EACH ROW
    EXECUTE PROCEDURE update_posts();

create unlogged table if not exists votes (
    thread_id bigint references threads (id) on delete cascade,
    u citext not null,
    voice int not null,
    foreign key (u) references users(nickname)
);

create index idx_find_votes on votes(u,thread_id);

CREATE OR REPLACE FUNCTION update_vote()
    RETURNS trigger AS
  $$
BEGIN
UPDATE threads SET votes = votes + NEW.voice WHERE id = NEW.thread_id;
RETURN NULL;
END;
  $$
LANGUAGE 'plpgsql';

CREATE OR REPLACE FUNCTION change_vote()
    RETURNS trigger AS
  $$
BEGIN
UPDATE threads SET votes = (votes + NEW.voice - OLD.voice) WHERE id = NEW.thread_id;
RETURN NULL;
END;
  $$
LANGUAGE 'plpgsql';

CREATE OR REPLACE FUNCTION add_post()
    RETURNS trigger AS
  $$
BEGIN
UPDATE forums SET posts = posts + 1 WHERE slug = NEW.forum;
RETURN NULL;
END;
  $$
LANGUAGE 'plpgsql';

CREATE OR REPLACE FUNCTION add_thread()
    RETURNS trigger AS
  $$
BEGIN
UPDATE forums SET threads = threads + 1 WHERE slug = NEW.forum;
RETURN NULL;
END;
  $$
LANGUAGE 'plpgsql';

CREATE TRIGGER set_vote AFTER INSERT ON votes
    FOR EACH ROW
    EXECUTE PROCEDURE update_vote();

CREATE TRIGGER update_vote AFTER update ON votes
    FOR EACH ROW
    EXECUTE PROCEDURE change_vote();

CREATE TRIGGER add_post AFTER INSERT ON posts
    FOR EACH ROW
    EXECUTE PROCEDURE add_post();

CREATE TRIGGER add_thread AFTER INSERT ON threads
    FOR EACH ROW
    EXECUTE PROCEDURE add_thread();

create unlogged table if not exists forum_users (
    forum citext not null,
    u citext COLLATE "C" not null,
    fullname text,
    about    text,
    email    citext,
    unique(forum, u),
    foreign key (forum) references forums(slug),
    foreign key (u) references users(nickname)
);

create index idx_fu_all on forum_users(forum,u);
create index idx_fu_forum on forum_users(forum);

CREATE OR REPLACE FUNCTION add_user_to_forum_thread()
    RETURNS trigger AS
  $$
BEGIN
insert into forum_users (forum, u, fullname, about, email) select new.forum, new.author, fullname, about, email from users where new.author = nickname on conflict do nothing;
RETURN NULL;
END;
  $$
LANGUAGE 'plpgsql';

CREATE OR REPLACE FUNCTION add_user_to_forum_post()
    RETURNS trigger AS
  $$
BEGIN
insert into forum_users (forum, u, fullname, about, email) select new.forum, new.author, fullname, about, email from users where new.author = nickname on conflict do nothing;
RETURN NULL;
END;
  $$
LANGUAGE 'plpgsql';


CREATE TRIGGER add_user_forum_thread AFTER INSERT ON threads
    FOR EACH ROW
    EXECUTE PROCEDURE add_user_to_forum_thread();

CREATE TRIGGER add_user_forum_post AFTER INSERT ON posts
    FOR EACH ROW
    EXECUTE PROCEDURE add_user_to_forum_post();