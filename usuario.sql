CREATE USER sistema_balance WITH PASSWORD 'sistema_balance_contra_super_secreta';
GRANT ALL PRIVILEGES ON DATABASE sistema_balance_test TO sistema_balance;

GRANT USAGE ON SCHEMA public TO sistema_balance;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO sistema_balance;

GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO sistema_balance;
