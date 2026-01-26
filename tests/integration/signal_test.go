package integration_test

import (
	"os"
	"os/exec"
	"syscall"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Signal Handling", func() {
	var session *gexec.Session
	var localAICommand *exec.Cmd

	BeforeEach(func() {
		// Prepare the command but don't start it yet
		// We assume 'local-ai' binary is built and in the current directory or PATH
		// Based on Makefile, binary name is local-ai
		localAICommand = exec.Command(os.Getenv("LOCALAI_BINARY_PATH"))
		if os.Getenv("LOCALAI_BINARY_PATH") == "" {
			localAICommand = exec.Command("./local-ai")
		}
	})

	AfterEach(func() {
		if session != nil {
			session.Kill()
		}
	})

	It("should exit on custom configured signal (SIGUSR1)", func() {
		localAICommand.Env = append(os.Environ(), "LOCALAI_EXIT_SIGNALS=SIGUSR1")
		// We need to set a dummy backend or disable things so it starts up quickly and stays running?
		// Or just run it. It usually starts a server.
		// We might need to give it a port to avoid conflicts if running in parallel, but integration tests usually handle this.
		// Let's rely on defaults or add --address if needed.
		
		var err error
		session, err = gexec.Start(localAICommand, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		// Wait for it to start (look for some log line or just wait a bit)
		// For now, let's wait a couple of seconds
		time.Sleep(2 * time.Second)

		// Send SIGUSR1
		if session.Command.Process != nil {
			err = session.Command.Process.Signal(syscall.SIGUSR1)
			Expect(err).NotTo(HaveOccurred())
		}

		// It should exit. main.go exits with os.Exit(1)
		Eventually(session, 5*time.Second).Should(gexec.Exit(1))
	})

	It("should exit on default signal (SIGINT) when no config provided", func() {
		// Ensure env var is not set
		localAICommand.Env = os.Environ() 
		// Remove LOCALAI_EXIT_SIGNALS if it's in Environ (it shouldn't be by default, but to be safe)
		// Simpler: just overwrite it with empty string? No, empty might mean "no signals"? 
		// In main.go: if len(cli.CLI.ExitSignals) > 0 { ... } else { default }
		// So unset it.
		
		var err error
		session, err = gexec.Start(localAICommand, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(2 * time.Second)

		if session.Command.Process != nil {
			err = session.Command.Process.Signal(syscall.SIGINT)
			Expect(err).NotTo(HaveOccurred())
		}

		Eventually(session, 5*time.Second).Should(gexec.Exit(1))
	})
})
