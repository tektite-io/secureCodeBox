---
title: "DefectDojo"
category: "hook"
type: "persistenceProvider"
state: "in progress"
usecase: "Publishes all Scan Reports to DefectDojo."
---

## About

The DefectDojo persistenceProvider hook imports the reports from the scans automatically into DefectDojo.
Scanners which are directly supported by the secureCodeBox and DefectDojo the hook will use the DefectDojo "import scan"
method to upload the results to DefectDojo.

When a scanner is supported by the secureCodeBox but not DefectDojo, the hooks imports the findings using the
"DefectDojo finding" api, this might lead to missing information in the DefectDojo findings as the fields
from secureCodeBox cannot be directly mapped to the fields from DefectDojo. 

## Deployment

Installing the DefectDojo persistenceProvider hook will add a _ReadOnly Hook_ to your namespace.

```bash
helm upgrade --install dd secureCodeBox/persistence-defectdojo \
    --set="defectdojo.url=https://defectdojo-django.default.svc" \
    --set="defectdojo.auth.username=admin" \
    --set="defectdojo.auth.apiKey=08b7...." \
```

## Chart Configuration

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| defectdojo.auth.apiKey | string | `""` | API v2 Key (just the key) |
| defectdojo.auth.username | string | `"admin"` |  |
| defectdojo.url | string | `"http://defectdojo-django.default.svc"` |  |
| image.repository | string | `"docker.io/j12934/hook-defectdojo"` | Hook image repository |
| image.tag | string | `nil` |  |

[elastic.io]: https://www.elastic.co/products/elasticsearch
