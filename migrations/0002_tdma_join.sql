-- +migrate Up
create table tdma_join (
    id bigserial primary key,
    created_at timestamp with time zone not null,
    dev_eui bytea not null,
    mc_seq int not null,
    tx_cycle int not null
);

-- +migrate Down
drop table tdma_join;

