--do not forget to create a user OWNER this database

-- 1. Создаем схему, если её еще нет

CREATE SCHEMA IF NOT EXISTS auth;

-- 2. Функция для авто-обновления updated_at
CREATE OR REPLACE FUNCTION auth.update_modified_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 3. Таблица USERS
CREATE TABLE IF NOT EXISTS auth.users (
                                          id SERIAL PRIMARY KEY,
                                          login VARCHAR(50) UNIQUE NOT NULL,
                                          email VARCHAR(150),
                                          created_at TIMESTAMPTZ DEFAULT NOW(),
                                          updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 4. Таблица ROLES
CREATE TABLE IF NOT EXISTS auth.roles (
                                          id SERIAL PRIMARY KEY,
                                          code VARCHAR(50) UNIQUE NOT NULL,
                                          name VARCHAR(100) NOT NULL,
                                          created_at TIMESTAMPTZ DEFAULT NOW(),
                                          updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 5. Таблица PERMISSIONS
CREATE TABLE IF NOT EXISTS auth.permissions (
                                                id SERIAL PRIMARY KEY,
                                                code VARCHAR(50) UNIQUE NOT NULL,
                                                name VARCHAR(100) NOT NULL,
                                                created_at TIMESTAMPTZ DEFAULT NOW(),
                                                updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 6. Связь PERMISSION <-> ROLES
CREATE TABLE IF NOT EXISTS auth.permissions_roles (
                                                      permission_id INTEGER REFERENCES auth.permissions(id) ON DELETE CASCADE,
                                                      role_id INTEGER REFERENCES auth.roles(id) ON DELETE CASCADE,
                                                      PRIMARY KEY (role_id, permission_id)
);

-- 7. Связь USERS <-> ROLES
CREATE TABLE IF NOT EXISTS auth.users_roles (
                                                user_id INTEGER REFERENCES auth.users(id) ON DELETE CASCADE,
                                                role_id INTEGER REFERENCES auth.roles(id) ON DELETE CASCADE,
                                                assigned_at TIMESTAMPTZ DEFAULT NOW(),
                                                PRIMARY KEY (user_id, role_id)
);

-- 8. Маппинг групп AD к ролям
CREATE TABLE IF NOT EXISTS auth.ldap_group_roles (
                                                     id SERIAL PRIMARY KEY,
                                                     ad_group_name VARCHAR(255) NOT NULL,
                                                     role_id INTEGER REFERENCES auth.roles(id) ON DELETE CASCADE,
                                                     UNIQUE (ad_group_name, role_id)
);

-- 9. Триггеры (через DO блок, чтобы не падали при повторном запуске)
DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_user_modtime') THEN
            CREATE TRIGGER update_user_modtime BEFORE UPDATE ON auth.users FOR EACH ROW EXECUTE PROCEDURE auth.update_modified_column();
        END IF;
        IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_role_modtime') THEN
            CREATE TRIGGER update_role_modtime BEFORE UPDATE ON auth.roles FOR EACH ROW EXECUTE PROCEDURE auth.update_modified_column();
        END IF;
        IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_perm_modtime') THEN
            CREATE TRIGGER update_perm_modtime BEFORE UPDATE ON auth.permissions FOR EACH ROW EXECUTE PROCEDURE auth.update_modified_column();
        END IF;
    END $$;

-- 10. ЗАПОЛНЕНИЕ ДАННЫМИ
INSERT INTO auth.roles (code, name)
VALUES ('admin', 'Администратор'), ('editor', 'Редактор')
ON CONFLICT (code) DO NOTHING;

INSERT INTO auth.permissions (code, name)
VALUES ('users.delete', 'Удаление пользователей'), ('users.create', 'Создание пользователей')
ON CONFLICT (code) DO NOTHING;

INSERT INTO auth.permissions_roles (role_id, permission_id)
SELECT r.id, p.id FROM auth.roles r, auth.permissions p
WHERE r.code = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO auth.ldap_group_roles (ad_group_name, role_id)
SELECT 'lab-test-admins', id FROM auth.roles WHERE code = 'admin'
ON CONFLICT DO NOTHING;