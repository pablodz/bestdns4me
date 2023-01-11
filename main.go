package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"text/tabwriter"
	"time"

	"github.com/schollz/progressbar/v3"
)

func main() {
	numOfTests := 5
	timeoutDnsResolve := 1 * time.Second
	// Get a list of public DNS providers from a public site that lists them for free
	dnsProviders, err := getPublicDNSProviders()
	if err != nil {
		log.Printf("Error getting public DNS providers: %s\n", err)
		return
	}
	domains, err := getDomains()
	if err != nil {
		log.Printf("Error getting Domains %s", err)
		return
	}

	// Set up the progress bar
	lenDNSProv := len(dnsProviders)
	lenDomains := len(domains)
	totalTests := numOfTests * lenDNSProv * lenDomains
	totalResolve := lenDNSProv

	// create and start new bar
	bar := progressbar.NewOptions(totalTests,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(false),
		// progressbar.OptionClearOnFinish(),
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
	for dnsName, dnsIp := range dnsProviders {
		bar.Add(1)
		go doMultipleDsnLookupHost(&dsnRequest{
			Timeout2Lookup: timeoutDnsResolve,
			domains:        domains,
			numberOfTtest:  numOfTests,
			providerIp:     dnsIp,
			providerName:   dnsName,
			resultCh:       resultCh,
		})
	}

	bar.Finish()

	// Print the results in a tab-separated table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	// print column names
	fmt.Fprintln(w, "\nDNS Owner\tDNS Provider\tAverage Time(ms)")

	// Wait for all goroutines to complete and store the results in the map
	for i := 0; i < totalResolve; i++ {
		result := <-resultCh
		if result.Error != nil {
			fmt.Fprintf(w, "%s\t%s\t%s\n", result.providerName, result.providerIp, "Timeout")
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\n", result.providerName, result.providerIp, fmt.Sprint(result.avgTime.Milliseconds()))
		}
		if i == totalResolve-1 {
			break
		}
	}

	w.Flush()

	// Message to the user
	fmt.Println("\n*We recommend using the DNS provider that has the fastest connection time.")
	fmt.Println("Less time is better.")
}

type dsnRequest struct {
	Timeout2Lookup time.Duration
	domains        []string
	numberOfTtest  int
	providerIp     string
	providerName   string
	resultCh       chan dnsResult
}

// dnsResult represents the result of a DNS lookup
type dnsResult struct {
	Error        error
	providerName string
	providerIp   string
	avgTime      time.Duration
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

func getDomains() ([]string, error) {
	return []string{
		// "amazon.com",
		// "apple.com",
		// "blogger.com",
		// "dropbox.com",
		// "ebay.com",
		// "facebook.com",
		"github.com",
		"google.com",
		// "hotmail.com",
		// "instagram.com",
		"linkedin.com",
		// "pinterest.com",
		// "reddit.com",
		// "salesforce.com",
		// "shopify.com",
		// "spotify.com",
		// "twitter.com",
		// "whatsapp.com",
		// "yahoo.com",
		"youtube.com",
	}, nil
}

func doMultipleDsnLookupHost(req *dsnRequest) {
	var totalTime time.Duration
	var avgTime time.Duration

	for domIdx := 0; domIdx < len(req.domains); domIdx++ {
		nErr := 0
		for nTest := 0; nTest < req.numberOfTtest; nTest++ {
			// Do the DNS lookup
			start := time.Now()
			_, err := doLookupHostWithTimeout(req.domains[domIdx], req.Timeout2Lookup)
			if err != nil {
				nErr++
				break
			}
			elapsed := time.Since(start)
			totalTime += elapsed
		}
		if nErr != 0 {
			// log.Println("Error", req.providerIp, req.providerName, req.Timeout2Lookup, req.domains[domIdx])
			req.resultCh <- dnsResult{
				Error:        errors.New("Time"),
				providerName: req.providerName,
				providerIp:   req.providerIp,
				avgTime:      req.Timeout2Lookup,
			}
			return
		}

	}

	avgTime = totalTime / time.Duration(req.numberOfTtest*len(req.domains))

	req.resultCh <- dnsResult{
		providerName: req.providerName,
		providerIp:   req.providerIp,
		avgTime:      avgTime,
	}
}

func doLookupHostWithTimeout(domain string, timeout time.Duration) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return net.DefaultResolver.LookupHost(ctx, domain)
}
