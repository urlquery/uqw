# uqw - urlquery-worker (v0.2)

A worker to submit and retrive reports from urlquery.net

Monitor files for input (URLs), and are submitted to urlquery.net for analysis.

Report data (screenshot, JSON) can be saved to a output directory further handling / analysis.
Supports retrival of report data via both webhooks and REST API.

## Usage example

Simple config
```yaml
---
apikey: <apikey> 

webhooks: 
  enabled: false
  listen: 0.0.0.0:8080

  reports:
    # Alerts from users YARA signatures
    alerted:
      enabled: false
      path: output/alerted/
      report: true        # retrive report data
      screenshot: true    # retrive screenshot
      domain_graph: false # retrive domain graph

    submitted:
      enabled: true
      path: output/submitted/
      report: true
      screenshot: true
      domain_graph: false


submit:
  - file: input/urls.public
    enabled: true

    settings:
      access: public
      tags: []

    # alternative to webhooks
    output:
      enabled: true
      path: output/submitted/
      report: true
      screenshot: true
      domain_graph: false
```

Start the worker
```bash
./uqw -config config.yaml
```

Submit a URL
```bash
echo "urlquery.net" >> input/urls.public
```