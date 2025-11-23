create table if not exists users (
    internal_id bigserial primary key,
    id text UNIQUE NOT NULL,
    name text not null,
    is_active boolean not null
);
create table if not exists teams (
    name text primary key
);
create table if not exists teams_users (
    internal_id bigserial primary key,
    team_name text references teams(name),
    user_id text references users(id)
);
create table if not exists pull_requests (
    internal_id bigserial primary key,
    id text UNIQUE NOT NULL,
    name text not null,
    author_id text references users(id),
    status text not null,
    merged_at timestamp
);
create table if not exists pr_reviewers (
    internal_id bigserial primary key,
    pull_request_id text references pull_requests(id),
    reviewer_id text references users(id)
);