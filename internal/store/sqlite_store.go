package store

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"leblanc.io/open-go-ssl-checker/internal/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(driver string, dsn string) (*Store, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()

		return nil, fmt.Errorf("database ping error: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) InitSchema() error {
	_, err := s.db.Exec(`
        CREATE TABLE IF NOT EXISTS projects (
            id TEXT PRIMARY KEY,
            name TEXT UNIQUE,
            host TEXT,
            port TEXT,
            type TEXT,
			allow_insecure BOOLEAN DEFAULT FALSE
        )
    `)
	if err != nil {
		return fmt.Errorf("error creating projects table: %w", err)
	}

	_, err = s.db.Exec(`
        CREATE TABLE IF NOT EXISTS certificate_checks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            check_time DATETIME,
            project_id TEXT,
            domains TEXT,
			ip TEXT,
			issuer TEXT,
            expiry_date TEXT,
            days_remaining INTEGER,
            FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE
        )
    `)
	if err != nil {
		return fmt.Errorf("error creating certificate_checks table: %w", err)
	}

	return nil
}

func (s *Store) AddProject(project types.Project) error {
	_, err := s.db.Exec(
		"INSERT INTO projects (id, name, host, port, type, allow_insecure) VALUES (?, ?, ?, ?, ?, ?)",
		project.ID,
		project.Name,
		project.Host,
		project.Port,
		project.Type,
		project.AllowInsecure,
	)
	if err != nil {
		return fmt.Errorf("error inserting project: %w", err)
	}

	return nil
}

func (s *Store) GetProject(id string) (*types.Project, error) {
	row := s.db.QueryRow(
		"SELECT id, name, host, port, type, allow_insecure FROM projects WHERE id = ?",
		id,
	)

	var p types.Project

	if err := row.Scan(&p.ID, &p.Name, &p.Host, &p.Port, &p.Type, &p.AllowInsecure); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Project not found
		}

		return nil, fmt.Errorf("error retrieving project %s: %w", id, err)
	}

	return &p, nil
}

func (s *Store) ListProjects() ([]types.Project, error) {
	rows, err := s.db.Query(
		"SELECT id, name, host, port, type, allow_insecure FROM projects ORDER BY name ASC",
	)
	if err != nil {
		return nil, fmt.Errorf("error retrieving projects list: %w", err)
	}

	defer rows.Close()

	var projects []types.Project

	for rows.Next() {
		var p types.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Host, &p.Port, &p.Type, &p.AllowInsecure); err != nil {
			return nil, fmt.Errorf("error scanning project: %w", err)
		}

		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over projects: %w", err)
	}

	return projects, nil
}

func (s *Store) DeleteProject(projectID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	_, err = tx.Exec("DELETE FROM projects WHERE id = ?", projectID)
	if err != nil {
		tx.Rollback()

		return fmt.Errorf("error deleting project %s: %w", projectID, err)
	}

	return tx.Commit()
}

func (s *Store) AddCertificateCheck(check types.CertificateCheck) error {
	_, err := s.db.Exec(`
        INSERT INTO certificate_checks (
            check_time, project_id, domains, ip, issuer, expiry_date, days_remaining
        ) VALUES (?, ?, ?, ?, ?, ?, ?)
    `, check.CheckTime, check.ProjectID, check.Domains, check.IP, check.Issuer, check.ExpiryDate, check.DaysRemaining)
	if err != nil {
		return fmt.Errorf("error inserting certificate check: %w", err)
	}

	return nil
}

func (s *Store) GetCertificateChecksForProject(projectID string) ([]types.CertificateCheck, error) {
	query := `
        SELECT cc.id, cc.check_time, cc.project_id, p.name, cc.domains, cc.ip, cc.issuer, cc.expiry_date, cc.days_remaining
        FROM certificate_checks cc
        JOIN projects p ON cc.project_id = p.id
        WHERE cc.project_id = ?
        ORDER BY cc.check_time DESC
    `
	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving check history for project %s: %w",
			projectID,
			err,
		)
	}

	defer rows.Close()

	var checks []types.CertificateCheck

	for rows.Next() {
		var c types.CertificateCheck
		err := rows.Scan(
			&c.ID,
			&c.CheckTime,
			&c.ProjectID,
			&c.ProjectName,
			&c.Domains,
			&c.IP,
			&c.Issuer,
			&c.ExpiryDate,
			&c.DaysRemaining,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning a certificate check: %w", err)
		}

		checks = append(checks, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"error iterating over certificate checks: %w",
			err,
		)
	}

	return checks, nil
}

func (s *Store) GetLatestChecksSummary() ([]types.ProjectCheckSummary, error) {
	query := `
        SELECT
            p.id,
            p.name,
            p.host,
            p.port,
            p.type,
			p.allow_insecure,
            cc.check_time,
            cc.domains,
			cc.ip,
			cc.issuer,
            cc.expiry_date,
            cc.days_remaining
        FROM projects p
        LEFT JOIN (
            SELECT
                project_id,
                check_time,
                domains,
				ip,
			    issuer,
                expiry_date,
                days_remaining,
                ROW_NUMBER() OVER(PARTITION BY project_id ORDER BY check_time DESC) as rn
            FROM certificate_checks
        ) cc ON p.id = cc.project_id AND cc.rn = 1
        ORDER BY p.name ASC;
    `
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving latest checks: %w",
			err,
		)
	}

	defer rows.Close()

	var summaries []types.ProjectCheckSummary

	for rows.Next() {
		var s types.ProjectCheckSummary

		var checkTime sql.NullTime
		var domains sql.NullString
		var ip sql.NullString
		var issuer sql.NullString
		var expiryDate sql.NullString
		var daysRemaining sql.NullInt64

		if err := rows.Scan(
			&s.ProjectID, &s.ProjectName, &s.Host, &s.Port, &s.Type, &s.AllowInsecure,
			&checkTime, &domains, &ip, &issuer, &expiryDate, &daysRemaining,
		); err != nil {
			return nil, fmt.Errorf("error scanning check summary: %w", err)
		}

		if checkTime.Valid {
			s.CheckTime = &checkTime.Time
		}

		if domains.Valid {
			s.Domains = domains.String
		}

		if ip.Valid {
			s.IP = ip.String
		}

		if issuer.Valid {
			s.Issuer = issuer.String
		}

		if expiryDate.Valid {
			s.ExpiryDate = expiryDate.String
		}

		if daysRemaining.Valid {
			days := int(daysRemaining.Int64)
			s.DaysRemaining = &days
		}

		summaries = append(summaries, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over projects: %w", err)
	}

	return summaries, nil
}

func (s *Store) GetProjectName(projectID string) (string, error) {
	var name string
	err := s.db.QueryRow("SELECT name FROM projects WHERE id = ?", projectID).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("project %s not found", projectID)
		}

		return "", fmt.Errorf(
			"error retrieving project name %s: %w",
			projectID,
			err,
		)
	}

	return name, nil
}
