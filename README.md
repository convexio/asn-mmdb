# asn-mmdb

Daily IP-to-ASN database in MaxMind DB format, built from
[iptoasn.com](https://iptoasn.com/) data.

The database uses the GeoLite2-ASN record layout and database type, so it works
as a drop-in ASN database for Logstash, Elasticsearch, `geoip2` libraries and
`mmdblookup`. It is not affiliated with MaxMind.

```json
{"autonomous_system_number": 15169, "autonomous_system_organization": "GOOGLE - Google LLC"}
```

## Releases

Every day a release tagged `YYYY.MM.DD` is published. Tags are never modified
after creation; the last 30 releases are kept.

| Asset               | Content                                                |
| ------------------- | ------------------------------------------------------ |
| `GeoLite2-ASN.mmdb` | Database                                               |
| `GeoLite2-ASN.tgz`  | Database archive for the Elastic GeoIP downloaders     |
| `database`          | Download index for the archive (relative URL)          |
| `SHA256SUMS`        | Checksums                                              |

Latest database:

```sh
curl -LO https://github.com/convexio/asn-mmdb/releases/latest/download/GeoLite2-ASN.mmdb
```

Behind caching proxies or repository mirrors, request a dated tag instead of
`latest`, e.g. `.../releases/download/2026.11.17/GeoLite2-ASN.mmdb`.

## Logstash and Elasticsearch

The download index at `https://convexio.github.io/asn-mmdb/v1/database` is
compatible with the managed GeoIP downloaders:

```yaml
# logstash.yml
xpack.geoip.downloader.endpoint: https://convexio.github.io/asn-mmdb/v1/database
```

```yaml
# elasticsearch.yml
ingest.geoip.downloader.endpoint: https://convexio.github.io/asn-mmdb/v1/database
```

Then use the `geoip` filter or processor with `default_database_type => "ASN"`
or `database_file: GeoLite2-ASN.mmdb`. To mirror the endpoint, copy
`GeoLite2-ASN.tgz` and `database` of a release into one directory and point
the downloader at the `database` file.

## License

Code: MIT. Data: [iptoasn.com](https://iptoasn.com/), Public Domain
([PDDL 1.0](https://opendatacommons.org/licenses/pddl/1-0/)).
