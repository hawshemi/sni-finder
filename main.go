package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	defaultAddress     = "0.0.0.0"
	defaultPort        = "443"
	defaultThreadCount = 128
	defaultTimeout     = 4
	outPutDef          = true
	outPutFileName     = "results.txt"
	domainsFileName    = "domains.txt"
	showFailDef        = false
	numIPsToCheck      = 10000
	workerPoolSize     = 100
)

var (
	log    = logrus.New()
	zeroIP = net.ParseIP("0.0.0.0")
	maxIP  = net.ParseIP("255.255.255.255")
	TlsVersions = map[uint16]string{
		0x0301: "1.0",
		0x0302: "1.1",
		0x0303: "1.2",
		0x0304: "1.3",
	}
)

type CustomTextFormatter struct {
	logrus.TextFormatter
}

type Scanner struct {
	addr           string
	port           string
	showFail       bool
	output         bool
	timeout        time.Duration
	wg             *sync.WaitGroup
	numberOfThread int
	mu             sync.Mutex
	ip             net.IP
	logFile        *os.File
	domainFile     *os.File
	dialer         *net.Dialer
	logChan        chan string
}

func (f *CustomTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	msg := entry.Message

	// Create the log entry without the "level=info" and with a new line
	return []byte(timestamp + msg + "\n\n"), nil
}

func (s *Scanner) Print(outStr string) {
	// Split the output string into IP address and the rest
	parts := strings.Split(outStr, " ")
	if len(parts) < 2 {
		return
	}
	
	ipAddress := parts[0]                // Extract the IP address part
	rest := strings.Join(parts[1:], " ") // Extract the rest of the message

	// Calculate the maximum IP address length
	maxIPLength := len("255.255.255.255")

	// Format the IP address with a fixed width
	formattedIP := fmt.Sprintf("%-*s", maxIPLength-8, ipAddress)

	// Create the final log entry with IP alignment
	logEntry := formattedIP + rest

	// Extract the domain from the log entry
	domain := extractDomain(logEntry)

	// Save the domain to domains.txt if it's not empty
	if domain != "" {
		saveDomain(domain, s.domainFile)
	}

	s.logChan <- logEntry
}

func extractDomain(logEntry string) string {
	// Split the log entry into words
	parts := strings.Fields(logEntry)

	// Search for a word that looks like a domain (contains a dot)
	for i, part := range parts {
		if strings.Contains(part, ".") && !strings.HasPrefix(part, "v") && i > 0 {
			// Split the part using ":" and take the first part (domain)
			domainParts := strings.Split(part, ":")
			return domainParts[0]
		}
	}

	return ""
}

func saveDomain(domain string, file *os.File) {
	if domain == "" {
		return
	}
	
	_, err := file.WriteString(domain + "\n")
	if err != nil {
		log.WithError(err).Error("Error writing domain into file")
	}
}

func main() {
	addrPtr := flag.String("addr", defaultAddress, "Destination to start scan")
	portPtr := flag.String("port", defaultPort, "Port to scan")
	threadPtr := flag.Int("thread", defaultThreadCount, "Number of threads to scan in parallel")
	outPutFile := flag.Bool("o", outPutDef, "Is output to results.txt")
	timeOutPtr := flag.Int("timeOut", defaultTimeout, "Time out of a scan")
	showFailPtr := flag.Bool("showFail", showFailDef, "Is Show fail logs")

	flag.Parse()
	
	// Initialize Logrus settings
	log.SetFormatter(&CustomTextFormatter{})
	log.SetLevel(logrus.InfoLevel) // Set the desired log level

	// Setup scanner with configuration
	s := setupScanner(*addrPtr, *portPtr, *showFailPtr, *outPutFile, *timeOutPtr, *threadPtr)
	if s == nil {
		return
	}
	defer s.cleanup()

	go s.logWriter()

	// Create a buffered channel for IPs to scan
	ipChan := make(chan net.IP, numIPsToCheck)

	// Start the worker pool
	for i := 0; i < s.numberOfThread; i++ {
		go s.worker(ipChan)
	}

	// Generate the IPs to scan and send them to the channel
	for i := 0; i < numIPsToCheck; i++ {
		nextIP := s.nextIP(true)
		if nextIP != nil {
			s.wg.Add(1)
			ipChan <- nextIP
		}
	}

	close(ipChan)

	// Wait for all scans to complete
	s.wg.Wait()
	close(s.logChan)
	log.Info("Scan completed.")
}

func setupScanner(addr, port string, showFail, output bool, timeoutSec int, threads int) *Scanner {
	s := &Scanner{
		addr:           addr,
		port:           port,
		showFail:       showFail,
		output:         output,
		timeout:        time.Duration(timeoutSec) * time.Second,
		wg:             &sync.WaitGroup{},
		numberOfThread: threads,
		ip:             net.ParseIP(addr),
		dialer:         &net.Dialer{},
		logChan:        make(chan string, numIPsToCheck),
	}

	// Open results.txt file for writing
	var err error
	s.logFile, err = os.OpenFile(outPutFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.WithError(err).Error("Failed to open log file")
		return nil
	}

	// Open domains.txt file for writing
	s.domainFile, err = os.OpenFile(domainsFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.WithError(err).Error("Failed to open domains.txt file")
		s.logFile.Close()
		return nil
	}

	return s
}

func (s *Scanner) cleanup() {
	if s.logFile != nil {
		s.logFile.Close()
	}
	if s.domainFile != nil {
		s.domainFile.Close()
	}
}

func (s *Scanner) logWriter() {
	for str := range s.logChan {
		log.Info(str) // Log with Info level
		if s.output {
			_, err := s.logFile.WriteString(str + "\n")
			if err != nil {
				log.WithError(err).Error("Error writing into file")
			}
		}
	}
}

func (s *Scanner) worker(ipChan <-chan net.IP) {
	for ip := range ipChan {
		s.Scan(ip)
		s.wg.Done()
	}
}

func (s *Scanner) nextIP(increment bool) net.IP {
	s.mu.Lock()
	defer s.mu.Unlock()

	ipb := big.NewInt(0).SetBytes(s.ip.To4())
	if increment {
		ipb.Add(ipb, big.NewInt(1))
	} else {
		ipb.Sub(ipb, big.NewInt(1))
	}

	b := ipb.Bytes()
	b = append(make([]byte, 4-len(b)), b...)
	nextIP := net.IP(b)

	if nextIP.Equal(zeroIP) || nextIP.Equal(maxIP) {
		return nil
	}

	s.ip = nextIP
	return s.ip
}

func (s *Scanner) Scan(ip net.IP) {
	str := ip.String()

	if ip.To4() == nil {
		str = "[" + str + "]"
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	conn, err := s.dialer.DialContext(ctx, "tcp", str+":"+s.port)
	if err != nil {
		if s.showFail {
			s.Print(fmt.Sprintf("Dial failed: %v", err))
		}
		return
	}
	defer conn.Close() // Ensure the connection is closed

	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	remoteIP := remoteAddr.IP.String()
	port := remoteAddr.Port
	line := fmt.Sprintf("%s:%d", remoteIP, port) + "\t"
	
	// Set deadline for TLS handshake
	conn.SetDeadline(time.Now().Add(s.timeout))
	
	c := tls.Client(conn, &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h2", "http/1.1"},
	})
	err = c.Handshake()

	if err != nil {
		if s.showFail {
			s.Print(fmt.Sprintf("%s - TLS handshake failed: %v", line, err))
		}
		return
	}
	defer c.Close() // Ensure the TLS client is also properly closed

	state := c.ConnectionState()
	alpn := state.NegotiatedProtocol

	if alpn == "" {
		alpn = "  "
	}

	if s.showFail || (state.Version == 0x0304 && alpn == "h2") {
		certSubject := ""
		if len(state.PeerCertificates) > 0 {
			certSubject = state.PeerCertificates[0].Subject.CommonName
		}

		// Filter out invalid certificates
		if isInvalidCertificate(certSubject) {
			return
		}

		s.Print(fmt.Sprint(" ", line, "---- TLS v", TlsVersions[state.Version], "    ALPN: ", alpn, " ----    ", certSubject, ":", s.port))
	}
}

func isInvalidCertificate(certSubject string) bool {
	numPeriods := strings.Count(certSubject, ".")
	return strings.HasPrefix(certSubject, "*") || 
	       certSubject == "localhost" || 
	       numPeriods != 1 || 
	       certSubject == "invalid2.invalid" || 
	       certSubject == "OPNsense.localdomain"
}
