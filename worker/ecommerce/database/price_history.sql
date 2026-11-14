create table if not exists price_history (
    id uuid primary key default gen_random_uuid(),
    host varchar(50) not null,
    external_id varchar(100) not null,
    price numeric not null,
    currency varchar(10) not null default 'INR',
    observed_at timestamptz not null default now()
);

create index if not exists price_history_host_external_id_observed_at_idx
    on price_history (host, external_id, observed_at);
