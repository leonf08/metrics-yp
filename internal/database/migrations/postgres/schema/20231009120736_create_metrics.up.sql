create table if not exists metrics(
    name text primary key,
    type text not null,
    value double precision not null
);