package storage

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.tyss.io/cj3636/dman/internal/config"
	"github.com/go-sql-driver/mysql"
)

type mariaRealBackend struct{ db *sql.DB }

func NewMariaBackend(cfgRoot string) (Backend, error) {
	return nil, errors.New("use NewMariaRealBackend for configured instance")
}

const mariaMaxRetries = 3

func (m *mariaRealBackend) retry(ctx context.Context, op string, fn func() error) error {
	var err error
	backoff := 50 * time.Millisecond
	for attempt := 0; attempt < mariaMaxRetries; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= 2
	}
	return err
}

// NewMariaRealBackend creates a MariaDB backend using configuration.
func NewMariaRealBackend(cfg *config.Config) (Backend, error) {
	m := cfg.Maria
	var tlsName string
	if m.TLS {
		conf := &tls.Config{InsecureSkipVerify: m.TLSInsecureSkip}
		if m.TLSServerName != "" {
			conf.ServerName = m.TLSServerName
		}
		if m.TLSCA != "" {
			pem, err := os.ReadFile(m.TLSCA)
			if err != nil {
				return nil, err
			}
			pool := x509.NewCertPool()
			if !pool.AppendCertsFromPEM(pem) {
				return nil, errors.New("failed to append maria ca")
			}
			conf.RootCAs = pool
		}
		if m.TLSCert != "" && m.TLSKey != "" {
			cert, err := tls.LoadX509KeyPair(m.TLSCert, m.TLSKey)
			if err != nil {
				return nil, err
			}
			conf.Certificates = []tls.Certificate{cert}
		}
		// Register custom name once (idempotent best-effort)
		tlsName = "dman_custom"
		_ = mysql.RegisterTLSConfig(tlsName, conf)
	}
	params := "parseTime=true&charset=utf8mb4&loc=UTC"
	if tlsName != "" {
		params += "&tls=" + tlsName
	} else if m.TLS {
		params += "&tls=true"
	}
	cred := m.User
	if m.Password != "" {
		cred += ":" + m.Password
	}
	addrPart := m.Addr
	if m.Socket != "" {
		addrPart = "unix(" + m.Socket + ")"
	} else {
		addrPart = "tcp(" + addrPart + ")"
	}
	dsn := cred + "@" + addrPart + "/" + m.DB + "?" + params
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(25)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	b := &mariaRealBackend{db: db}
	if err := b.ensureSchema(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return b, nil
}

func (m *mariaRealBackend) ensureSchema(ctx context.Context) error {
	_, err := m.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS dman_files (
		user VARCHAR(128) NOT NULL,
		rel TEXT NOT NULL,
		data LONGBLOB NOT NULL,
		PRIMARY KEY(user(64), rel(255))
	)`)
	return err
}

func (m *mariaRealBackend) sanitize(user, rel string) (string, string, error) {
	if user == "" {
		return "", "", errors.New("empty user")
	}
	rel = filepath.ToSlash(strings.TrimPrefix(rel, "./"))
	if rel == "" {
		return "", "", errors.New("empty path")
	}
	if len(rel) > MaxPathLen {
		return "", "", errors.New("path too long")
	}
	if strings.HasPrefix(rel, "/") {
		return "", "", errors.New("absolute path disallowed")
	}
	if strings.Contains(rel, "..") {
		return "", "", errors.New("path traversal disallowed")
	}
	return user, rel, nil
}

func (m *mariaRealBackend) Save(user, rel string, r io.Reader) error {
	u, p, err := m.sanitize(user, rel)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	return m.retry(ctx, "save", func() error {
		_, e := m.db.ExecContext(ctx, `REPLACE INTO dman_files (user, rel, data) VALUES (?,?,?)`, u, p, b)
		return e
	})
}

func (m *mariaRealBackend) Open(user, rel string) (*os.File, error) {
	u, p, err := m.sanitize(user, rel)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	var data []byte
	err = m.retry(ctx, "open", func() error {
		row := m.db.QueryRowContext(ctx, `SELECT data FROM dman_files WHERE user=? AND rel=?`, u, p)
		return row.Scan(&data)
	})
	if err != nil {
		return nil, err
	}
	f, err := os.CreateTemp("", ".dman-maria-*.")
	if err != nil {
		return nil, err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return nil, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

func (m *mariaRealBackend) List() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	var out []string
	err := m.retry(ctx, "list", func() error {
		rows, e := m.db.QueryContext(ctx, `SELECT user, rel FROM dman_files`)
		if e != nil {
			return e
		}
		defer rows.Close()
		for rows.Next() {
			var u, r string
			if er := rows.Scan(&u, &r); er != nil {
				return er
			}
			out = append(out, u+"/"+r)
		}
		return rows.Err()
	})
	return out, err
}

func (m *mariaRealBackend) Delete(user, rel string) error {
	u, p, err := m.sanitize(user, rel)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.retry(ctx, "delete", func() error {
		_, e := m.db.ExecContext(ctx, `DELETE FROM dman_files WHERE user=? AND rel=?`, u, p)
		return e
	})
}
