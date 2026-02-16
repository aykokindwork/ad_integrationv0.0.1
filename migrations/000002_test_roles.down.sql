-- Удаляем маппинг групп (самое вложенное)
DELETE FROM ldap_group_roles WHERE ad_group_name IN ('lab-test-admins');

-- Удаляем связи прав и ролей
DELETE FROM roles_permissions WHERE role_id IN (SELECT id FROM roles WHERE code IN ('admin', 'editor', 'user'));

-- Удаляем сами права
DELETE FROM permissions WHERE code IN ('users.delete', 'users.create', 'users.view');

-- Удаляем роли
DELETE FROM roles WHERE code IN ('admin', 'editor', 'user');  