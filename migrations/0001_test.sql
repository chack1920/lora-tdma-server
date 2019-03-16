-- +migrate Up
create table test (
    id bigserial primary key,
    created_at timestamp with time zone not null,
    f_cnt int not null
);

-- +migrate Down
drop table test;

