-- +goose Up
INSERT INTO organizations (id, name, slug, status)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'InfluencerEdge System',
    'influencer-edge-system',
    'active'
)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO roles (id, scope_type, scope_id, name, description)
SELECT
    '00000000-0000-0000-0000-000000000002',
    'organization',
    o.id,
    'admin',
    'Platform administrator'
FROM organizations o
WHERE o.slug = 'influencer-edge-system'
ON CONFLICT (scope_type, scope_id, name) DO NOTHING;

INSERT INTO user_roles (id, user_id, role_id, organization_id)
SELECT
    gen_random_uuid(),
    u.id,
    r.id,
    o.id
FROM users u
CROSS JOIN organizations o
INNER JOIN roles r ON r.scope_type = 'organization' AND r.scope_id = o.id AND r.name = 'admin'
WHERE u.email = 'fatma.oz315@gmail.com'
  AND o.slug = 'influencer-edge-system'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM user_roles ur
USING roles r, organizations o
WHERE ur.role_id = r.id
  AND ur.organization_id = o.id
  AND r.name = 'admin'
  AND o.slug = 'influencer-edge-system';

DELETE FROM roles r
USING organizations o
WHERE r.scope_type = 'organization'
  AND r.scope_id = o.id
  AND r.name = 'admin'
  AND o.slug = 'influencer-edge-system';

DELETE FROM organizations WHERE slug = 'influencer-edge-system';
