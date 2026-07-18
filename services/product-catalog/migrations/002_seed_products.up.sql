-- Five well-known items with FIXED UUIDs. Inventory seeding, Toxiproxy demos
-- and the k6 script reference these deterministic IDs.
INSERT INTO products (id, name, description, price, category) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Standard Widget',  'Everyday widget, always in stock (stock 500)',  9.99,  'widgets'),
    ('00000000-0000-0000-0000-000000000002', 'Premium Widget',   'Higher-grade widget (stock 100)',               24.99, 'widgets'),
    ('00000000-0000-0000-0000-000000000003', 'Hot Gadget',       'Popular gadget that sells out fast (stock 20)', 49.99, 'gadgets'),
    ('00000000-0000-0000-0000-000000000004', 'Flash Sale Gizmo', 'Limited flash-sale item (stock 5)',             99.99, 'gadgets'),
    ('00000000-0000-0000-0000-000000000099', 'Ghost Item',       'Listed but permanently out of stock (stock 0)', 4.99,  'misc')
ON CONFLICT (id) DO NOTHING;

-- 295 generated catalog filler items (deterministic content, random UUIDs)
INSERT INTO products (name, description, price, category)
SELECT
    'Product ' || n,
    'Auto-seeded catalog item #' || n,
    round((5 + (n * 37 % 495) + (n % 100) / 100.0)::numeric, 2),
    (ARRAY['widgets', 'gadgets', 'tools', 'accessories', 'misc'])[1 + (n % 5)]
FROM generate_series(1, 295) AS n;
