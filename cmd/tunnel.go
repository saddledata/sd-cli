package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/saddledata/sd-cli/internal/api"
	"github.com/saddledata/sd-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	localHost  string
	localPort  int
	gatewayUrl string
)

type WsWrapper struct {
	conn *websocket.Conn
	mu   sync.Mutex
	buf  []byte
}

func NewWsWrapper(conn *websocket.Conn) *WsWrapper {
	return &WsWrapper{conn: conn}
}

func (w *WsWrapper) Read(p []byte) (int, error) {
	if len(w.buf) > 0 {
		n := copy(p, w.buf)
		w.buf = w.buf[n:]
		return n, nil
	}

	for {
		msgType, msg, err := w.conn.ReadMessage()
		if err != nil {
			return 0, err
		}

		if msgType == websocket.BinaryMessage {
			n := copy(p, msg)
			if n < len(msg) {
				w.buf = msg[n:]
			}
			return n, nil
		}
	}
}

func (w *WsWrapper) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	err := w.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *WsWrapper) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.Close()
}

var tunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: "Manage and serve outbound database tunnels",
}

var tunnelServeCmd = &cobra.Command{
	Use:   "serve [tunnel-id]",
	Short: "Start a local proxy for a specific tunnel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tunnelID := args[0]
		
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		_, ctx, err := config.GetActiveContext(cfg, context)
		if err != nil {
			return err
		}

		client := api.NewClient(ctx)

		fmt.Printf("Fetching credentials for tunnel %s...\n", tunnelID)
		
		// 1. Fetch credentials
		resp, err := client.Get(fmt.Sprintf("/tunnels/%s/credentials", tunnelID))
		if err != nil {
			return fmt.Errorf("failed to fetch tunnel credentials: %w", err)
		}

		var creds map[string]interface{}
		if err := json.Unmarshal(resp, &creds); err != nil {
			return fmt.Errorf("failed to parse credentials: %w", err)
		}

		// Determine local target
		targetHost := localHost
		if targetHost == "" {
			if h, ok := creds["host"].(string); ok {
				targetHost = h
			} else {
				targetHost = "127.0.0.1"
			}
		}

		targetPort := localPort
		if targetPort == 0 {
			if p, ok := creds["port"].(float64); ok {
				targetPort = int(p)
			} else {
				targetPort = 5432 // Default postgres
			}
		}

		targetAddr := fmt.Sprintf("%s:%d", targetHost, targetPort)

		assignedPort := 0
		if ap, ok := creds["_gateway_assigned_port"].(float64); ok {
			assignedPort = int(ap)
		}

		// 3. Connect to Gateway URL setup
		gwURL := gatewayUrl
		if gwURL == "" {
			gwURL = "ws://localhost:8082"
		}

		parsedGwURL, err := url.Parse(gwURL)
		if err != nil {
			return fmt.Errorf("invalid gateway URL: %w", err)
		}

		wsURL := fmt.Sprintf("%s/v1/tunnel/connect?tunnel_id=%s", parsedGwURL.String(), tunnelID)
		headers := http.Header{}
		headers.Add("X-Api-Key", ctx.ApiKey)
		dialer := *websocket.DefaultDialer

		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-stopChan
			fmt.Println("\nShutting down tunnel...")
			os.Exit(0)
		}()

		// Loop to support sequential client connections
		for {
			wsConn, _, err := dialer.Dial(wsURL, headers)
			if err != nil {
				log.Printf("Failed to connect to gateway: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if assignedPort > 0 {
				fmt.Printf("\n🚀 Tunnel is ACTIVE! Waiting for client connection...\nConnect securely via: tcp://<user>:<password>@%s:%d\n\n", parsedGwURL.Hostname(), assignedPort)
			} else {
				fmt.Println("\n🚀 Tunnel is ACTIVE! Waiting for client connection...")
			}

			// Wait for the gateway to signal that a client has connected via TCP
			msgType, msg, err := wsConn.ReadMessage()
			if err != nil {
				wsConn.Close()
				fmt.Println("Gateway disconnected. Re-establishing tunnel...")
				time.Sleep(1 * time.Second)
				continue
			}

			if msgType != websocket.TextMessage || string(msg) != "CONNECTED" {
				wsConn.Close()
				fmt.Println("Received invalid signal from gateway. Re-establishing tunnel...")
				time.Sleep(1 * time.Second)
				continue
			}

			fmt.Println("Client connected! Dialing local database...")

			tcpConn, err := net.Dial("tcp", targetAddr)
			if err != nil {
				log.Printf("Failed to connect to local database %s: %v", targetAddr, err)
				wsConn.Close()
				time.Sleep(5 * time.Second)
				continue
			}
			
			fmt.Println("✓ Local database connected. Proxying traffic...")

			wsWrapper := NewWsWrapper(wsConn)
			errChan := make(chan error, 2)

			go func() {
				_, err := io.Copy(wsWrapper, tcpConn)
				errChan <- err
			}()

			go func() {
				_, err := io.Copy(tcpConn, wsWrapper)
				errChan <- err
			}()

			// Block until one of the proxy streams finishes
			err = <-errChan
			if err != nil && err != io.EOF {
				// Don't clutter logs with expected EOF errors
			}
			
			tcpConn.Close()
			wsConn.Close()
			fmt.Println("Client disconnected. Re-establishing tunnel...")
			time.Sleep(1 * time.Second) // small delay before reconnect
		}
	},
}

func init() {
	tunnelServeCmd.Flags().StringVar(&localHost, "host", "", "Override the local database host")
	tunnelServeCmd.Flags().IntVar(&localPort, "port", 0, "Override the local database port")
	tunnelServeCmd.Flags().StringVar(&gatewayUrl, "gateway-url", "ws://localhost:8082", "The URL of the Saddle Tunnel Gateway")

	tunnelCmd.AddCommand(tunnelServeCmd)
	RootCmd.AddCommand(tunnelCmd)
}
