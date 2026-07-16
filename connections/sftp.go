package connections

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/pkg/sftp"
	helperconfig "github.com/radian-solusi/go-helpers/config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type sftpWrapper struct {
	client *sftp.Client
	conn   *ssh.Client
	mu     sync.Mutex
}

// NewSFTPClient connects to the configured SFTP server and returns an SFTP
// client. Host, user, and password are required. Port defaults to 22.
//
// Host-key policy: if cfg.KnownHostsFile is set, knownhosts.New is used.
// Setting cfg.HostKey to "insecure-ignore" allows ssh.InsecureIgnoreHostKey
// (dev/test only — NOT for production). Otherwise KnownHostsFile is required.
func NewSFTPClient(cfg helperconfig.SftpConfig) (SFTP, error) {
	if cfg.Host == "" {
		return nil, errors.New("sftp: host is required")
	}
	if cfg.User == "" {
		return nil, errors.New("sftp: user is required")
	}
	if cfg.Password == "" {
		return nil, errors.New("sftp: password is required")
	}
	port := cfg.Port
	if port == 0 {
		port = 22
	}

	var hostKeyCallback ssh.HostKeyCallback
	switch {
	case cfg.KnownHostsFile != "":
		cb, err := knownhosts.New(cfg.KnownHostsFile)
		if err != nil {
			return nil, fmt.Errorf("sftp: parse known_hosts: %w", err)
		}
		hostKeyCallback = cb
	case cfg.HostKey == "insecure-ignore":
		// ponytail: dev/test only; upgrade to knownhosts or pinned key for production
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	default:
		return nil, errors.New("sftp: known_hosts_file or host_key='insecure-ignore' is required")
	}

	sshCfg := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            []ssh.AuthMethod{ssh.Password(cfg.Password)},
		HostKeyCallback: hostKeyCallback,
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", port))
	sshClient, err := ssh.Dial("tcp", addr, sshCfg)
	if err != nil {
		return nil, fmt.Errorf("sftp: ssh dial %s: %w", addr, err)
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("sftp: open sftp session: %w", err)
	}
	return &sftpWrapper{client: sftpClient, conn: sshClient}, nil
}

func (s *sftpWrapper) Client() *sftp.Client { return s.client }

func (s *sftpWrapper) IsConnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return false
	}
	_, err := s.client.Getwd()
	return err == nil
}

func (s *sftpWrapper) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var errs []error
	if s.client != nil {
		errs = append(errs, s.client.Close())
	}
	if s.conn != nil {
		errs = append(errs, s.conn.Close())
	}
	return errors.Join(errs...)
}

func (s *sftpWrapper) UploadFile(path string, data []byte, perm os.FileMode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := s.client.Create(path)
	if err != nil {
		return fmt.Errorf("sftp create %q: %w", path, err)
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("sftp write %q: %w", path, err)
	}
	return s.client.Chmod(path, perm)
}

func (s *sftpWrapper) DownloadFile(path string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := s.client.Open(path)
	if err != nil {
		return nil, fmt.Errorf("sftp open %q: %w", path, err)
	}
	defer f.Close()
	return io.ReadAll(f)
}

func (s *sftpWrapper) DeleteFile(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.client.Remove(path)
}

func (s *sftpWrapper) FileExists(path string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.client.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *sftpWrapper) EnsureDir(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.client.MkdirAll(path)
}
