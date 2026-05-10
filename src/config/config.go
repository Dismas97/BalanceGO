package config

import (
    "time"
	"os"
	"log"
	"strconv"
)

type Config struct {
	ServerPort string
	SecretJWT []byte
    DBHost string
    DBPort string
    DBUser string
    DBPassword string
    DBName string
    DBMaxOpenConns int
    DBMaxIdleConns int
    DBConnMaxLifetime time.Duration
	SesionRefreshMargen int
	LogPath string
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if strValue := os.Getenv(key); strValue != "" {
		if intValue, err := strconv.Atoi(strValue); err == nil {
			return intValue
		}
		log.Printf("Advertencia: variable %s = '%s' no es un entero válido, usando %d", key, strValue, defaultValue)
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if strValue := os.Getenv(key); strValue != "" {
		if durValue, err := time.ParseDuration(strValue); err == nil {
			return durValue
		}
		log.Printf("Advertencia: variable %s = '%s' no es una duración válida (ej: '5m', '2h'), usando %v", key, strValue, defaultValue)
	}
	return defaultValue
}

var MainConfig Config = Config{
		ServerPort:          getEnvString("SERVER_PORT", "9821"),
		SecretJWT:           []byte(getEnvString("SECRET_JWT", "una-contra-super-mega-muy-secreta")),
		DBHost:              getEnvString("DB_HOST", "127.0.0.1"),
		DBPort:              getEnvString("DB_PORT", "5432"),
		DBUser:              getEnvString("DB_USER", "sistema_balance"),
		DBPassword:          getEnvString("DB_PASSWORD", "sistema_balance_contra_super_secreta"),
		DBName:              getEnvString("DB_NAME", "sistema_balance_test"),
		DBMaxOpenConns:      getEnvInt("DB_MAX_OPEN_CONNS", 10),
		DBMaxIdleConns:      getEnvInt("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxLifetime:   getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		SesionRefreshMargen: getEnvInt("SESSION_REFRESH_MARGEN", 12),
		LogPath:             getEnvString("LOG_PATH", "./logs/full.log"),
	}
