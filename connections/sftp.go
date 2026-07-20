package connections

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sync"
	"time"

	"github.com/radian-solusi/microservice-helpers/config"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	defaultSFTPPort               = 22
	defaultSFTPDialTimeout        = 10 * time.Second
	sshFxNoSuchFile        uint32 = 2
)

type sftpConnection struct {
	cfg       config.SftpConfig
	addr      string
	timeout   time.Duration
	sshClient *ssh.Client
	client    *sftp.Client
	mu        sync.Mutex
}

func NewSFTPClient(cfg config.SftpConfig) (SFTP, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("sftp host is required")
	}
	if cfg.User == "" {
		return nil, fmt.Errorf("sftp user is required")
	}
	if cfg.Password == "" {
		return nil, fmt.Errorf("sftp password is required")
	}

	port := cfg.Port
	if port == 0 {
		port = defaultSFTPPort
	}

	conn := &sftpConnection{
		cfg:     cfg,
		addr:    fmt.Sprintf("%s:%d", cfg.Host, port),
		timeout: defaultSFTPDialTimeout,
	}

	if err := conn.connect(); err != nil {
		return nil, err
	}

	log.Printf("SFTP client connected to %s", conn.addr)
	return conn, nil
}

func (s *sftpConnection) Client() *sftp.Client {
	return s.client
}

func (s *sftpConnection) UploadFile(remotePath string, data []byte, perm os.FileMode) error {
	if err := s.ensureConnection(); err != nil {
		return err
	}

	if perm == 0 {
		perm = 0o644
	}

	dir := path.Dir(remotePath)
	if dir != "." && dir != "/" {
		if err := s.client.MkdirAll(dir); err != nil {
			return fmt.Errorf("failed to ensure remote directory %s: %w", dir, err)
		}
	}

	file, err := s.client.OpenFile(remotePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("failed to open remote file %s: %w", remotePath, err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write remote file %s: %w", remotePath, err)
	}

	if err := s.client.Chmod(remotePath, perm); err != nil {
		return fmt.Errorf("failed to set permissions for %s: %w", remotePath, err)
	}

	return nil
}

func (s *sftpConnection) DownloadFile(remotePath string) ([]byte, error) {
	if err := s.ensureConnection(); err != nil {
		return nil, err
	}

	file, err := s.client.Open(remotePath)
	if err != nil {
		if isNotExist(err) {
			return nil, os.ErrNotExist
		}
		return nil, fmt.Errorf("failed to open remote file %s: %w", remotePath, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read remote file %s: %w", remotePath, err)
	}

	return data, nil
}

func (s *sftpConnection) ListDir(remoteDir string) ([]string, error) {
	if remoteDir == "" {
		remoteDir = "."
	}

	if err := s.ensureConnection(); err != nil {
		return nil, err
	}

	entries, err := s.client.ReadDir(remoteDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list remote directory %s: %w", remoteDir, err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}

	return names, nil
}

func (s *sftpConnection) DeleteFile(remotePath string) error {
	if err := s.ensureConnection(); err != nil {
		return err
	}

	if err := s.client.Remove(remotePath); err != nil {
		if isNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete remote file %s: %w", remotePath, err)
	}

	return nil
}

func (s *sftpConnection) FileExists(remotePath string) (bool, error) {
	if err := s.ensureConnection(); err != nil {
		return false, err
	}

	if _, err := s.client.Stat(remotePath); err != nil {
		if isNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to stat remote file %s: %w", remotePath, err)
	}

	return true, nil
}

func (s *sftpConnection) EnsureDir(remoteDir string) error {
	if remoteDir == "" || remoteDir == "." || remoteDir == "/" {
		return nil
	}

	if err := s.ensureConnection(); err != nil {
		return err
	}

	if err := s.client.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("failed to create remote directory %s: %w", remoteDir, err)
	}

	return nil
}

func (s *sftpConnection) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var closeErr error

	if s.client != nil {
		if err := s.client.Close(); err != nil {
			closeErr = errors.Join(closeErr, err)
		}
		s.client = nil
	}

	if s.sshClient != nil {
		if err := s.sshClient.Close(); err != nil {
			closeErr = errors.Join(closeErr, err)
		}
		s.sshClient = nil
	}

	return closeErr
}

func (s *sftpConnection) IsConnected() bool {
	return s.client != nil
}

func (s *sftpConnection) ensureConnection() error {
	if s.client != nil {
		return nil
	}
	return s.connect()
}

func (s *sftpConnection) connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil {
		return nil
	}

	sshConfig := &ssh.ClientConfig{
		User:            s.cfg.User,
		Auth:            []ssh.AuthMethod{ssh.Password(s.cfg.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         s.timeout,
	}

	sshClient, err := ssh.Dial("tcp", s.addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to establish SSH connection to %s: %w", s.addr, err)
	}

	client, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return fmt.Errorf("failed to initialize SFTP client: %w", err)
	}

	s.sshClient = sshClient
	s.client = client
	return nil
}

func isNotExist(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, os.ErrNotExist) {
		return true
	}

	var statusErr *sftp.StatusError
	if errors.As(err, &statusErr) {
		return statusErr.Code == sshFxNoSuchFile
	}

	return false
}
