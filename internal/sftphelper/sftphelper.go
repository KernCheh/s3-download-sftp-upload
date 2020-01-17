package sftphelper

import (
	"context"
	"fmt"
	"io"
	"net"
	"path"

	"github.com/pkg/sftp"
	"github.com/sephora-sea/s3-download-sftp-upload/internal/config"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	sshClient *ssh.Client
}

// GetClient established a SFTP connection with the credentials provided in:
// `SFTP_HOST`, `SFTP_PORT`, `SFTP_USERNAME`, `SFTP_PASSWORD`
func GetClient() (*Client, error) {
	sshConfig := &ssh.ClientConfig{
		User: config.GetInstance().SftpUserName,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.GetInstance().SftpPassword),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil // We do not provide hostkey check for now
		},
	}

	// Required because remote server uses not-so-secure key exchange
	sshConfig.Config.KeyExchanges = append(sshConfig.Config.KeyExchanges, "diffie-hellman-group-exchange-sha1", "diffie-hellman-group-exchange-sha256")

	connURI := fmt.Sprintf("%s:%s", config.GetInstance().SftpHost, config.GetInstance().SftpPort)

	fmt.Println("[SFTP Helper] Connecting to SFTP server at", connURI)
	c, err := ssh.Dial("tcp", connURI, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("[SFTP Helper] Error connecting: %s", err.Error())
	}

	fmt.Println("[SFTP Helper] Connected to SFTP server")
	return &Client{sshClient: c}, nil
}

// Upload will put a file in the sftp server which the `Client` is connected to. Use this method concurrently with the read buffer for maximum throughput
// reader - read buffer of contents to upload
// directory - directory to upload to from the sftp home, e.g '/internal/'
// filename - destination filename of file to upload
func (c *Client) Upload(reader io.Reader, directory, filename string) error {
	sftp, err := sftp.NewClient(c.sshClient)
	if err != nil {
		return fmt.Errorf("[SFTP Helper] Error creating client: %s", err.Error())
	}

	defer sftp.Close()

	fmt.Println("[SFTP Helper] Commencing upload for", filename)
	uploadFileName := path.Join(directory, filename)

	// leave your mark
	f, err := sftp.Create(uploadFileName)
	if err != nil {
		return fmt.Errorf("[SFTP Helper] Error creating file %s: %s", uploadFileName, err.Error())
	}

	bcount, err := f.ReadFrom(reader)
	if err != nil {
		return fmt.Errorf("[SFTP Helper] Error uploading file contents of %s: %s", uploadFileName, err.Error())
	}

	fmt.Printf("[SFTP Helper] %v bytes uploaded for %v\n", bcount, uploadFileName)

	// check it's there
	fi, err := sftp.Lstat(uploadFileName)
	if err != nil {
		return fmt.Errorf("[SFTP Helper] Error with file integrity check with %s: %s", uploadFileName, err.Error())
	}

	fmt.Printf("[SFTP Helper] File upload successful: %+v\n", fi)

	return nil
}

func (c *Client) UploadWithContext(ctx context.Context, reader io.Reader, directory, filename string) error {
	var err error
	okChan := make(chan bool)
	go func() {
		err = c.Upload(reader, directory, filename)
		okChan <- true
	}()

	select {
	case <-okChan:

	case <-ctx.Done():
		return ctx.Err()
	}
	return err
}
