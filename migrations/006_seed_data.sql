-- Description: Initial seed — admin user, two outlets (matches Flutter sample data),
--              their categories, and third-party config stubs.
-- Run order: 6  (safe to re-run)

-- Admin user 
-- Password: Admin@1234  (bcrypt hash — change in production!)

INSERT INTO users (id, name, email, mobile, password_hash, role, is_active)
VALUES (
    'aaaaaaaa-0000-0000-0000-000000000001',
    'System Admin',
    'admin@prayosha.com',
    '9999999999',
    '$2a$12$K8GpN.jOWGmHHKpPsJwT8.YN7X6lQ5WsF4bOz1mJDfG3aP5Ue0zBi',   -- Admin@1234
    'admin',
    TRUE
) ON CONFLICT (email) DO NOTHING;

-- Outlets

INSERT INTO outlets (id, name, ref_id, type, city, state, is_active)
VALUES
    ('bbbbbbbb-0000-0000-0000-000000000001',
     'Aarthi Cake Magic',
     '363317',
     'dine_in',
     'Chennai', 'Tamil Nadu', TRUE),
    ('bbbbbbbb-0000-0000-0000-000000000002',
     'Ambattur Aarthi Sweets and Bakery',
     '383514',
     'dine_in',
     'Ambattur', 'Tamil Nadu', TRUE)
ON CONFLICT (ref_id) DO NOTHING;

-- Outlet access for admin

INSERT INTO outlet_accesses (user_id, outlet_id, role)
VALUES
    ('aaaaaaaa-0000-0000-0000-000000000001', 'bbbbbbbb-0000-0000-0000-000000000001', 'admin'),
    ('aaaaaaaa-0000-0000-0000-000000000001', 'bbbbbbbb-0000-0000-0000-000000000002', 'admin')
ON CONFLICT (user_id, outlet_id) DO NOTHING;

-- Default zones

INSERT INTO zones (id, name, outlet_id)
VALUES
    ('cccccccc-0000-0000-0000-000000000001', 'Main Hall',  'bbbbbbbb-0000-0000-0000-000000000001'),
    ('cccccccc-0000-0000-0000-000000000002', 'Takeaway',   'bbbbbbbb-0000-0000-0000-000000000001'),
    ('cccccccc-0000-0000-0000-000000000003', 'Main Hall',  'bbbbbbbb-0000-0000-0000-000000000002')
ON CONFLICT DO NOTHING;

-- Categories

INSERT INTO categories (id, name, outlet_id, sort_order, is_active)
VALUES
    -- Outlet 1
    ('dddddddd-0000-0000-0000-000000000001', 'Cakes',      'bbbbbbbb-0000-0000-0000-000000000001', 1, TRUE),
    ('dddddddd-0000-0000-0000-000000000002', 'Pastries',   'bbbbbbbb-0000-0000-0000-000000000001', 2, TRUE),
    ('dddddddd-0000-0000-0000-000000000003', 'Beverages',  'bbbbbbbb-0000-0000-0000-000000000001', 3, TRUE),
    -- Outlet 2
    ('dddddddd-0000-0000-0000-000000000004', 'Sweets',     'bbbbbbbb-0000-0000-0000-000000000002', 1, TRUE),
    ('dddddddd-0000-0000-0000-000000000005', 'Bakery',     'bbbbbbbb-0000-0000-0000-000000000002', 2, TRUE),
    ('dddddddd-0000-0000-0000-000000000006', 'Beverages',  'bbbbbbbb-0000-0000-0000-000000000002', 3, TRUE)
ON CONFLICT DO NOTHING;

-- Sample menu items 

INSERT INTO menu_items (id, name, category_id, outlet_id, price, tax_rate, is_veg, sort_order)
VALUES
    -- Outlet 1 cakes
    ('eeeeeeee-0000-0000-0000-000000000001', 'Chocolate Truffle Cake',
     'dddddddd-0000-0000-0000-000000000001', 'bbbbbbbb-0000-0000-0000-000000000001',
     850.00, 5.00, TRUE, 1),
    ('eeeeeeee-0000-0000-0000-000000000002', 'Vanilla Birthday Cake',
     'dddddddd-0000-0000-0000-000000000001', 'bbbbbbbb-0000-0000-0000-000000000001',
     650.00, 5.00, TRUE, 2),
    -- Outlet 1 pastries
    ('eeeeeeee-0000-0000-0000-000000000003', 'Croissant',
     'dddddddd-0000-0000-0000-000000000002', 'bbbbbbbb-0000-0000-0000-000000000001',
     80.00, 5.00, TRUE, 1),
    -- Outlet 2 sweets
    ('eeeeeeee-0000-0000-0000-000000000004', 'Kaju Katli',
     'dddddddd-0000-0000-0000-000000000004', 'bbbbbbbb-0000-0000-0000-000000000002',
     600.00, 5.00, TRUE, 1),
    ('eeeeeeee-0000-0000-0000-000000000005', 'Mysore Pak',
     'dddddddd-0000-0000-0000-000000000004', 'bbbbbbbb-0000-0000-0000-000000000002',
     400.00, 5.00, TRUE, 2)
ON CONFLICT DO NOTHING;

-- Third-party config stubs

INSERT INTO third_party_configs (outlet_id, platform, is_active, config)
VALUES
    ('bbbbbbbb-0000-0000-0000-000000000001', 'zomato',    FALSE, '{}'),
    ('bbbbbbbb-0000-0000-0000-000000000001', 'swiggy',    FALSE, '{}'),
    ('bbbbbbbb-0000-0000-0000-000000000001', 'foodpanda', FALSE, '{}'),
    ('bbbbbbbb-0000-0000-0000-000000000002', 'zomato',    FALSE, '{}'),
    ('bbbbbbbb-0000-0000-0000-000000000002', 'swiggy',    FALSE, '{}')
ON CONFLICT (outlet_id, platform) DO NOTHING;
