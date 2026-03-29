-- Categories (hierarchical via parent_id)
create table categories (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  slug text unique not null,
  parent_id uuid references categories(id) on delete set null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index idx_categories_slug on categories(slug);
create index idx_categories_parent_id on categories(parent_id);

-- Category attributes (defines what attributes a category's products have)
create type attribute_type as enum ('text', 'number', 'enum');

create table category_attributes (
  id uuid primary key default gen_random_uuid(),
  category_id uuid not null references categories(id) on delete cascade,
  name text not null,
  type attribute_type not null default 'text',
  required boolean not null default false,
  sort_order integer not null default 0
);

create index idx_category_attributes_category on category_attributes(category_id);

-- Attribute options (enum values for enum-type attributes)
create table attribute_options (
  id uuid primary key default gen_random_uuid(),
  category_attribute_id uuid not null references category_attributes(id) on delete cascade,
  value text not null,
  sort_order integer not null default 0
);

create index idx_attribute_options_attr on attribute_options(category_attribute_id);

-- Enable RLS
alter table categories enable row level security;
alter table category_attributes enable row level security;
alter table attribute_options enable row level security;

-- Public read access for catalog
create policy "categories_public_read" on categories
  for select using (true);

create policy "category_attributes_public_read" on category_attributes
  for select using (true);

create policy "attribute_options_public_read" on attribute_options
  for select using (true);

-- Service role has full access (Go API uses service_role key)
create policy "categories_service_all" on categories
  for all using (auth.role() = 'service_role');

create policy "category_attributes_service_all" on category_attributes
  for all using (auth.role() = 'service_role');

create policy "attribute_options_service_all" on attribute_options
  for all using (auth.role() = 'service_role');

-- Updated_at trigger
create or replace function update_updated_at()
returns trigger as $$
begin
  new.updated_at = now();
  return new;
end;
$$ language plpgsql;

create trigger categories_updated_at
  before update on categories
  for each row execute function update_updated_at();
