package bd

import (
    "fmt"
    "github.com/jmoiron/sqlx"
    "sistema-balance/config"
)

var DB *sqlx.DB = nil

func NewConnection(cfg config.Config)  error {
	if DB != nil {
		return nil
	}
    dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
    db, err := sqlx.Connect("postgres", dsn)
    if err != nil {
        return err
    }
    db.SetMaxOpenConns(cfg.DBMaxOpenConns)
    db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.DBConnMaxIdletime)
	DB = db
    return nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

/*
   func InitEmpresa(empresa_id int, db_con *sqlx.DB) error {

   //TODO inicializar cuentas por cada empresa_id, cuentas "magicas" basicas como "ingresos", "inventario", "caja", "provedores", etc. que permitan abstraer cosas como ingresos anonimos de un cliente de la calle, un inventario de lo que existe en stock, provedores que rellenan dicho stock x dinero, etc.

   //TODO inicializar unidades basicas, metricas, masa, volumen, cantidades discretas, cantidades continuas etc.
   //TODO inicalizar activos basicos como "moneda_local", "usd", "eur"
	return err
}*/
