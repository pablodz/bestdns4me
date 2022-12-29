package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"text/tabwriter"
	"time"

	"github.com/schollz/progressbar/v3"
)

func main() {
	// Get a list of public DNS providers from a public site that lists them for free
	dnsProviders, err := getPublicDNSProviders()
	if err != nil {
		fmt.Printf("Error getting public DNS providers: %s\n", err)
		return
	}

	amountOfTestsPerDNS := 10
	// Set up the progress bar
	lenDNS := len(dnsProviders)

	// create and start new bar
	bar := progressbar.NewOptions(lenDNS*amountOfTestsPerDNS,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(false),
		// progressbar.OptionSetWidth(80),
		progressbar.OptionSetDescription("[cyan][👽][reset] Doing lookups..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	// Use a channel to receive the results from the goroutines
	resultCh := make(chan dnsResult)

	// Test each DNS provider in a separate goroutine
	for provider, ip := range dnsProviders {
		go func(ip string, provider string) {
			var total time.Duration
			for i := 0; i < amountOfTestsPerDNS; i++ {
				// Update the progress bar
				bar.Add(1)
				// generate lookup
				domains := []string{"google.com", "youtube.com", "hotmail.com", "whatsapp.com", "instagram.com"}
				domain := domains[rand.Intn(len(domains))]
				// Do the DNS lookup
				start := time.Now()
				_, err := net.LookupHost(domain)
				elapsed := time.Since(start)

				if err != nil {
					fmt.Printf("Error connecting to %s: %s\n", ip, err)
					continue
				}

				total += elapsed
			}
			average := total / time.Duration(amountOfTestsPerDNS)
			resultCh <- dnsResult{
				providerName: provider,
				providerIp:   ip,
				elapsed:      average,
			}
		}(ip, provider)
	}

	// Print the results in a tab-separated table
	w := &tabwriter.Writer{}
	// Wait for all goroutines to complete and store the results in the map
	for i := 0; i < len(dnsProviders); i++ {
		result := <-resultCh
		if i == 0 {
			// fmt.Println()
			// fmt.Println()
			// print settings in the test
			w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			// print column names
			fmt.Fprintln(w, "\nDNS Owner\tDNS Provider\tAverage Time(microseconds)(Less is better)")
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", result.providerName, result.providerIp, fmt.Sprint(result.elapsed.Microseconds()))
	}

	w.Flush()

	// Message to the user
	fmt.Println()
	fmt.Println("*We recommend using the DNS provider that has the fastest connection time.")

}

// dnsResult represents the result of a DNS lookup
type dnsResult struct {
	providerName string
	providerIp   string
	elapsed      time.Duration
}

// getPublicDNSProviders fetches a list of public DNS providers from a public site
// that lists them for free.
func getPublicDNSProviders() (map[string]string, error) {
	// Replace this placeholder function with a real implementation that fetches
	// the list of public DNS providers from a public site.
	return map[string]string{
		"Google":             "8.8.8.8",
		"Google.2":           "4.4.4.4",
		"Cloudflare":         "1.1.1.1",
		"Cloudflare.2":       "1.0.0.1",
		"Quad9":              "9.9.9.9",
		"Quad9.2":            "149.112.112.9",
		"OpenDNS":            "208.67.222.222",
		"OpenDNS.2":          "208.67.220.220",
		"Level 3":            "4.2.2.1",
		"Norton ConnectSafe": "199.85.126.10",
		"Hurricane Electric": "74.82.42.42",
		"CleanBrowsing":      "185.228.168.168",
		"CleanBrowsing.2":    "185.228.169.9",
		"Yandex":             "77.88.8.8",
		"AdGuard":            "176.103.130.130",
		"AdGuard.2":          "176.103.130.131",
		"Verisign":           "64.6.64.6",
		"Verisign.2":         "64.6.65.6",
	}, nil
}
