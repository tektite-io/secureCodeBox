// SPDX-FileCopyrightText: the secureCodeBox authors
//
// SPDX-License-Identifier: Apache-2.0

export const Hooks = [
  {
    title: "Azure Monitor",
    type: "persistenceProvider",
    usecase: "Publishes all Scan Findings to Azure Monitor.",
    path: "docs/hooks/azure-monitor",
    imageUrl: "img/integrationIcons/Default.svg",
  },
  {
    title: "Cascading Scans",
    type: "processing",
    usecase: "Cascading Scans based declarative Rules.",
    path: "docs/hooks/cascading-scans",
    imageUrl: "img/integrationIcons/Default.svg",
  },
  {
    title: "DefectDojo",
    type: "persistenceProvider",
    usecase: "Publishes all Scan Reports to OWASP DefectDojo.",
    path: "docs/hooks/defectdojo",
    imageUrl: "img/integrationIcons/DefectDojo.svg",
  },
  {
    title: "Dependency-Track",
    type: "persistenceProvider",
    usecase: "Publishes all CycloneDX SBOMs to Dependency-Track.",
    path: "docs/hooks/dependency-track",
    imageUrl: "img/integrationIcons/Default.svg",
  },
  {
    title: "Elasticsearch",
    type: "persistenceProvider",
    usecase: "Publishes all Scan Findings to Elasticsearch.",
    path: "docs/hooks/elasticsearch",
    imageUrl: "img/integrationIcons/Elasticsearch.svg",
  },
  {
    title: "Finding Post Processing",
    type: "dataProcessing",
    usecase: "Updates fields for findings meeting specified conditions.",
    path: "docs/hooks/finding-post-processing",
    imageUrl: "img/integrationIcons/Default.svg",
  },
  {
    title: "Generic WebHook",
    type: "integration",
    usecase: "Publishes Scan Findings as WebHook.",
    path: "docs/hooks/generic-webhook",
    imageUrl: "img/integrationIcons/Default.svg",
  },
  {
    title: "Notification WebHook",
    type: "integration",
    usecase: "Publishes Scan Summary to MS Teams, Slack and others.",
    path: "docs/hooks/notification-webhook",
    imageUrl: "img/integrationIcons/Default.svg",
  },
  {
    title: "Static Report",
    type: "persistenceProvider",
    usecase: "Publishes all Scan Findings as HTML Report.",
    path: "docs/hooks/static-report",
    imageUrl: "img/integrationIcons/Default.svg",
  },
  {
    title: "Update Field",
    type: "dataProcessing",
    usecase: "Updates fields in finding results.",
    path: "docs/hooks/update-field",
    imageUrl: "img/integrationIcons/Default.svg",
  },
];

export const Scanners = [
  {
    title: "ffuf",
    type: "Webserver",
    usecase: "Webserver and WebApplication Elements and Content Discovery",
    path: "docs/scanners/ffuf",
    imageUrl: "img/integrationIcons/ffuf.svg",
  },
  {
    title: "Git Repo Scanner",
    type: "Repository",
    usecase: "Discover Git repositories",
    path: "docs/scanners/git-repo-scanner",
    imageUrl: "img/integrationIcons/Default.svg",
  },
  {
    title: "Gitleaks",
    type: "Repository",
    usecase: "Find potential secrets in repositories",
    path: "docs/scanners/gitleaks",
    imageUrl: "img/integrationIcons/Gitleaks.svg",
  },
  {
    title: "Kube Hunter",
    type: "Kubernetes",
    usecase: "Kubernetes Vulnerability Scanner",
    path: "docs/scanners/kube-hunter",
    imageUrl: "img/integrationIcons/Default.svg",
  },
  {
    title: "Ncrack",
    type: "Authentication",
    usecase: "Network authentication bruteforcing",
    path: "docs/scanners/ncrack",
    imageUrl: "img/integrationIcons/Ncrack.svg",
  },
  {
    title: "Nikto",
    type: "Webserver",
    usecase: "Webserver Vulnerability Scanner",
    path: "docs/scanners/nikto",
    imageUrl: "img/integrationIcons/Nikto.svg",
  },
  {
    title: "Nmap",
    type: "Network",
    usecase: "Network discovery and security auditing",
    path: "docs/scanners/nmap",
    imageUrl: "img/integrationIcons/Nmap.svg",
  },
  {
    title: "Nuclei",
    type: "Website",
    usecase: "Nuclei is a fast, template based vulnerability scanner.",
    path: "docs/scanners/nuclei",
    imageUrl: "img/integrationIcons/Nuclei.svg",
  },
  {
    title: "Screenshooter",
    type: "WebApplication",
    usecase: "Takes Screenshots of websites",
    path: "docs/scanners/screenshooter",
    imageUrl: "img/integrationIcons/Screenshooter.svg",
  },
  {
    title: "Semgrep",
    type: "Repository",
    usecase: "Static Code Analysis",
    path: "docs/scanners/semgrep",
    imageUrl: "img/integrationIcons/Semgrep.svg",
  },
  {
    title: "SSH-audit",
    type: "SSH",
    usecase: "SSH Configuration and Policy Scanner",
    path: "docs/scanners/ssh-audit",
    imageUrl: "img/integrationIcons/Default.svg",
  },
  {
    title: "SSLyze",
    type: "SSL",
    usecase: "SSL/TLS Configuration Scanner",
    path: "docs/scanners/sslyze",
    imageUrl: "img/integrationIcons/SSLyze.svg",
  },
  {
    title: "subfinder",
    type: "Network",
    usecase: "Subdomain Enumeration Scanner",
    path: "docs/scanners/subfinder",
    imageUrl: "img/integrationIcons/Default.svg",
  },
  {
    title: "Trivy SBOM",
    type: "Container",
    usecase: "Container Dependency Scanner",
    path: "docs/scanners/trivy-sbom",
    imageUrl: "img/integrationIcons/Default.svg",
  },
  {
    title: "Trivy",
    type: "Container",
    usecase: "Container Vulnerability Scanner",
    path: "docs/scanners/trivy",
    imageUrl: "img/integrationIcons/Trivy.svg",
  },
  {
    title: "Whatweb",
    type: "Network",
    usecase: "Website identification",
    path: "docs/scanners/whatweb",
    imageUrl: "img/integrationIcons/Whatweb.svg",
  },
  {
    title: "WPScan",
    type: "CMS",
    usecase: "Wordpress Vulnerability Scanner",
    path: "docs/scanners/wpscan",
    imageUrl: "img/integrationIcons/WPScan.svg",
  },
  {
    title: "ZAP Automation Framework",
    type: "WebApplication",
    usecase: "WebApp & OpenAPI Vulnerability Scanner",
    path: "docs/scanners/zap-automation-framework",
    imageUrl: "img/integrationIcons/Default.svg",
  },
];
export default { Hooks, Scanners };
