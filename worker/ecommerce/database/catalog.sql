create extension if not exists vector;

create table if not exists origin (
    id uuid primary key default gen_random_uuid(),
    code varchar(50) not null unique
);

create table if not exists category (
    id uuid primary key default gen_random_uuid(),
    code varchar(50) not null unique
);

create table if not exists shop (
    id uuid primary key default gen_random_uuid(),
    origin_id uuid not null,
    category_id uuid not null,
    host varchar(50) not null unique,
    name varchar(50),
    constraint fk_origin_id
        foreign key (origin_id) references origin (id),
    constraint fk_category_id
        foreign key (category_id) references category (id)
);

create table if not exists product (
    id uuid primary key default gen_random_uuid(),
    shop_id uuid not null,
    external_id varchar(100) not null,
    handle varchar(100),
    url varchar(255),
    title varchar(100),
    vendor varchar(50),
    description text,
    specs jsonb,
    summary text,
    raw_key varchar(50),
    product_hash bytea not null,
    embedding vector(1024),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    constraint fk_shop_id
        foreign key (shop_id) references shop (id),
    constraint uq_product_shop_external unique (shop_id, external_id)
);
