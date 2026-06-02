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
	DB = db
    return nil
}



/*
func InitDatabase(db_con *sqlx.DB) error {
	emp := dto.Empresa{
        ID: 1,
        Nombre: "EmpresaSuperAdmin",
        Telefono:  sql.NullString{String: "12312123", Valid: true},
        Email: sql.NullString{String: "admin@test.com", Valid: true},
        Direccion: sql.NullString{String: "Direccion1", Valid: true},
    }
	sql_query := `INSERT INTO empresa (id, nombre, telefono, email, direccion) VALUES (:id, :nombre, :telefono, :email, :direccion) ON CONFLICT (id) DO NOTHING;`
	
	_, err := db_con.NamedExec(sql_query,emp)
    if err != nil {
		log.Printf("Error: %v", err)
        return  err
    }

	usr := dto.Usuario{
		ID: 1,
		Usuario: "admin",
        Apellido: sql.NullString{String: "ApellidoAdmin", Valid: true},
		Nombre: sql.NullString{String: "SuperAdmin", Valid: true},
        Telefono:  sql.NullString{String: "12312123", Valid: true},
        Email: sql.NullString{String: "admin@test.com", Valid: true},
        Direccion: sql.NullString{String: "Direccion1", Valid: true},
	}

	contra_aux, err := crypto.HashPassword("123456")
    if err != nil {
		log.Printf("Error: %v", err)
        return  err
    }
	
	usr.Contra = contra_aux
	
	sql_query = `INSERT INTO usuario (id, usuario, contra, nombre, apellido, telefono, email, direccion, empresa_id) VALUES (1, :usuario, :contra, :nombre, :apellido, :telefono, :email, :direccion, 1) ON CONFLICT (usuario,empresa_id) DO NOTHING`
	
	_, err = db_con.NamedExec(sql_query,usr)
	return err
}*/
