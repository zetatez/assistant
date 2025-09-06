package xdb

import (
	"database/sql"
	"fmt"
	"time"

	"context"
	"net/url"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type DBPoolConfig struct {
	DriverName      string        `mapstructure:"driver_name"`
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdletime time.Duration `mapstructure:"conn_max_idletime"`
	PingTimeout     time.Duration `mapstructure:"ping_timeout"`
}

func NewDBPool(ctx context.Context, cfg DBPoolConfig) (*sql.DB, error) {
	db, err := sql.Open(cfg.DriverName, cfg.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdletime)

	pingCtx := ctx
	if cfg.PingTimeout > 0 {
		var cancel context.CancelFunc
		pingCtx, cancel = context.WithTimeout(ctx, cfg.PingTimeout)
		defer cancel()
	}

	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func BuildDSN(driverName string, params map[string]string) (string, error) {
	switch strings.ToLower(driverName) {
	case "mysql":
		return buildMySQLDSN(params)
	case "obmysql":
		return buildOBMySQLDSN(params)
	case "postgresql":
		return buildPostgresDSN(params)
	case "sqlite":
		path := params["path"]
		if path == "" {
			return "", fmt.Errorf("sqlite path is required")
		}
		return path, nil
	default:
		return "", fmt.Errorf("unsupported driver: %s", driverName)
	}
}

func buildMySQLDSN(params map[string]string) (string, error) {
	user := params["user"]
	if user == "" {
		return "", fmt.Errorf("mysql user is required")
	}

	host := params["host"]
	if host == "" {
		host = "127.0.0.1"
	}

	port := params["port"]
	if port == "" {
		port = "3306"
	}

	dbname := params["dbname"]
	if dbname == "" {
		return "", fmt.Errorf("mysql dbname is required")
	}

	query := url.Values{}
	for key, value := range params {
		if strings.HasPrefix(key, "param.") {
			query.Set(strings.TrimPrefix(key, "param."), value)
		}
	}

	password := params["password"]
	creds := user
	if password != "" {
		creds = fmt.Sprintf("%s:%s", user, password)
	}

	if len(query) > 0 {
		return fmt.Sprintf("%s@tcp(%s:%s)/%s?%s", creds, host, port, dbname, query.Encode()), nil
	}

	return fmt.Sprintf("%s@tcp(%s:%s)/%s", creds, host, port, dbname), nil
}

func buildPostgresDSN(params map[string]string) (string, error) {
	user := params["user"]
	if user == "" {
		return "", fmt.Errorf("postgres user is required")
	}

	host := params["host"]
	if host == "" {
		host = "127.0.0.1"
	}

	port := params["port"]
	if port == "" {
		port = "5432"
	}

	dbname := params["dbname"]
	if dbname == "" {
		return "", fmt.Errorf("postgres dbname is required")
	}

	query := url.Values{}
	for key, value := range params {
		if strings.HasPrefix(key, "param.") {
			query.Set(strings.TrimPrefix(key, "param."), value)
		}
	}

	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, params["password"]),
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   dbname,
	}
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	return u.String(), nil
}

func buildOBMySQLDSN(params map[string]string) (string, error) {
	user := params["user"]
	if user == "" {
		return "", fmt.Errorf("ob mysql user is required")
	}

	tenant := params["tenant"]
	if tenant == "" {
		return "", fmt.Errorf("ob mysql tenant is required")
	}

	cluster := params["cluster"]
	username := fmt.Sprintf("%s@%s", user, tenant)
	if cluster != "" {
		username = fmt.Sprintf("%s#%s", username, cluster)
	}

	host := params["host"]
	if host == "" {
		host = "127.0.0.1"
	}

	port := params["port"]
	if port == "" {
		port = "2883"
	}

	dbname := params["dbname"]
	if dbname == "" {
		return "", fmt.Errorf("ob mysql dbname is required")
	}

	query := url.Values{}
	for key, value := range params {
		if strings.HasPrefix(key, "param.") {
			query.Set(strings.TrimPrefix(key, "param."), value)
		}
	}

	password := params["password"]
	creds := username
	if password != "" {
		creds = fmt.Sprintf("%s:%s", username, password)
	}

	if len(query) > 0 {
		return fmt.Sprintf("%s@tcp(%s:%s)/%s?%s", creds, host, port, dbname, query.Encode()), nil
	}

	return fmt.Sprintf("%s@tcp(%s:%s)/%s", creds, host, port, dbname), nil
}
