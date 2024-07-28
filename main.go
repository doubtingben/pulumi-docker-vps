package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"golang.org/x/crypto/ssh"
)

const (
	ubuntuImage = "ubuntu-20-04-x64"
	region      = "nyc3"
	size        = "s-1vcpu-1gb"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Read cloud-init file content from userData.txt
		userData, err := os.ReadFile("userData.txt")
		if err != nil {
			return fmt.Errorf("failed to read userData.txt: %v", err)
		}

		// Handle SSH key
		sshKey, err := handleSSHKey(ctx)
		if err != nil {
			return fmt.Errorf("failed to handle SSH key: %v", err)
		}

		// Create a new DigitalOcean Droplet
		droplet, err := createDroplet(ctx, string(userData), sshKey)
		if err != nil {
			return fmt.Errorf("failed to create droplet: %v", err)
		}

		// Export the Droplet's IP address
		ctx.Export("dropletIp", droplet.Ipv4Address)
		return nil
	})
}

func handleSSHKey(ctx *pulumi.Context) (*digitalocean.SshKey, error) {
	stackName := ctx.Stack()
	keyFileName := fmt.Sprintf("%s_id_rsa", stackName)
	privateKeyPath := filepath.Join(".", keyFileName)
	publicKeyPath := privateKeyPath + ".pub"

	publicKey, err := getOrCreateSSHKey(privateKeyPath, publicKeyPath)
	if err != nil {
		return nil, err
	}

	return digitalocean.NewSshKey(ctx, "my-ssh-key", &digitalocean.SshKeyArgs{
		Name:      pulumi.String(fmt.Sprintf("%s-ssh-key", stackName)),
		PublicKey: pulumi.String(publicKey),
	})
}

func getOrCreateSSHKey(privateKeyPath, publicKeyPath string) (string, error) {
	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		if _, err := generateSSHKeyPair(privateKeyPath, publicKeyPath); err != nil {
			return "", fmt.Errorf("failed to generate SSH key pair: %v", err)
		}
	}

	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read public key file: %v", err)
	}
	return string(publicKeyBytes), nil
}

func createDroplet(ctx *pulumi.Context, userData string, sshKey *digitalocean.SshKey) (*digitalocean.Droplet, error) {
	return digitalocean.NewDroplet(ctx, "my-droplet", &digitalocean.DropletArgs{
		Image:    pulumi.String(ubuntuImage),
		Region:   pulumi.String(region),
		Size:     pulumi.String(size),
		UserData: pulumi.String(userData),
		SshKeys:  pulumi.StringArray{sshKey.Fingerprint},
	})
}

func generateSSHKeyPair(privateKeyPath, publicKeyPath string) (string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", err
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", err
	}

	publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)

	if err := os.WriteFile(privateKeyPath, privateKeyPEM, 0600); err != nil {
		return "", err
	}

	if err := os.WriteFile(publicKeyPath, publicKeyBytes, 0644); err != nil {
		return "", err
	}

	return string(publicKeyBytes), nil
}
