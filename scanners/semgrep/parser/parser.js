// SPDX-FileCopyrightText: the secureCodeBox authors
//
// SPDX-License-Identifier: Apache-2.0

const severityMap = new Map([
  ["info", "INFORMATIONAL"],
  ["warning", "MEDIUM"],
  ["error", "HIGH"],
]);

export async function parse(fileContent) {
  const report = JSON.parse(fileContent);

  if (!report || !report.results) {
    return [];
  }

  return report.results.flatMap((result) => {
    // Assemble location as path to file and line range
    const location =
      result.path + ":" + result.start.line + "-" + result.end.line;

    // Name of the finding is the rule ID from semgrep
    const name = result.check_id;

    // Description is either the message from result.extra.message, or a placeholder message
    const description =
      result.extra.message ||
      "(No description provided in semgrep rule - when using a custom rule, set the 'message' key)";

    // Category of the finding - use either result.extra.metadata.category, or a placeholder
    const category = result.extra.metadata.category || "semgrep-result";

    // severity of the issue: translate semgrep severity levels (INFO, WARNING, ERROR) to those of SCB (INFORMATIONAL, LOW, MEDIUM, HIGH)
    const severity = severityMap.has(result.extra.severity.toLowerCase())
      ? severityMap.get(result.extra.severity.toLowerCase())
      : "INFORMATIONAL";

    const cwe = result.extra.metadata?.cwe;
    const cweReference = cwe ? String(cwe).substring(4, 6) : null;

    const references = [
      // Map metadata references to an array of URL reference objects
      ...(result.extra.metadata?.references?.map((link) => ({
        type: "URL",
        value: link,
      })) || []),
      // If a CWE reference exists, add CWE and URL reference objects for it
      ...(cweReference
        ? [
            {
              type: "CWE",
              value: `CWE-${cweReference}`,
            },
            {
              type: "URL",
              value: `https://cwe.mitre.org/data/definitions/${cweReference}.html`,
            },
          ]
        : []),
    ];

    const attributes = {
      // Common weakness enumeration, if available
      cwe: result.extra.metadata.cwe || null,
      // OWASP category, if available
      owasp_category: result.extra.metadata.owasp || null,
      // References given in the rule
      references: references.length > 0 ? references : null,
      // Link to the semgrep rule
      rule_source: result.extra.metadata.source || null,
      // Which line of code matched?
      // TODO: Do we actually want to record this? There are also secret-detector rules for semgrep,
      // so maybe you don't actually want the plaintext match to be recorded unencrypted in some S3 bucket?
      // "matching_lines": result.extra.lines,
    };

    return {
      name,
      location,
      description,
      category,
      severity,
      attributes,
    };
  });
}
