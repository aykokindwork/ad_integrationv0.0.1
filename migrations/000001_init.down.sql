-- 1. Удаляем триггеры (сначала их)
DROP TRIGGER IF EXISTS update_user_modtime ON auth.users;
DROP TRIGGER IF EXISTS update_role_modtime ON auth.roles;
DROP TRIGGER IF EXISTS update_perm_modtime ON auth.permissions;

-- 2. Удаляем таблицы в ОБРАТНОМ порядке (чтобы FK не ругались)
DROP TABLE IF EXISTS auth.ldap_group_roles;
DROP TABLE IF EXISTS auth.users_roles;
DROP TABLE IF EXISTS auth.permissions_roles;
DROP TABLE IF EXISTS auth.permissions;
DROP TABLE IF EXISTS auth.roles;
DROP TABLE IF EXISTS auth.users;

-- 3. Удаляем общие объекты
DROP FUNCTION IF EXISTS auth.update_modified_column();
DROP SCHEMA IF EXISTS auth CASCADE;