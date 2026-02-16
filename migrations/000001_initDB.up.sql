-- 2. Функция для авто-обновления updated_at
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 3. Таблица USERS
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(150),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 4. Таблица ROLES
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 5. Таблица PERMISSIONS
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 6. Связь PERMISSION <-> ROLES (Многие-ко-многим)
CREATE TABLE permissions_roles (
    permission_id INTEGER REFERENCES permissions(id) ON DELETE CASCADE,
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- 7. Связь USERS <-> ROLES (Многие-ко-многим)
-- Мы делаем так, потому что юзер тоже может иметь несколько ролей
CREATE TABLE users_roles (
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- 8. Маппинг групп AD к ролям (Многие-ко-многим)
-- То, что мы обсуждали: одна группа может давать несколько ролей!
CREATE TABLE ldap_group_roles (
    id SERIAL PRIMARY KEY,
    ad_group_name VARCHAR(255) NOT NULL,
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    UNIQUE (ad_group_name, role_id)
);

-- 9. Вешаем триггеры на обновление времени
CREATE TRIGGER update_user_modtime BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE update_modified_column();
CREATE TRIGGER update_role_modtime BEFORE UPDATE ON roles FOR EACH ROW EXECUTE PROCEDURE update_modified_column();
CREATE TRIGGER update_perm_modtime BEFORE UPDATE ON permissions FOR EACH ROW EXECUTE PROCEDURE update_modified_column();