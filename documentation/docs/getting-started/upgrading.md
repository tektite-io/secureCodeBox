---
# SPDX-FileCopyrightText: the secureCodeBox authors
#
# SPDX-License-Identifier: Apache-2.0


title: "Upgrading secureCodeBox"
sidebar_label: Upgrading
path: "docs/getting-started/upgrading"
sidebar_position: 3

---
## From 4.X to 5.X

### Removed / Replaced ScanTypes

* `zap-baseline-scan` and `zap-advanced` in favor of the `zap-automation-framework`. The `zap-automation-framework` ScanTpye includes all functionalities of the removed ScanTypes and can be customized easily. The default ScanType for the AutoDiscovery has been changed to the `zap-automation-framework` as well. For migrating to the `zap-automation-framework` please refer to [migration to zap-automation framework](/docs/scanners/zap-automation-framework#migration-to-zap-automation-framework) guide.
* `amass` has been replaced with `subfinder`. Amass is still an amzing tool, but with its focus on becoming more of a standalone platform / database for attack surfaces keeping it integrated and updated in the secureCodeBox was getting harder and harder. [subfinder](https://github.com/projectdiscovery/subfinder) is a very good replacement for subdomain discovery, thats also generally quicker and produces a similar result.
* `kubeaudit` was removed as the scanner itself [isn't maintaned anymore](https://github.com/Shopify/kubeaudit?tab=readme-ov-file#-deprecation-notice-). As a replacement you can use the `trivy` with it's `k8s` scanning mode, see [trivy ScanType k8s example](https://www.securecodebox.io/docs/scanners/trivy#k8s).
* `typo3scan` was removed as the scanner itself [isn't maintaned anymore](https://github.com/whoot/Typo3Scan?tab=readme-ov-file#unsupported). Most security aspects of typo3 are now hard to verify from the outside as it requires authentication (which is really good). Some typo3 security aspects (e.g. a incomplete installation) can be verified by [nuclei](https://www.securecodebox.io/docs/scanners/nuclei).
* `doggo` was removed. Doggo was added primarily as an experimentation to be used to deduplicate duplicate scan target from cascading rules based on DNS entries. That approach hasn't worked out unfortunately. The doggo integration has been non-functional for a while (see: https://github.com/secureCodeBox/secureCodeBox/issues/2853). As an alternative, nuclei already includes some DNS record based checks, if checks for specific records are required custom nuclei rules could be used to fulfil those requirements.
* `cmseek` was removed. cmseek has seen little updates in the last years. Our secureCodeBox integration with cmseek was always pretty basic, only supporting joomla (a specfifc CMS) results, which hasn't been a big focus for us. As a replacement we recommend using nuclei which has joomla rules which will likely receive more updates in the future.

➡️  [Reference: #2670](https://github.com/secureCodeBox/secureCodeBox/issues/2670)

### Renamed ClusterRole and ClusterRoleBinding

To avoid naming collisions with other cluster‑scoped resources, the operator's ClusterRole formerly called `manager-role` has been renamed to `securecodebox‑manager-role`, and the corresponding ClusterRoleBinding `manager-rolebinding` is now `securecodebox‑manager-rolebinding`. The official Helm chart will automatically create and reference these new names when you update the operator.

If you maintain a custom deployment that directly references `manager-role` or `manager-rolebinding`, be sure to update those references to `securecodebox‑manager-role` and `securecodebox‑manager-rolebinding` respectively.

➡️  [Reference: #3002](https://github.com/secureCodeBox/secureCodeBox/pull/3002)

### Changes to trivy k8s scope (namespace / cluster)

The `kubeauditScope` on the `trivy` ScanType chart was renamed to `k8sScanScope` Scope. The previous name was used for consistency with the `kubeaudit` ScanType, but it never really made sense and was confusing.
The default `k8sScanScope` scope was also changed from `cluster` to `namespace`, The cluster mode needs cluster wide permissions, which makes the trivy chart hard to install in properly locked down RBAC setups.

➡️  [Reference: #3025](https://github.com/secureCodeBox/secureCodeBox/pull/3025)

### Removed Integrated Elasticsearch and Kibana Helm Charts

The integrated Elasticsearch and Kibana Helm charts have been dropped from the Persistence ElasticSearch Hook. These charts were intended as a quick-start option, but since Elastic no longer provides their own Helm charts, they have been removed. The documentation has been updated with guidance on setting up an Elasticsearch cluster using the [ECK operator](https://www.elastic.co/elastic-cloud-kubernetes).

➡️  [Reference: #2892](https://github.com/secureCodeBox/secureCodeBox/issues/2892)

### Changed Default Elasticsearch Index

The default Elasticsearch index has been updated from `scbv2` to `scb`. The inclusion of `v2` was a confusing oversight that has been outdated since the release of secureCodeBox v3.
If you had previously ingested finding using the scbv2 index prefix you can keep using it by setting the `indexPrefix` helm value back to `scbv2` or by migrating your existing indexes to match the new naming scheme.

➡️  [Reference: #2892](https://github.com/secureCodeBox/secureCodeBox/issues/2892)

## From 3.X to 4.X

### Renamed the docker images of demo-targets to include a "demo-target-" prefix

The docker images for the custom demo targets built by secureCodeBox include a "demo-target-" prefix i.e:
  * securecodebox/old-typo3 --> securecodebox/**demo-target**-old-typo3
  * securecodebox/old-joomla --> securecodebox/**demo-target**-old-joomla
  * securecodebox/old-wordpress --> securecodebox/**demo-target**-old-wordpress

These images are usually used for testing and demo purposes. If you use these images, you would have to rename them appropriately.

➡️  [Reference: #1360](https://github.com/secureCodeBox/secureCodeBox/pull/1360)

### Changed name of Container AutoDiscovery scans

Previously scheduled scans generated by the container autodiscovery are named in the format `scan-image_name-at-image_hash`. The resulting scan pod will be called `scan-scan-image_name-at-image_hash`.
To avoid the duplicate “scan-scan”, the scheduled scans from the container autodiscovery are renamed. As a result, the container autodiscovery will no longer correctly “recognize” the old scans anymore. It will instead create new scans according to the new naming scheme. The old scheduled scans must be deleted manually.

➡️  [Reference: #1193](https://github.com/secureCodeBox/secureCodeBox/pull/1193)

### Cascading rules are disabled by default

Having the Cascading rules enabled by default on scanner helm install, has led to some confusion on the users side as mentioned in issue [#914](https://github.com/secureCodeBox/secureCodeBox/issues/914). As a result Cascading rules will have to be explicitly enabled by setting the `cascadingRules.enabled` value to `true`. For example as so:
```yaml
helm upgrade --install nmap oci://ghcr.io/securecodebox/helm/nmap --set=cascadingRules.enabled=true 
```

➡️  [Reference: #1347](https://github.com/secureCodeBox/secureCodeBox/pull/1347)

### Service Autodiscovery - Managed-by label assumed to be presented for all scans

Old versions of the operator did not set `app.kubernetes.io/managed-by` label. Starting with V4 the service autodiscovery will assume every scheduled scan created by the autodiscovery will have this label. This means that older scheduled scans without the label will not be detected by the service autodiscovery and new duplicate scheduled scans will be created. Old scheduled scans without the `app.kubernetes.io/managed-by` label must be deleted manually.  
➡️  [Reference: #1349](https://github.com/secureCodeBox/secureCodeBox/pull/1349)

### Rename host to hostname in zap findings

The `zap` and `zap-advanced` parsers where changed to increase the consistency between the different SCB scanners. The `zap` and `zap-advanced` findings will now report `host` as `hostname`.

➡️  [Reference: #1346](https://github.com/secureCodeBox/secureCodeBox/pull/1346)

### Container AutoDiscovery enabled by default and more consistent behavior compared to Service AutoDiscovery

The container autodiscovery will now be enabled by default. Additionally the container autodiscovery will now check if the configured scantype is installed in the namespace before it creates a scheduled scan (just like the service autodiscovery).

➡️  [Reference: #1112](https://github.com/secureCodeBox/secureCodeBox/pull/1112)

### Update Nuclei Finding Format

The format for [Nuclei](https://www.securecodebox.io/docs/scanners/nuclei) have been updated. With it some attributes have been moved / renamed to more closely match the nuclei format and to be able to support most / all information contained from the findings.

➡️  [Reference: #1350](https://github.com/secureCodeBox/secureCodeBox/pull/1350)

### Renamed Amass `attributes.name` to `attributes.hostname`

One of the optional attributes in the finding format of Amass has been renamed from Amass attributes.name to attributes.hostname. If you are parsing this attribute in your application. Renaming it on your end is required

➡️  [Reference: #1605](https://github.com/secureCodeBox/secureCodeBox/pull/1605)

### Removed AngularCSTI Scanner Integration

The integration for the AngularCSTI scanner has been removed and will no longer be officially supported.
Demand for the scanner was low and AngularJS (1.x) has been officially [deprecated in 2022](https://blog.angular.io/discontinued-long-term-support-for-angularjs-cc066b82e65a). Using AngularJS in 2023 is therefore a security liability, even when the specific site is not vulnerable to CSTI.

➡️  [Reference: #1649](https://github.com/secureCodeBox/secureCodeBox/pull/1649)

### Renamed the scanners ssh-scan and sslyze `hint` to `mitigation`

We added a new attribute to the finding format called `mitigation` which is used to store information about how to mitigate the finding/issue. The `hint` attribute of the findings of the scanners `ssh-scan` and `sslyze` has been renamed to `mitigation` to be more consistent with the other scanners.

➡️  [Reference: #1639](https://github.com/secureCodeBox/secureCodeBox/pull/1639)

### AutoDiscovery takes a list of scans in config file

The AutoDiscovery now takes a list of scans in its configuration. Each configured scan needs a unique name so that the AutoDiscovery is able to distinguish between scans. Because of this it is possible to configure multiple scans with the same scan type.
To upgrade wrap your existing scanConfig into a list and rename `scanConfig` to `scanConfigs`, see the default [values.yaml of the AutoDiscovery](https://github.com/secureCodeBox/secureCodeBox/blob/main/auto-discovery/kubernetes/values.yaml) as reference.

Example config:
```yaml
    scanConfigs:
      - scanType: trivy-image
        name: "my-trivy-scan"
        parameters:
          - "{{ .ImageID }}"
        repeatInterval: "168h"
      - scanType: trivy-image
        name: "second-trivy-scan"
        parameters:
          - "{{ .ImageID }}"
        repeatInterval: "168h"
```
➡️  [Reference: #1447](https://github.com/secureCodeBox/secureCodeBox/pull/1447)

### Findings Format: inconsistent ip address fields removed, replaced with standardized `ip_addresses`

In v3 and previous version there was no standardized format for ip address. Depending on the scanner the in attribute fields with different names and generally was only limited to a single address and to ipv4 only.
The `hostOrIP` helper for CascadingRule has been updated to use the new standardized field. If a finding has multiple ip addresses and no hostname the `hostOrIP` helper sorts the list of ips alphabetically and picks the first one.

With v4 the following changes were made to scanner formats to standardize the ip addresses information:

- amass: added `ip_addresses` field to attributes. The addresses still remains as is contains additional information like the CIDR & ASN
- doggo: added `ip_addresses` field to attributes. Only set for A & AAAA records
- ncrack: renamed `ip_address` field to `ip_addresses` and changed format to a list to support multiple addresses
- nikto: renamed `ip_address` field to `ip_addresses` and changed format to a list to support multiple addresses
- nuclei: renamed `ip` field to `ip_addresses` and changed format to a list to support multiple addresses
- ssh-audit: renamed `ip_address` field to `ip_addresses` and changed format to a list to support multiple addresses
- sslyze: renamed `ip_address` field to `ip_addresses` and changed format to a list to support multiple addresses
- test-scan: renamed `ip_address` field to `ip_addresses` and changed format to a list to support multiple addresses
- whatweb: renamed `ipAddress` field to `ip_addresses` and changed format to a list to support multiple addresses
- wpscan: renamed `ip_address` field to `ip_addresses` and changed format to a list to support multiple addresses

As a example the findings for nmap has been changed like the following:

```diff
{
  "name": "Host: juice-shop.demo-targets.svc.cluster.local",
  "category": "Host",
  "description": "Found a host",
  "location": "juice-shop.demo-targets.svc.cluster.local",
  "severity": "INFORMATIONAL",
  "osi_layer": "NETWORK",
  "attributes": {
-    "ip_address": "10.102.186.25",
+    "ip_addresses": ["10.102.186.25"],
    "hostname": "juice-shop.demo-targets.svc.cluster.local",
    "operating_system": null
  }
}
```

➡️  Reference: #1701, #1748

### `ssh-scan` (Mozilla ssh_scan) considered deprecated

SSH-Scan (Mozilla ssh_scan) is now considered deprecated as the tool is no longer maintained by mozilla. As a replacement we've added integration for [ssh-audit](https://github.com/jtesta/ssh-audit) as a replacement. The ssh-scan integration is still in this release but will be removed in a upcoming (feature) release. @Reet00 & @sofi0071

➡️  [Reference: #1713](https://github.com/secureCodeBox/secureCodeBox/pull/1713)

## From 2.X to 3.X

### Upgraded Kubebuilder Version to v3
The CRD's are now using `apiextensions.k8s.io/v1` instead of `apiextensions.k8s.io/v1beta1` which requries at least Kubernetes Version 1.16 or higher.
The Operator now uses the new kubebuilder v3 command line flag for enabling leader election and setting the metrics port. If you are using the official secureCodeBox Helm Charts for your deployment this has been updated automatically. 

If you are using a custom deployment you have to change the `--enable-leader-election` flag to `--leader-elect` and `--metrics-addr` to `--metrics-bind-address`. For more context see: https://book.kubebuilder.io/migration/v2vsv3.html#tldr-of-the-new-gov3-plugin

➡️  [Reference: #512](https://github.com/secureCodeBox/secureCodeBox/pull/512)

### Restructured the secureCodeBox HelmCharts to introduce more consistency in HelmChart Values
The secureCodeBox HelmCharts for hooks and scanners are following a new structure for all HelmChart Values:

Instead of secureCodebox Version 2 example:

```yaml
image:
  # image.repository -- Container Image to run the scan
  repository: owasp/zap2docker-stable
  # image.tag -- defaults to the charts appVersion
  tag: null

parserImage:
  # parserImage.repository -- Parser image repository
  repository: docker.io/securecodebox/parser-zap
  # parserImage.tag -- Parser image tag
  # @default -- defaults to the charts version
  tag: null

parseJob:
  # parseJob.ttlSecondsAfterFinished -- seconds after which the Kubernetes job for the parser will be deleted. Requires the Kubernetes TTLAfterFinished controller: https://kubernetes.io/docs/concepts/workloads/controllers/ttlafterfinished/
  ttlSecondsAfterFinished: null

scannerJob:
  # scannerJob.ttlSecondsAfterFinished -- seconds after which the Kubernetes job for the scanner will be deleted. Requires the Kubernetes TTLAfterFinished controller: https://kubernetes.io/docs/concepts/workloads/controllers/ttlafterfinished/
  ttlSecondsAfterFinished: null
  # scannerJob.backoffLimit -- There are situations where you want to fail a scan Job after some amount of retries due to a logical error in configuration etc. To do so, set backoffLimit to specify the number of retries before considering a scan Job as failed. (see: https://kubernetes.io/docs/concepts/workloads/controllers/job/#pod-backoff-failure-policy)
  # @default -- 3
  backoffLimit: 3
```

The new HelmChart Values structure in secureCodebox Version 3 looks like:

```yaml
parser:
  image:
    # parser.image.repository -- Parser image repository
    repository: docker.io/securecodebox/parser-zap
    # parser.image.tag -- Parser image tag
    # @default -- defaults to the charts version
    tag: null

  # parser.ttlSecondsAfterFinished -- seconds after which the Kubernetes job for the parser will be deleted. Requires the Kubernetes TTLAfterFinished controller: https://kubernetes.io/docs/concepts/workloads/controllers/ttlafterfinished/
  ttlSecondsAfterFinished: null
  # @default -- 3
  backoffLimit: 3

scanner:
  image:
    # scanner.image.repository -- Container Image to run the scan
    repository: owasp/zap2docker-stable
    # scanner.image.tag -- defaults to the charts appVersion
    tag: null

  # scanner.ttlSecondsAfterFinished -- seconds after which the Kubernetes job for the scanner will be deleted. Requires the Kubernetes TTLAfterFinished controller: https://kubernetes.io/docs/concepts/workloads/controllers/ttlafterfinished/
  ttlSecondsAfterFinished: null
  # scanner.backoffLimit -- There are situations where you want to fail a scan Job after some amount of retries due to a logical error in configuration etc. To do so, set backoffLimit to specify the number of retries before considering a scan Job as failed. (see: https://kubernetes.io/docs/concepts/workloads/controllers/job/#pod-backoff-failure-policy)
  # @default -- 3
  backoffLimit: 3
```
➡️  [Reference: #472](https://github.com/secureCodeBox/secureCodeBox/issues/472)
➡️  [Reference: #483](https://github.com/secureCodeBox/secureCodeBox/pull/483)
➡️  [Reference: #484](https://github.com/secureCodeBox/secureCodeBox/pull/484)

### Added scanner.appendName to chart values
Using `{{ .Release.name }}` in the `nmap` HelmChart Name for `scanTypes` causes issues when using this chart as a dependency of another chart. All scanners HelmCharts already used a fixed name for the `scanType` they introduce, with one exception: the `nmap` scanner HelmChart. 

The nmap exception was originally introduced to make it possible configure yourself an `nmap-privilidged` scanType, which is capable of running operating system scans which requires some higher privileges: https://www.securecodebox.io/docs/scanners/nmap#operating-system-scans

This idea for extending the name of a scanType is now in Version 3 general available for all HelmCharts.

The solution was to add a new HelmChart Value `scanner.appendName` for appending a suffix to the already defined scanType name. 
Example: the `scanner.nameAppend: -privileged` for the ZAP scanner will create `zap-baseline-scan-privileged`, `zap-api-scan-privileged`, `zap-full-scan-privileged` as new scanTypes instead of `zap-baseline-scan`, `zap-api-scan`, `zap-full-scan`.

➡️  [Reference: #469](https://github.com/secureCodeBox/secureCodeBox/pull/469)

### Renamed demo-apps to demo-targets
The provided vulnerable demos are renamed from `demo-apps` to `demo-targets`, this includes the namespace and the folder of the [helmcharts](https://github.com/secureCodeBox/secureCodeBox/tree/main/demo-targets).

### Renamed the hook declarative-subsequent-scans to cascading-scans
The hook responsible for cascading scans is renamed from `declarative-subsequent-scans` to `cascading-scans`.

➡️  [Reference: #481](https://github.com/secureCodeBox/secureCodeBox/pull/481)

### Fixed Name Consistency In Docker Images / Repositories
For the docker images for scanners and parsers we already had the naming convention of prefixing these images with `scanner-` or `parser-`.

Hook images however were named inconsistently (some prefixed with `hook-` some unprefixed).
To introduce more consistency we renamed all hook images and prefix them with `hook-` like we did with parser and scanner images.

Please beware of this if you are referencing some of our hook images in your own HelmCharts or custom implementations.

➡️  [Reference: #500](https://github.com/secureCodeBox/secureCodeBox/pull/500)

### Renamed `lurcher` to `lurker`

In the 3.0 release, we corrected the misspelling in `lurcher`. To remove the remains after upgrade, delete the old service accounts and roles from the namespaces where you have executed scans in the past:

```bash
# Find relevant namespaces
kubectl get serviceaccounts --all-namespaces | grep lurcher

# Delete role, role binding and service account for the specific namespace 
kubectl --namespace <NAMESPACE> delete serviceaccount lurcher
kubectl --namespace <NAMESPACE> delete rolebindings lurcher
kubectl --namespace <NAMESPACE> delete role lurcher
```

➡️  [Reference: #537](https://github.com/secureCodeBox/secureCodeBox/pull/537)

### Removed Hook Teams Webhook
We implemented a more general *[notification hook](https://www.securecodebox.io/docs/hooks/notification-hook)* which can be used to notify different systems like *[MS Teams](https://www.securecodebox.io/docs/hooks/notification-hook#configuration-of-a-ms-teams-notification)* and *[Slack](https://www.securecodebox.io/docs/hooks/notification-hook#configuration-of-a-slack-notification)* and also [Email](https://www.securecodebox.io/docs/hooks/notification-hook#configuration-of-an-email-notification) based in a more flexible way with [custom message templates](https://www.securecodebox.io/docs/hooks/notification-hook#custom-message-templates). With this new hook in place it is not nessesary to maintain the preexisting MS Teams Hook any longer and therefore we removed it.

➡️  [Reference: #570](https://github.com/secureCodeBox/secureCodeBox/pull/570)
