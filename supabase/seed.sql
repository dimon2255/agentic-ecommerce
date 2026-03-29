-- Categories
insert into categories (id, name, slug, parent_id) values
  ('a1000000-0000-0000-0000-000000000001', 'Electronics', 'electronics', null),
  ('a1000000-0000-0000-0000-000000000002', 'Clothing', 'clothing', null),
  ('a1000000-0000-0000-0000-000000000003', 'Laptops', 'laptops', 'a1000000-0000-0000-0000-000000000001'),
  ('a1000000-0000-0000-0000-000000000004', 'T-Shirts', 't-shirts', 'a1000000-0000-0000-0000-000000000002');

-- Category Attributes
-- Laptops: RAM, Storage
insert into category_attributes (id, category_id, name, type, required, sort_order) values
  ('b1000000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000003', 'RAM', 'enum', true, 0),
  ('b1000000-0000-0000-0000-000000000002', 'a1000000-0000-0000-0000-000000000003', 'Storage', 'enum', true, 1);

-- T-Shirts: Size, Color
insert into category_attributes (id, category_id, name, type, required, sort_order) values
  ('b1000000-0000-0000-0000-000000000003', 'a1000000-0000-0000-0000-000000000004', 'Size', 'enum', true, 0),
  ('b1000000-0000-0000-0000-000000000004', 'a1000000-0000-0000-0000-000000000004', 'Color', 'enum', true, 1);

-- Attribute Options
-- RAM: 8GB, 16GB, 32GB
insert into attribute_options (category_attribute_id, value, sort_order) values
  ('b1000000-0000-0000-0000-000000000001', '8GB', 0),
  ('b1000000-0000-0000-0000-000000000001', '16GB', 1),
  ('b1000000-0000-0000-0000-000000000001', '32GB', 2);

-- Storage: 256GB, 512GB, 1TB
insert into attribute_options (category_attribute_id, value, sort_order) values
  ('b1000000-0000-0000-0000-000000000002', '256GB', 0),
  ('b1000000-0000-0000-0000-000000000002', '512GB', 1),
  ('b1000000-0000-0000-0000-000000000002', '1TB', 2);

-- Size: S, M, L, XL
insert into attribute_options (category_attribute_id, value, sort_order) values
  ('b1000000-0000-0000-0000-000000000003', 'S', 0),
  ('b1000000-0000-0000-0000-000000000003', 'M', 1),
  ('b1000000-0000-0000-0000-000000000003', 'L', 2),
  ('b1000000-0000-0000-0000-000000000003', 'XL', 3);

-- Color: Black, White, Blue, Red
insert into attribute_options (category_attribute_id, value, sort_order) values
  ('b1000000-0000-0000-0000-000000000004', 'Black', 0),
  ('b1000000-0000-0000-0000-000000000004', 'White', 1),
  ('b1000000-0000-0000-0000-000000000004', 'Blue', 2),
  ('b1000000-0000-0000-0000-000000000004', 'Red', 3);

-- Products
insert into products (id, category_id, name, slug, description, base_price, status, images) values
  ('c1000000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000003', 'ProBook 15', 'probook-15', 'Professional laptop with stunning display and all-day battery life.', 999.99, 'active', '{"https://images.unsplash.com/photo-1496181133206-80ce9b88a853?w=600&h=600&fit=crop"}'),
  ('c1000000-0000-0000-0000-000000000002', 'a1000000-0000-0000-0000-000000000003', 'UltraSlim Air', 'ultraslim-air', 'Ultra-lightweight laptop for professionals on the go.', 1299.99, 'active', '{"https://images.unsplash.com/photo-1517336714731-489689fd1ca8?w=600&h=600&fit=crop"}'),
  ('c1000000-0000-0000-0000-000000000003', 'a1000000-0000-0000-0000-000000000004', 'Classic Cotton Tee', 'classic-cotton-tee', 'Soft 100% cotton t-shirt. Essential wardrobe staple.', 24.99, 'active', '{"https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?w=600&h=600&fit=crop"}'),
  ('c1000000-0000-0000-0000-000000000004', 'a1000000-0000-0000-0000-000000000004', 'Performance Dry-Fit', 'performance-dry-fit', 'Moisture-wicking athletic t-shirt for active lifestyles.', 34.99, 'active', '{"https://images.unsplash.com/photo-1581655353564-df123a1eb820?w=600&h=600&fit=crop"}');

-- SKUs for ProBook 15
insert into skus (id, product_id, sku_code, price_override, status) values
  ('d1000000-0000-0000-0000-000000000001', 'c1000000-0000-0000-0000-000000000001', 'PROBOOK-8-256', null, 'active'),
  ('d1000000-0000-0000-0000-000000000002', 'c1000000-0000-0000-0000-000000000001', 'PROBOOK-16-512', 1199.99, 'active'),
  ('d1000000-0000-0000-0000-000000000003', 'c1000000-0000-0000-0000-000000000001', 'PROBOOK-32-1TB', 1499.99, 'active');

-- SKU attribute values for ProBook 15
insert into sku_attribute_values (sku_id, category_attribute_id, value) values
  ('d1000000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '8GB'),
  ('d1000000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000002', '256GB'),
  ('d1000000-0000-0000-0000-000000000002', 'b1000000-0000-0000-0000-000000000001', '16GB'),
  ('d1000000-0000-0000-0000-000000000002', 'b1000000-0000-0000-0000-000000000002', '512GB'),
  ('d1000000-0000-0000-0000-000000000003', 'b1000000-0000-0000-0000-000000000001', '32GB'),
  ('d1000000-0000-0000-0000-000000000003', 'b1000000-0000-0000-0000-000000000002', '1TB');

-- SKUs for Classic Cotton Tee (4 colors x 4 sizes = 16, showing a subset)
insert into skus (id, product_id, sku_code, status) values
  ('d1000000-0000-0000-0000-000000000010', 'c1000000-0000-0000-0000-000000000003', 'CTEE-BLK-S', 'active'),
  ('d1000000-0000-0000-0000-000000000011', 'c1000000-0000-0000-0000-000000000003', 'CTEE-BLK-M', 'active'),
  ('d1000000-0000-0000-0000-000000000012', 'c1000000-0000-0000-0000-000000000003', 'CTEE-BLU-M', 'active'),
  ('d1000000-0000-0000-0000-000000000013', 'c1000000-0000-0000-0000-000000000003', 'CTEE-WHT-L', 'active');

insert into sku_attribute_values (sku_id, category_attribute_id, value) values
  ('d1000000-0000-0000-0000-000000000010', 'b1000000-0000-0000-0000-000000000003', 'S'),
  ('d1000000-0000-0000-0000-000000000010', 'b1000000-0000-0000-0000-000000000004', 'Black'),
  ('d1000000-0000-0000-0000-000000000011', 'b1000000-0000-0000-0000-000000000003', 'M'),
  ('d1000000-0000-0000-0000-000000000011', 'b1000000-0000-0000-0000-000000000004', 'Black'),
  ('d1000000-0000-0000-0000-000000000012', 'b1000000-0000-0000-0000-000000000003', 'M'),
  ('d1000000-0000-0000-0000-000000000012', 'b1000000-0000-0000-0000-000000000004', 'Blue'),
  ('d1000000-0000-0000-0000-000000000013', 'b1000000-0000-0000-0000-000000000003', 'L'),
  ('d1000000-0000-0000-0000-000000000013', 'b1000000-0000-0000-0000-000000000004', 'White');

-- Custom fields (reporting metadata)
insert into custom_fields (entity_type, entity_id, key, value) values
  ('product', 'c1000000-0000-0000-0000-000000000001', 'supplier', 'TechCorp'),
  ('product', 'c1000000-0000-0000-0000-000000000001', 'tag', 'bestseller'),
  ('product', 'c1000000-0000-0000-0000-000000000003', 'supplier', 'Cotton Co'),
  ('product', 'c1000000-0000-0000-0000-000000000003', 'tag', 'essentials'),
  ('product', 'c1000000-0000-0000-0000-000000000003', 'season', 'SS26');
