-- Products
create type product_status as enum ('draft', 'active', 'archived');

create table products (
  id uuid primary key default gen_random_uuid(),
  category_id uuid not null references categories(id) on delete restrict,
  name text not null,
  slug text unique not null,
  description text,
  base_price numeric(10,2) not null check (base_price >= 0),
  status product_status not null default 'draft',
  images text[] default '{}',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index idx_products_slug on products(slug);
create index idx_products_category on products(category_id);
create index idx_products_status on products(status);

-- SKUs
create type sku_status as enum ('active', 'inactive');

create table skus (
  id uuid primary key default gen_random_uuid(),
  product_id uuid not null references products(id) on delete cascade,
  sku_code text unique not null,
  price_override numeric(10,2) check (price_override >= 0),
  status sku_status not null default 'active',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index idx_skus_product on skus(product_id);
create index idx_skus_code on skus(sku_code);

-- SKU attribute values
create table sku_attribute_values (
  id uuid primary key default gen_random_uuid(),
  sku_id uuid not null references skus(id) on delete cascade,
  category_attribute_id uuid not null references category_attributes(id) on delete cascade,
  value text not null
);

create index idx_sku_attr_values_sku on sku_attribute_values(sku_id);

-- Custom fields (reporting metadata only)
create type entity_type as enum ('product', 'sku');

create table custom_fields (
  id uuid primary key default gen_random_uuid(),
  entity_type entity_type not null,
  entity_id uuid not null,
  key text not null,
  value text not null
);

create index idx_custom_fields_entity on custom_fields(entity_type, entity_id);
create index idx_custom_fields_key on custom_fields(key);

-- Page views (for admin report)
create table page_views (
  id uuid primary key default gen_random_uuid(),
  product_id uuid not null references products(id) on delete cascade,
  session_id text,
  viewed_at timestamptz not null default now()
);

create index idx_page_views_product on page_views(product_id);

-- RLS policies
alter table products enable row level security;
alter table skus enable row level security;
alter table sku_attribute_values enable row level security;
alter table custom_fields enable row level security;
alter table page_views enable row level security;

-- Public read
create policy "products_public_read" on products
  for select using (status = 'active');

create policy "skus_public_read" on skus
  for select using (status = 'active');

create policy "sku_attr_values_public_read" on sku_attribute_values
  for select using (true);

create policy "custom_fields_public_read" on custom_fields
  for select using (true);

create policy "page_views_public_insert" on page_views
  for insert with check (true);

-- Service role full access
create policy "products_service_all" on products
  for all using (auth.role() = 'service_role');

create policy "skus_service_all" on skus
  for all using (auth.role() = 'service_role');

create policy "sku_attr_values_service_all" on sku_attribute_values
  for all using (auth.role() = 'service_role');

create policy "custom_fields_service_all" on custom_fields
  for all using (auth.role() = 'service_role');

create policy "page_views_service_all" on page_views
  for all using (auth.role() = 'service_role');

-- Updated_at triggers
create trigger products_updated_at
  before update on products
  for each row execute function update_updated_at();

create trigger skus_updated_at
  before update on skus
  for each row execute function update_updated_at();
