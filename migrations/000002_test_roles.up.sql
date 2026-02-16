-- 1. Создаем базовые роли
INSERT INTO roles (code, name) VALUES 
('admin', 'Администратор'),
('editor', 'Редактор');

-- 2. Создаем конкретные права
INSERT INTO permissions (code, name) VALUES 
('users.delete', 'Удаление пользователей'),
('users.create', 'Создание пользователей');

-- 3. Привязываем права к роли admin (даем ему всё)
INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id 
FROM roles r, permissions p 
WHERE r.code = 'admin';

-- 4. ТВОЯ ГЛАВНАЯ СВЯЗКА (LDAP -> DB)
INSERT INTO ldap_group_roles (ad_group_name, role_id)
VALUES ('lab-test-admins', (SELECT id FROM roles WHERE code = 'admin'));