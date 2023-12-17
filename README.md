# uqw - urlquery-worker (v0.2)

A worker for submitting URLs and retrive reports from http://urlquery.net

It monitors files for input (URLs) which are submitted to urlquery.net for analysis.

Report data (screenshot, JSON) can be retrived and saved to a output directory for further handling / analysis.
Supports retrival of report data via both webhooks and REST API.

Get a APIKEY by creating a account at:
https://urlquery.net/user/signup


## Usage example

Simple config
```yaml
---
apikey: <apikey> 

webhooks: 
  enabled: false
  listen: 0.0.0.0:8080

  # Setup webhook listeners for report events
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
      access: public # public, restricted, private
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