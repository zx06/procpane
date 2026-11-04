package ssh

import (
	"net"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func signersFromAgent(sock string) ([]gossh.Signer, error) {
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	ag := agent.NewClient(conn)
	return ag.Signers()
}
