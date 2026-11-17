// Command asn-mmdb converts the iptoasn.com ip2asn-combined TSV into a
// GeoLite2-ASN compatible MaxMind DB.
package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/oschwald/maxminddb-golang/v2"
)

func main() {
	epoch := flag.Int64("epoch", 0, "build epoch (defaults to now)")
	flag.Parse()
	if flag.NArg() != 2 {
		log.Fatal("usage: asn-mmdb [-epoch N] ip2asn-combined.tsv.gz GeoLite2-ASN.mmdb")
	}
	if err := build(flag.Arg(0), flag.Arg(1), *epoch); err != nil {
		log.Fatal(err)
	}
	if err := verify(flag.Arg(1)); err != nil {
		log.Fatal(err)
	}
}

func build(input, output string, epoch int64) error {
	tree, err := mmdbwriter.New(mmdbwriter.Options{
		BuildEpoch:   epoch,
		DatabaseType: "GeoLite2-ASN",
		Description: map[string]string{
			"en": "ASN data from iptoasn.com (PDDL 1.0); not affiliated with MaxMind",
		},
		Languages:               []string{"en"},
		IPVersion:               6,
		RecordSize:              24,
		IncludeReservedNetworks: true,
	})
	if err != nil {
		return err
	}

	f, err := os.Open(input)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(gz)
	ranges := 0
	for line := 1; scanner.Scan(); line++ {
		fields := strings.SplitN(scanner.Text(), "\t", 5)
		if len(fields) != 5 {
			return fmt.Errorf("line %d: expected 5 fields", line)
		}
		asn, err := strconv.ParseUint(fields[2], 10, 32)
		if err != nil {
			return fmt.Errorf("line %d: %w", line, err)
		}
		if asn == 0 {
			continue
		}
		start, end := net.ParseIP(fields[0]), net.ParseIP(fields[1])
		if start == nil || end == nil {
			return fmt.Errorf("line %d: invalid range", line)
		}
		if v4 := start.To4(); v4 != nil {
			start, end = v4, end.To4()
		}
		record := mmdbtype.Map{
			"autonomous_system_number":       mmdbtype.Uint32(asn),
			"autonomous_system_organization": mmdbtype.String(fields[4]),
		}
		if err := tree.InsertRange(start, end, record); err != nil {
			return fmt.Errorf("line %d: %w", line, err)
		}
		ranges++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if ranges == 0 {
		return fmt.Errorf("no ranges in %s", input)
	}

	out, err := os.Create(output)
	if err != nil {
		return err
	}
	if _, err := tree.WriteTo(out); err != nil {
		out.Close()
		return err
	}
	log.Printf("wrote %d ranges to %s", ranges, output)
	return out.Close()
}

func verify(path string) error {
	db, err := maxminddb.Open(path)
	if err != nil {
		return err
	}
	defer db.Close()
	var record struct {
		Number uint32 `maxminddb:"autonomous_system_number"`
	}
	for addr, want := range map[string]uint32{"1.1.1.1": 13335, "2606:4700:4700::1111": 13335} {
		if err := db.Lookup(netip.MustParseAddr(addr)).Decode(&record); err != nil {
			return err
		}
		if record.Number != want {
			return fmt.Errorf("%s: got AS%d, want AS%d", addr, record.Number, want)
		}
	}
	return nil
}
