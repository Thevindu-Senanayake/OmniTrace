--   ...0001  Standard Widget   500  never runs out
--   ...0002  Premium Widget    100  runs out mid-test
--   ...0003  Hot Gadget         20  depletes fast under load
--   ...0004  Flash Sale Gizmo    5  gone in seconds
--   ...0099  Ghost Item          0  always OUT_OF_STOCK immediately

INSERT INTO inventory (item_id, stock) VALUES
    ('00000000-0000-0000-0000-000000000001', 500),
    ('00000000-0000-0000-0000-000000000002', 100),
    ('00000000-0000-0000-0000-000000000003', 20),
    ('00000000-0000-0000-0000-000000000004', 5),
    ('00000000-0000-0000-0000-000000000099', 0)
ON CONFLICT (item_id) DO NOTHING;
